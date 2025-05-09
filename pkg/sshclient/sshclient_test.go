package sshclient

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gliderlabs/ssh" // Using gliderlabs/ssh for mock server
	gossh "golang.org/x/crypto/ssh"
)

// generateTestKey generates a new RSA private key for testing purposes.
// Returns private key PEM, public key authorized_keys format bytes, and error.
func generateTestKey(bits int, passphrase []byte) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	var pemBlock *pem.Block
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	var privatePEM []byte
	if len(passphrase) > 0 {
		// Para compatibilidad con versiones anteriores de Go, usamos EncryptPEMBlock
		// aunque esté obsoleto (seguro para pruebas)
		//nolint:staticcheck // Usamos función obsoleta intencionalmente por compatibilidad
		pemBlock, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", privBytes, passphrase, x509.PEMCipherAES256)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encrypt private key: %w", err)
		}
		privatePEM = pem.EncodeToMemory(pemBlock)
	} else {
		pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}
		privatePEM = pem.EncodeToMemory(pemBlock)
	}

	// Generate public key in authorized_keys format
	pubKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ssh public key: %w", err)
	}
	publicBytes := gossh.MarshalAuthorizedKey(pubKey)

	return privatePEM, publicBytes, nil
}

// startMockSSHServer starts a mock SSH server on a random port for testing.
// It returns the server address (host:port) and a function to stop the server.
func startMockSSHServer(t *testing.T, handler ssh.Handler, options ...ssh.Option) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Listen on a random available port
	if err != nil {
		t.Fatalf("Failed to listen on mock SSH port: %v", err)
	}
	addr := listener.Addr().String()

	server := ssh.Server{
		Handler: handler,
		// Add other options like host key, auth handlers etc.
	}

	// Add a host key to avoid client rejection
	hostKeySigner, err := generateSigner(2048)
	if err != nil {
		listener.Close()
		t.Fatalf("Failed to generate host key: %v", err)
	}
	server.AddHostKey(hostKeySigner)

	// Apply provided options (like auth methods)
	for _, opt := range options {
		if err := server.SetOption(opt); err != nil {
			listener.Close()
			t.Fatalf("Failed to set server option: %v", err)
		}
	}

	go func() {
		err := server.Serve(listener)
		// Don't log ErrServerClosed as it's expected on shutdown
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			// Log other errors, but don't fail the test here as it runs async
			log.Printf("Mock SSH server error: %v", err)
		}
	}()

	stop := func() {
		// Allow time for server to potentially start before closing
		time.Sleep(50 * time.Millisecond)
		server.Close()
		// Allow time for server goroutine to exit cleanly
		time.Sleep(50 * time.Millisecond)
	}

	return addr, stop
}

// generateSigner creates a gossh.Signer from a new RSA key.
func generateSigner(bits int) (gossh.Signer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return gossh.NewSignerFromKey(privateKey)
}

// --- Test Cases for ConnectAndShell ---

// TestConnectAndShell_AuthErrors focuses on errors *before* dialing.
func TestConnectAndShell_AuthErrors(t *testing.T) {
	keyPassphrase := "testpass"
	privateKeyEnc, _, err := generateTestKey(2048, []byte(keyPassphrase))
	if err != nil {
		t.Fatalf("Failed to generate encrypted test key: %v", err)
	}

	tests := []struct {
		name          string
		cfg           SSHConfig
		expectError   bool
		errorContains string // Substring expected in the error message
	}{
		{
			name:          "No Auth Methods Provided",
			cfg:           SSHConfig{Address: "localhost:2222", User: "test"},
			expectError:   true,
			errorContains: "no authentication methods successfully configured",
		},
		{
			name:          "Invalid Key (Malformed)",
			cfg:           SSHConfig{Address: "localhost:2222", User: "test", Key: []byte("-----BEGIN INVALID KEY-----\ninvalid\n-----END INVALID KEY-----")},
			expectError:   true,
			errorContains: "failed to parse private key",
		},
		{
			name:          "Encrypted Key (Missing Passphrase)",
			cfg:           SSHConfig{Address: "localhost:2222", User: "test", Key: privateKeyEnc},
			expectError:   true,
			errorContains: "failed to parse private key", // ParsePrivateKey fails first
		},
		{
			name:          "Encrypted Key (Incorrect Passphrase)",
			cfg:           SSHConfig{Address: "localhost:2222", User: "test", Key: privateKeyEnc, Passphrase: "wrongpass"},
			expectError:   true,
			errorContains: "failed to parse private key with passphrase", // ParseRawPrivateKeyWithPassphrase fails
		},
	}

	// Temporarily redirect log output during tests
	originalLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLogOutput)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We expect errors before dialing for these specific cases.
			err := ConnectAndShell(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("ConnectAndShell() expected an error, but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ConnectAndShell() expected error containing '%s', but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("ConnectAndShell() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestConnectAndShell_Integration uses a mock SSH server for more realistic testing.
func TestConnectAndShell_Integration(t *testing.T) {
	// 1. Generate test keys
	privateKey, publicKeyBytes, err := generateTestKey(2048, nil)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	password := "testpassword"

	// 2. Setup Mock Server Handlers
	// Simple echo handler for testing basic shell interaction
	mockShellHandler := func(s ssh.Session) {
		// Simulate PTY allocation check (optional, depends on client behavior)
		ptyReq, _, isPty := s.Pty()
		if isPty {
			fmt.Fprintf(s, "PTY Request: %s, %dx%d\n", ptyReq.Term, ptyReq.Window.Height, ptyReq.Window.Width)
		} else {
			fmt.Fprintln(s, "No PTY requested.")
		}

		fmt.Fprintf(s, "Hello %s\n", s.User())
		// Echo back input until "exit" or EOF
		buf := make([]byte, 1024)
		for {
			n, err := s.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(s.Stderr(), "Read error: %v\n", err)
				}
				break // Exit loop on EOF or error
			}
			line := strings.TrimSpace(string(buf[:n]))
			fmt.Fprintf(s, "You typed: %s\n", line)
			if line == "exit" {
				break
			}
		}
		fmt.Fprintln(s, "Exiting mock shell.")
		// No explicit exit status needed, closing the stream is sufficient
	}

	// Configure server for public key auth
	publicKeyAuth, _, _, _, err := gossh.ParseAuthorizedKey(publicKeyBytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}
	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		// Allow connection if user is 'keyuser' and key matches
		return ctx.User() == "keyuser" && ssh.KeysEqual(key, publicKeyAuth)
	})

	// Configure server for password auth
	passwordOption := ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
		// Allow connection if user is 'passuser' and password matches
		return ctx.User() == "passuser" && pass == password
	})

	// 3. Start Mock Server
	addr, stopServer := startMockSSHServer(t, mockShellHandler, publicKeyOption, passwordOption)
	defer stopServer()

	// --- Test Cases ---
	tests := []struct {
		name           string
		cfg            SSHConfig
		input          string   // Commands to send to the shell via Stdin
		expectOutput   []string // Substrings expected in Stdout
		expectError    bool
		errorContains  string
		nonInteractive bool // Flag to simulate non-TTY stdin
	}{
		{
			name: "Successful Connection with Key",
			cfg: SSHConfig{
				Address: addr,
				User:    "keyuser",
				Key:     privateKey,
			},
			input:        "test command\nexit\n",
			expectOutput: []string{"Hello keyuser", "You typed: test command", "Exiting mock shell."},
			expectError:  false,
		},
		{
			name: "Successful Connection with Password",
			cfg: SSHConfig{
				Address:  addr,
				User:     "passuser",
				Password: password,
			},
			input:        "another command\nexit\n",
			expectOutput: []string{"Hello passuser", "You typed: another command", "Exiting mock shell."},
			expectError:  false,
		},
		{
			name: "Failed Connection (Wrong User for Key)",
			cfg: SSHConfig{
				Address: addr,
				User:    "wronguser", // User doesn't match key auth rule
				Key:     privateKey,
			},
			input:         "",
			expectError:   true,
			errorContains: "ssh: handshake failed: ssh: unable to authenticate", // Or similar auth error
		},
		{
			name: "Failed Connection (Wrong Password)",
			cfg: SSHConfig{
				Address:  addr,
				User:     "passuser",
				Password: "wrongpassword",
			},
			input:         "",
			expectError:   true,
			errorContains: "ssh: handshake failed: ssh: unable to authenticate",
		},
		{
			name: "Non-Interactive Session (No PTY)",
			cfg: SSHConfig{
				Address: addr,
				User:    "keyuser",
				Key:     privateKey,
			},
			input:          "command1\nexit\n",
			expectOutput:   []string{"No PTY requested.", "Hello keyuser", "You typed: command1"}, // Server should indicate no PTY
			expectError:    false,
			nonInteractive: true,
		},
	}

	originalLogOutput := log.Writer()
	log.SetOutput(io.Discard) // Suppress client logs during test runs
	defer log.SetOutput(originalLogOutput)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 4. Redirect Stdin/Stdout/Stderr for the test
			oldStdin := os.Stdin
			oldStdout := os.Stdout
			oldStderr := os.Stderr

			stdinReader, stdinWriter, _ := os.Pipe()
			stdoutReader, stdoutWriter, _ := os.Pipe()
			stderrReader, stderrWriter, _ := os.Pipe()

			// If nonInteractive, use the pipe directly. Otherwise, keep os.Stdin
			// as the original TTY (if available) or the pipe if not a TTY.
			// The ConnectAndShell function itself checks term.IsTerminal.
			// Forcing non-interactive mode is best done by ensuring term.IsTerminal(fd) returns false.
			// Replacing os.Stdin with a pipe achieves this reliably in tests.
			os.Stdin = stdinReader // Use pipe for controlled input

			os.Stdout = stdoutWriter
			os.Stderr = stderrWriter

			// Cleanup function
			cleanup := func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
				os.Stderr = oldStderr
				stdinReader.Close()
				stdinWriter.Close()
				stdoutReader.Close()
				stdoutWriter.Close()
				stderrReader.Close()
				stderrWriter.Close()
			}
			defer cleanup()

			// 5. Run ConnectAndShell in a goroutine
			errChan := make(chan error, 1)
			go func() {
				// Ensure ConnectAndShell uses the redirected Stdin fd
				// Note: ConnectAndShell reads os.Stdin directly, so replacement is key.
				errChan <- ConnectAndShell(tt.cfg)
			}()

			// 6. Interact (if input provided)
			if tt.input != "" {
				// Write input slightly delayed to allow connection setup
				time.Sleep(150 * time.Millisecond)
				_, err := stdinWriter.Write([]byte(tt.input))
				if err != nil {
					// Don't fail test here, let ConnectAndShell return error if pipe breaks
					log.Printf("Warning: Failed to write to stdin pipe: %v", err)
				}
				// Close writer to signal EOF to the remote shell (important!)
				stdinWriter.Close()
			} else {
				// If no input, close writer immediately after short delay
				time.Sleep(50 * time.Millisecond)
				stdinWriter.Close()
			}

			// 7. Read Output & Wait for Finish
			var stdoutOutput bytes.Buffer
			var stderrOutput bytes.Buffer
			readDone := make(chan struct{})

			go func() {
				_, _ = io.Copy(&stdoutOutput, stdoutReader)
				readDone <- struct{}{} // Signal stdout reading finished
			}()
			go func() {
				_, _ = io.Copy(&stderrOutput, stderrReader)
				readDone <- struct{}{} // Signal stderr reading finished
			}()

			var returnedErr error
			select {
			case returnedErr = <-errChan:
				// ConnectAndShell finished, close writers to unblock readers
				stdoutWriter.Close()
				stderrWriter.Close()
			case <-time.After(5 * time.Second): // Timeout
				// If timed out, try closing writers to force reader exit
				stdoutWriter.Close()
				stderrWriter.Close()
				t.Fatal("ConnectAndShell timed out")
			}

			// Wait for both readers to finish
			<-readDone
			<-readDone

			// 8. Assertions
			stdoutStr := stdoutOutput.String()
			stderrStr := stderrOutput.String()

			// Log output for debugging if test fails
			if t.Failed() {
				t.Logf("Stdout:\n%s", stdoutStr)
				t.Logf("Stderr:\n%s", stderrStr)
			}

			if tt.expectError {
				if returnedErr == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tt.errorContains != "" && !strings.Contains(returnedErr.Error(), tt.errorContains) && !strings.Contains(stderrStr, tt.errorContains) {
					// Check both returned error and stderr for expected message
					t.Errorf("Expected error containing '%s', but got error: %v, stderr: %s", tt.errorContains, returnedErr, stderrStr)
				}
			} else {
				if returnedErr != nil {
					t.Errorf("Unexpected error: %v\nStderr: %s", returnedErr, stderrStr)
				}
				for _, expected := range tt.expectOutput {
					if !strings.Contains(stdoutStr, expected) {
						t.Errorf("Expected stdout to contain '%s', but got:\n%s", expected, stdoutStr)
					}
				}
			}
		})
	}
}
