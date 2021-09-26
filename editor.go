package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func editor(val string) (string, error) {
	f, err := os.CreateTemp("", "plan_edit")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if err := os.WriteFile(f.Name(), []byte(val), 0755); err != nil {
		return "", err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	editorCmd := append(strings.Split(editor, " "), f.Name())
	cmd := exec.Command(editorCmd[0], editorCmd[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	newval, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	return string(newval), nil
}
