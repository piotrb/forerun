package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"syscall"

	shellwords "github.com/mattn/go-shellwords"
)

// Config type based on what is in the Procfile
type Config map[string]string

// read Procfile and parse it.
func readProcfile(procfile string) (Config, error) {
	content, err := ioutil.ReadFile(procfile)
	if err != nil {
		return nil, err
	}

	procs := Config{}

	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) != 2 || tokens[0][0] == '#' {
			continue
		}
		k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
		procs[k] = v
	}

	return procs, nil
}

func commandPrep(parts ...string) *exec.Cmd {
	command := parts[0]
	remainingParts := parts[1:len(parts)]
	return exec.Command(command, remainingParts...)
}

func envFromCmd(cmd string) ([]string, string, error) {
	words, err := shellwords.Parse(cmd)
	if err != nil {
		return nil, cmd, fmt.Errorf("Failed parsing command words: %v", err)
	}

	done := false
	result := []string{}

	for _, word := range words {
		if !done {
			match, _ := regexp.MatchString("^[^=]+=[^=]+$", word)
			if match {
				result = append(result, word)
			} else {
				done = true
			}
		}
	}

	words = words[len(result):len(words)]

	return result, strings.Join(words, " "), nil
}

func main() {
	flag.Parse()

	config, error := readProcfile("Procfile")
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

	additionalEnv, newCmdString, addEnvErr := envFromCmd(cmdString)
	if addEnvErr != nil {
		fmt.Fprintf(os.Stderr, "Failed getting additional env: %v", addEnvErr)
		os.Exit(1)
	}

	words := []string{"/bin/bash", "-c", fmt.Sprintf("exec %s", newCmdString)}

	cmd := commandPrep(words...)
	env := append(os.Environ(), additionalEnv...)

	fmt.Printf("[forerun] Running command: %v ...\n", words)
	if len(additionalEnv) > 0 {
		fmt.Printf("[forerun] Additional env: %v\n", additionalEnv)
	}

	cmd.Env = env

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	handleSingnals("forerun", []os.Signal{syscall.SIGINT, syscall.SIGTERM}, func(signal os.Signal) {
		// do nothing
	})

	err := cmd.Run()
	handleCmdExit(cmd, err, "[forerun] ")
}
