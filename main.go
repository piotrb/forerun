package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/kballard/go-shellquote"
	"github.com/subosito/gotenv"
	"github.com/piotrb/forerun/forego"
)

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

	config, error := forego.ReadConfig("Procfile")
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

	words, error := shellquote.Split(cmdString)
	if error != nil {
		fmt.Fprintf(os.Stderr, "[forerun] Failed parsing command: %+v", error)
		os.Exit(1)
	}

	cmd := commandPrep(words...)
	fmt.Printf("[forerun] Running command: %v ...\n", words)

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[forerun] Command failed - %v\n", err)
		os.Exit(1)
	}
}
