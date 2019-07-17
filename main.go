package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {

	if len(os.Args) < 3 {
		log.Fatalf("usage: %s [output delay] [command]", os.Args[0])
	}

	outputDelay, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatalf("failed to get output delay: %s", err)
	}
	args := os.Args[2:]

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = outW
	cmd.Stderr = errW

	err = cmd.Start()
	if err != nil {
		log.Fatalf("failed to start process: %s", err)
	}
	forwardInterrupts(cmd)

	delayTimer := time.NewTimer(time.Duration(outputDelay) * time.Second)
	checkTimer := time.NewTimer(time.Duration(1 / 10 * time.Second))

End:
	for {
		select {
		case <-delayTimer.C:
			break End
		case <-checkTimer.C:
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				_, _ = io.Copy(os.Stdout, outR)
				_, _ = io.Copy(os.Stderr, errR)
				os.Exit(cmd.ProcessState.ExitCode())
			}
			checkTimer.Reset(time.Duration(1 / 10 * time.Second))
		}
	}

	go func() { _, _ = io.Copy(os.Stdout, outR) }()
	go func() { _, _ = io.Copy(os.Stderr, errR) }()

	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			os.Exit(waitStatus.ExitStatus())
		}
		log.Fatalf("waiting for program exit failed: %s", err)
		os.Exit(127)
	}

	os.Exit(0)
}

func forwardInterrupts(cmd *exec.Cmd) {
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		for s := range gracefulStop {
			err := cmd.Process.Signal(s)
			if err != nil {
				log.Fatalf("failed to signal process: %s", err)
			}
		}
	}()
}
