package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if stdout, err := os.Stdout.Stat(); err == nil && stdout.Mode()&os.ModeDevice == 0 {
		log.Fatal("pls don't redirect stdout")
	}
	if stdin, err := os.Stdin.Stat(); err == nil && stdin.Mode()&os.ModeDevice == 0 {
		log.Fatal("pls don't redirect stdin")
	}
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "be verbose")
	flag.Parse()

	originalBytes, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	originalText := string(originalBytes)
	originalResources, err := parse(strings.NewReader(originalText))
	if err != nil {
		log.Fatal(err)
	}
	seen := make(map[string]int)
	for i, r := range originalResources {
		if r.ID == "" {
			log.Fatalf("missing ID for resource on line %d", i+1)
		}
		if prev, ok := seen[r.ID]; ok {
			log.Fatalf("duplicate resource #%s at line %d (previously seen on line %d)", r.ID, i+1, prev)
		}
		seen[r.ID] = i
	}
	edited, err := editor(originalText)
	if err != nil {
		log.Fatal(err)
	}
	desiredState, err := parse(strings.NewReader(edited))
	if err != nil {
		log.Fatal(err)
	}

	plan := diff(originalResources, desiredState)

	if len(plan) == 0 {
		fmt.Println("No changes to make.")
		return
	}
	var (
		creates Plans
		updates Plans
		deletes Plans
	)
	for _, p := range plan {
		switch p.Change {
		case Create:
			creates = append(creates, p)
		case Update:
			updates = append(updates, p)
		case Delete:
			deletes = append(deletes, p)
		}
	}
	if len(creates) > 0 {
		fmt.Println(color(fmt.Sprintf("will CREATE:\n%s", creates), Green))
	}
	if len(updates) > 0 {
		fmt.Println(color(fmt.Sprintf("will UPDATE:\n%s", updates), Cyan))
	}
	if len(deletes) > 0 {
		fmt.Println(color(fmt.Sprintf("will DELETE:\n%s", deletes), Red))
	}

	// Allow the user to back out at this point,
	// if they're not happy with the human-readable diff summary
	fmt.Println("\npress ENTER to continue, Ctrl-C or q to quit")
	confirm, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	if strings.TrimSpace(confirm) == "q" {
		return
	}

	// Record an approved JSON diff in a temp file,
	// just in case something downstream fails,
	// we don't want the user to have to retype all their changes
	out, _ := json.Marshal(plan)
	f, err := os.CreateTemp("", "plan_")
	if err == nil {
		os.WriteFile(f.Name(), out, 0755)
		fmt.Printf("\nplan written to %s, just in case\n", f.Name())
	}

	if verbose {
		os.Stdout.Write(out)
	}

	commandString := promptCommand()
	cmd := exec.Command("sh", "-c", commandString)
	cmd.Stdin = bytes.NewReader(out)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
