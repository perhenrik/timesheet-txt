package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	binaryPath   = "./tmp/timesheet"
	pollInterval = 300 * time.Millisecond
)

func main() {
	if err := os.MkdirAll("tmp", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create tmp directory: %v\n", err)
		os.Exit(1)
	}

	runner := &appRunner{}

	if err := buildBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "initial build failed: %v\n", err)
		os.Exit(1)
	}
	if err := runner.start(); err != nil {
		fmt.Fprintf(os.Stderr, "initial start failed: %v\n", err)
		os.Exit(1)
	}

	lastState, err := computeGoFileState(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to compute initial file state: %v\n", err)
		_ = runner.stop()
		os.Exit(1)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-ticker.C:
			currentState, stateErr := computeGoFileState(".")
			if stateErr != nil {
				fmt.Fprintf(os.Stderr, "watch error: %v\n", stateErr)
				continue
			}

			if currentState == lastState {
				continue
			}

			lastState = currentState
			fmt.Fprintln(os.Stderr, "rebuilding...")
			if buildErr := buildBinary(); buildErr != nil {
				fmt.Fprintf(os.Stderr, "build failed: %v\n", buildErr)
				continue
			}

			if restartErr := runner.restart(); restartErr != nil {
				fmt.Fprintf(os.Stderr, "restart failed: %v\n", restartErr)
				continue
			}
			fmt.Fprintln(os.Stderr, "restarted")
		case <-sigCh:
			_ = runner.stop()
			return
		}
	}
}

func buildBinary() error {
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type appRunner struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	waitCh chan error
}

func (r *appRunner) start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cmd := exec.Command(binaryPath, "tui")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	r.cmd = cmd
	r.waitCh = waitCh
	return nil
}

func (r *appRunner) stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopLocked()
}

func (r *appRunner) restart() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.stopLocked(); err != nil {
		return err
	}

	cmd := exec.Command(binaryPath, "tui")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	r.cmd = cmd
	r.waitCh = waitCh
	return nil
}

func (r *appRunner) stopLocked() error {
	if r.cmd == nil || r.waitCh == nil {
		return nil
	}

	_ = r.cmd.Process.Signal(os.Interrupt)

	select {
	case err := <-r.waitCh:
		r.cmd = nil
		r.waitCh = nil
		if err == nil {
			return nil
		}
		if isExitErr(err) {
			return nil
		}
		return err
	case <-time.After(1200 * time.Millisecond):
		_ = r.cmd.Process.Kill()
		err := <-r.waitCh
		r.cmd = nil
		r.waitCh = nil
		if err == nil || isExitErr(err) {
			return nil
		}
		return err
	}
}

func isExitErr(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

func computeGoFileState(root string) (string, error) {
	entries := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "tmp" || name == "bin" || name == "release" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}

		entries = append(entries, fmt.Sprintf("%s|%d|%d", path, info.ModTime().UnixNano(), info.Size()))
		return nil
	})
	if err != nil {
		return "", err
	}

	return strings.Join(entries, "\n"), nil
}
