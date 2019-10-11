package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func handleSingnals(progname string, signals []os.Signal, handler func(os.Signal)) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)

	go func() {
		for sig := range signalChan {
			fmt.Printf("[%s] Received %v ...\n", progname, sig)
			handler(sig)
		}
	}()
}

func statusFromCmd(cmd *exec.Cmd) (*syscall.WaitStatus, error) {
	if cmd.ProcessState != nil {
		status := cmd.ProcessState.Sys().(syscall.WaitStatus)
		return &status, nil
	} else {
		return nil, errors.New("no cmd ProcessState")
	}
}

func handleCmdExit(cmd *exec.Cmd, err error, prefix string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sCommand failed - %v\n", prefix, err)
	}

	var statusErr error = nil
	if cmd.ProcessState != nil {
		var status *syscall.WaitStatus = nil
		status, statusErr = statusFromCmd(cmd)
		if statusErr != nil {
			fmt.Printf("%sFailed Getting Status: %v", prefix, statusErr)
		} else {
			os.Exit(status.ExitStatus())
		}
	}

	if err != nil {
		os.Exit(1)
	}
}
