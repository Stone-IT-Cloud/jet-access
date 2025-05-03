package main

import (
	"fmt"
	"os"

	ssh "github.com/Stone-IT-Cloud/jet-access/pkg/sshclient"
)

func main() {
	fmt.Println("Hello, World!")
	ssh_config := ssh.SSHConfig{
		Address: "45.55.41.188:22",
		User: "root",
	}
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}
	
	// Read the SSH key file
	ssh_key, err := os.ReadFile(homeDir + "/.ssh/id_rsa")
	if err != nil {
		fmt.Printf("Error reading SSH key file: %v\n", err)
		return
	}
	ssh_config.Key = ssh_key
	err = ssh.ConnectAndShell(ssh_config)
	if err != nil {
		fmt.Printf("Error connecting to SSH: %v\n", err)
	}
}