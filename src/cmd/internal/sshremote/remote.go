// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package sshremote

import (
	"bytes"
	"fmt"

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
func ExecuteRemoteCommand(destination string, clientConfig *ssh.ClientConfig, command string) Response {

	conn, err := ssh.Dial("tcp", destination, clientConfig)
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
