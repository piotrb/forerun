package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"syscall"

	shellwords "github.com/buildkite/shellwords"
)

var handledSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGHUP,
	syscall.SIGTERM,
	syscall.SIGTTIN,
	syscall.SIGTTOU,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
}

// Config type based on what is in the Procfile
type Config map[string]string

var version = "unspecified"

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
	words, err := shellwords.Split(cmd)
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
	initLog("forerun")

	log.Printf("Version: %v\n", version)

	var useExec = true

	flag.Parse()

	config, error := readProcfile("Procfile")
	if error != nil {
		log.Fatalf("Failed reading Procfile: %v\n", error)
	}

	if len(flag.Args()) == 0 {
		log.Fatalf("Missing task name parameter\n")
	}

	requestedCmd := flag.Args()[0]

	cmdString, ok := config[requestedCmd]
	if !ok {
		keys := reflect.ValueOf(config).MapKeys()
		log.Fatalf("Entry not found in Procfile: %v (valid commands: %v)\n", requestedCmd, keys)
	}

	additionalEnv, newCmdString, addEnvErr := envFromCmd(cmdString)
	if addEnvErr != nil {
		log.Fatalf("Failed getting additional env: %v", addEnvErr)
	}

	match, _ := regexp.MatchString("(;|&&|\\|\\|)", newCmdString)
	if match {
		useExec = false
		log.Printf("found complex command string, not using exec ...\n")
	}

	if useExec {
		log.Printf("Running command using exec ...\n")
		newCmdString = fmt.Sprintf("exec %s", newCmdString)
	}

	words := []string{"/bin/bash", "-c", newCmdString}

	cmd := commandPrep(words...)
	env := append(os.Environ(), additionalEnv...)

	log.Printf("Running command: %v ...\n", words)
	if len(additionalEnv) > 0 {
		log.Printf("Additional env: %v\n", additionalEnv)
	}

	cmd.Env = env

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	handleSingnals("forerun", handledSignals, func(signal os.Signal) {
		log.Printf("Relaying %v to child pid: %v", signal, cmd.Process.Pid)
		cmd.Process.Signal(signal)
	})

	err := cmd.Run()
	handleCmdExit(cmd, err, "[forerun] ")
}
