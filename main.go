package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/kballard/go-shellquote"
	"github.com/subosito/gotenv"
)

type Config map[string]string

// from https://github.com/ddollar/forego
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

func command_prep(parts ...string) *exec.Cmd {
	command := parts[0]
	remaining_parts := parts[1:len(parts)]
	return exec.Command(command, remaining_parts...)
}

func main() {
	flag.Parse()

	if config, error := ReadConfig("Procfile"); error != nil {
		fmt.Fprintf(os.Stderr, "[forerun] Failed reading Procfile: %v\n", error)
		os.Exit(1)
	} else {
		requested_cmd := flag.Args()[0]
		if cmd_string, ok := config[requested_cmd]; !ok {
			keys := reflect.ValueOf(config).MapKeys()
			fmt.Fprintf(os.Stderr, "[forerun] Entry not found in Procfile: %v (valid commands: %v)\n", requested_cmd, keys)
			os.Exit(1)
		} else {
			words, error := shellquote.Split(cmd_string)
			if error != nil {
				fmt.Fprintf(os.Stderr, "[forerun] Failed parsing command: %+v", error)
				os.Exit(1)
			}
			cmd := command_prep(words...)
			fmt.Printf("[forerun] Running command: %v ...\n", words)
			err := cmd.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[forerun] Command failed - %v\n", err)
				os.Exit(1)
			}
		}
	}
}
