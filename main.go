package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"syscall"

	"github.com/subosito/gotenv"
)

// Config type based on what is in the Procfile
type Config map[string]string

// ReadConfig from https://github.com/ddollar/forego
func ReadConfig(filename string) (Config, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return make(Config), nil
	}
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	config := make(Config)
	for key, val := range gotenv.Parse(fd) {
		config[key] = val
	}
	return config, nil
}

func commandPrep(parts ...string) *exec.Cmd {
	command := parts[0]
	remainingParts := parts[1:len(parts)]
	return exec.Command(command, remainingParts...)
}

func main() {
	flag.Parse()

	config, error := ReadConfig("Procfile")
	if error != nil {
		fmt.Fprintf(os.Stderr, "[forerun] Failed reading Procfile: %v\n", error)
		os.Exit(1)
	}

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "[forerun] Missing task name parameter\n")
		os.Exit(1)
	}

	requestedCmd := flag.Args()[0]

	cmdString, ok := config[requestedCmd]
	if !ok {
		keys := reflect.ValueOf(config).MapKeys()
		fmt.Fprintf(os.Stderr, "[forerun] Entry not found in Procfile: %v (valid commands: %v)\n", requestedCmd, keys)
		os.Exit(1)
	}

	words := []string{"/bin/bash", "-c", fmt.Sprintf("exec %s", cmdString)}

	cmd := commandPrep(words...)
	fmt.Printf("[forerun] Running command: %v ...\n", words)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	handleSingnals("forerun", []os.Signal{syscall.SIGINT, syscall.SIGTERM}, func(signal os.Signal) {
		// do nothing
	})

	err := cmd.Run()
	handleCmdExit(cmd, err, "[forerun] ")
}
