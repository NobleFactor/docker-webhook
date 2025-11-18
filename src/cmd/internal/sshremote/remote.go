package sshremote

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// getExitReason returns a reason phrase for the given exit code
func getExitReason(exitCode int) string {
	switch exitCode {
	case 0:
		return "OK"
	case 1:
		return "General Error"
	case 2:
		return "Invalid Usage"
	case 126:
		return "Command Cannot Execute"
	case 127:
		return "Command Not Found"
	case 128:
		return "Invalid Exit Argument"
	case 130:
		return "Terminated by Signal"
	case 137:
		return "Killed by Signal"
	case 255:
		return "SSH Connection/Authorization Failed"
	default:
		return fmt.Sprintf("Exit Code %d", exitCode)
	}
}

// Response mirrors the JSON output structure

type Response struct {
	Status        int     `json:"status"`
	Reason        string  `json:"reason"`
	Stdout        *string `json:"stdout"`
	Stderr        *string `json:"stderr"`
	Error         *string `json:"error"`
	CorrelationId string  `json:"correlationId"`
}

// ExecuteRemoteCommand performs the core logic of remote-mac
func ExecuteRemoteCommand(destination, command string) Response {

	// Parse destination as user@host

	parts := strings.Split(destination, "@")
	if len(parts) != 2 {
		errorMsg := "Invalid destination format"
		return Response{Error: &errorMsg, Status: -1, Reason: "SSH Error"}
	}

	user := parts[0]
	host := parts[1]

	// Load private key

	keyPath := "/usr/local/etc/webhook/ssh/id_rsa"

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		errorMsg := "Failed to read private key: " + err.Error()
		return Response{Error: &errorMsg, Status: -1, Reason: "SSH Error"}
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		errorMsg := "Failed to parse private key: " + err.Error()
		return Response{Error: &errorMsg, Status: -1, Reason: "SSH Error"}
	}

	// Configure

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect

	conn, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		errorMsg := "Failed to connect: " + err.Error()
		return Response{Error: &errorMsg, Status: -1, Reason: "SSH Error"}
	}
	defer conn.Close()

	// Create session

	session, err := conn.NewSession()
	if err != nil {
		errorMsg := "Failed to create session: " + err.Error()
		return Response{Error: &errorMsg, Status: -1, Reason: "SSH Error"}
	}
	defer session.Close()

	// Run command

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(command)

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	var stdout *string
	if stdoutStr != "" {
		stdout = &stdoutStr
	}

	var stderr *string
	if stderrStr != "" {
		stderr = &stderrStr
	}

	var exitCode int
	var reason string
	var errorPtr *string

	if err == nil {
		exitCode = 0
		reason = "OK"
	} else if exitErr, ok := err.(*ssh.ExitError); ok {
		exitCode = exitErr.ExitStatus()
		reason = getExitReason(exitCode)
	} else {
		errorMsg := err.Error()
		errorPtr = &errorMsg
		exitCode = -1
		reason = "SSH Error"
	}

	response := Response{
		Stdout: stdout,
		Stderr: stderr,
		Error:  errorPtr,
		Status: exitCode,
		Reason: reason,
	}

	return response
}
