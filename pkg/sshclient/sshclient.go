// pkg/sshclient/sshclient.go

package sshclient

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term" // For interacting with the terminal (getting size, setting raw mode)
)

// SSHConfig holds the necessary parameters to establish an SSH connection.
type SSHConfig struct {
	Address    string // Target server address (e.g., "hostname:port" or "ip:port")
	User       string // SSH username
	Port       string
	Key        []byte // Optional: Content of the private SSH key (can be nil or empty)
	Passphrase string // Optional: Passphrase for the private key (can be empty)
	Password   string // Optional: Password for password authentication (can be empty, alternative to Key)
}

// ConnectAndShell establishes an SSH connection using the provided configuration
// and starts an interactive shell session, connecting local Stdin/Stdout/Stderr.
func ConnectAndShell(cfg SSHConfig) error {
	// --- 1. Prepare Authentication Methods ---
	authMethods := []ssh.AuthMethod{}

	// Add key-based authentication if key content is provided
	if len(cfg.Key) > 0 {
		signer, err := ssh.ParsePrivateKey(cfg.Key)
		if err != nil {
			// If parsing fails, try with passphrase if provided
			if cfg.Passphrase != "" {
				rawKey, err := ssh.ParseRawPrivateKeyWithPassphrase(cfg.Key, []byte(cfg.Passphrase))
				if err != nil {
					return fmt.Errorf("failed to parse private key with passphrase: %w", err)
				}
				signer, err = ssh.NewSignerFromKey(rawKey)
				if err != nil {
					return fmt.Errorf("failed to create signer from parsed key: %w", err)
				}

			} else {
				// If no passphrase was provided or passphrase also failed
				return fmt.Errorf("failed to parse private key: %w", err)
			}
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Add password authentication if password is provided (Key takes precedence if both exist)
	// A real-world scenario might prioritize key auth, but this adds password if key isn't used.
	// You might adjust this logic based on your Vault secret structure and priority.
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication methods successfully configured (no valid key or password provided)")
	}

	// --- 2. Configure the SSH Client ---
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: authMethods,
		// HostKeyCallback is CRITICAL for security.
		// InsecureIgnoreHostKey() is DANGEROUS and should ONLY be used for initial testing/MVP validation.
		// A proper implementation MUST verify the host key against a trusted source (e.g., known_hosts file).
		// For this MVP, we use InsecureIgnoreHostKey() BUT IT MUST BE REPLACED.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // !!! SECURITY RISK - DO NOT USE IN PRODUCTION !!!
		// Recommended (Future): HostKeyCallback: ssh.KnownHosts(getKnownHostsFile()), // Implement known_hosts handling
	}

	log.Printf("Attempting SSH connection to %s@%s...", cfg.User, cfg.Address)

	// --- 3. Establish the Connection ---
	client, err := ssh.Dial("tcp", cfg.Address, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH server %s: %w", cfg.Address, err)
	}
	defer client.Close() // Ensure client connection is closed when function exits

	log.Println("SSH connection established.")

	// --- 4. Create a Session ---
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close() // Ensure session is closed when function exits

	// --- 5. Set up Terminal (PTY) for Interactive Shell ---
	// Get the file descriptor for standard input
	fd := int(os.Stdin.Fd())
	// Check if the terminal is interactive
	if term.IsTerminal(fd) {
		// Put the terminal in raw mode to handle shell input correctly
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			log.Printf("Warning: Could not set terminal to raw mode: %v. Interactive shell features may be limited.", err)
			// Continue without raw mode, but warn the user
		} else {
			// Restore the terminal state when the function exits
			defer term.Restore(fd, oldState)
		}

		// Get the terminal size to inform the remote session
		width, height, err := term.GetSize(fd)
		if err != nil {
			log.Printf("Warning: Could not get terminal size: %v", err)
			width, height = 80, 24 // Use a default size if unable to get actual size
		}

		// Request a pseudo-terminal (PTY)
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,     // enable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed, affects Ctrl+C
			ssh.TTY_OP_OSPEED: 14400, // output speed
		}

		if err := session.RequestPty("xterm-256color", height, width, modes); err != nil { // Use a common term type
			return fmt.Errorf("failed to request PTY: %w", err)
		}

	} else {
		log.Println("Stdin is not a terminal. Running in non-interactive mode.")
		// No PTY requested for non-interactive input
	}

	// --- 6. Connect Standard I/O Streams ---
	// Connect local standard input to the remote session's standard input
	session.Stdin = os.Stdin
	// Connect remote session's standard output to local standard output
	session.Stdout = os.Stdout
	// Connect remote session's standard error to local standard error
	session.Stderr = os.Stderr

	// --- 7. Start the Remote Shell ---
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start remote shell: %w", err)
	}

	log.Println("Interactive shell started. Type 'exit' to disconnect.")

	// --- 8. Wait for the Session to End ---
	// This blocks until the remote shell session is closed (e.g., user types 'exit', connection drops)
	if err := session.Wait(); err != nil {
		// Check if the error is just a non-zero exit status from the remote command/shell
		if exitErr, ok := err.(*ssh.ExitError); ok {
			log.Printf("Session ended with non-zero exit status: %d", exitErr.ExitStatus())
			// Depending on requirements, you might return this error or nil
			return nil // Often, a non-zero shell exit isn't treated as a tool failure
		}
		return fmt.Errorf("SSH session ended with unexpected error: %w", err)
	}

	log.Println("SSH session disconnected gracefully.")

	return nil // Indicate successful connection and session handling
}

// TODO (Future): Implement a proper HostKeyCallback using a known_hosts file
/*
func getKnownHostsFile() string {
	// This function would determine the path to the known_hosts file based on OS
	// and user's home directory.
	// For example:
	// usr, err := user.Current()
	// if err != nil {
	//     log.Printf("Warning: Could not determine user home directory: %v", err)
	//     return "" // Indicate no known_hosts file could be found
	// }
	// return filepath.Join(usr.HomeDir, ".ssh", "known_hosts")
	return "" // Placeholder
}
*/

// TODO (Future): Implement SCP functionality in this package (e.g., UploadFile, DownloadFile functions)
