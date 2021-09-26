package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Prompt the user for output commands to pipe to,
// including fzf completion from the command history for a nice UX
func promptCommand() string {
	cmd := exec.Command("sh", "-c", `touch .plan-execute-history && fzf --prompt "executor command> " --print-query < .plan-execute-history`)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	// If there's only one line, it must be a new command.
	// Let's record it to the history for next time
	output := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	commandString := output[len(output)-1]
	if len(output) <= 1 {
		f, err := os.OpenFile(".plan-execute-history", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(commandString + "\n")
			f.Close()
		}
	}
	return commandString
}
