package main

import (
	"errors"
	"fmt"
	"log"
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
	}
	return nil, errors.New("no cmd ProcessState")
}

func handleCmdExit(cmd *exec.Cmd, err error, prefix string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sCommand Run Failed - %v\n", prefix, err)
	}

	var statusErr error = nil
	if cmd.ProcessState != nil {
		var status *syscall.WaitStatus = nil
		status, statusErr = statusFromCmd(cmd)
		if statusErr != nil {
			fmt.Printf("%sFailed Getting Status: %v\n", prefix, statusErr)
		} else {
			if status.ExitStatus() > -1 {
				os.Exit(status.ExitStatus())
			} else if status.Signal() > -1 {
				fmt.Printf("%sexited with signal: %v\n", prefix, status.Signal())
				os.Exit(0)
			} else {
				fmt.Printf("%sstatus exited: %v\n", prefix, status.Exited())
				fmt.Printf("%sstatus exit status: %v\n", prefix, status.ExitStatus())
				fmt.Printf("%sexited with signal: %v\n", prefix, status.Signal())
				fmt.Printf("%sstatus stop signal: %v\n", prefix, status.StopSignal())
				fmt.Printf("%sstatus stop signal: %v\n", prefix, status.StopSignal())
				fmt.Printf("%sstatus trap cause: %v\n", prefix, status.TrapCause())
				os.Exit(-1)
			}
		}
	}

	if err != nil {
		os.Exit(1)
	}
}

func initLog(progname string) {
	log.SetPrefix(fmt.Sprintf("[%s]: ", progname))
	log.SetFlags(0)
}
