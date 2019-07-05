package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
)

var (
	// ErrEOF signals that we did not expect to receive EOF but we did.
	ErrEOF = errors.New("Received EOF before expected line of output")
)

var (
	reOutputLine        = regexp.MustCompile(`^(\[[\d\:]+\])\s*(\[[^\]]*\])\:\s*((?:.+))$`)
	reOutputLoadingDone = regexp.MustCompile(`^(?i)Done\s\([\d\.]+s\)!.*`)
)

func waitFor(lines chan string, re *regexp.Regexp) bool {
	for line := range lines {
		parts := reOutputLine.FindStringSubmatch(line)
		if parts != nil {
			message := parts[3]
			if re.MatchString(message) {
				return true
			}
		}

	}
	return false
}

// Task represents an executable task that hooks into the IO of the minecraft process to do something.
type Task interface {
	Execute(lines chan string, stdin io.Writer) (bool, error)
}

type loadTask struct {
}

func (t *loadTask) Execute(lines chan string, stdin io.Writer) (bool, error) {
	if !waitFor(lines, reOutputLoadingDone) {
		return false, ErrEOF
	}
	return false, nil
}

// NewLoadTask creates a new task that waits until the minecraft server is finished loading.
func NewLoadTask() Task {
	return &loadTask{}
}

type backupTask struct {
	interval int
	path     string
}

func (t *backupTask) Execute(lines chan string, stdin io.Writer) (bool, error) {
	return false, nil
}

// NewBackupTask creates a new backup task that performs a backup to the given path, every interval.
func NewBackupTask(interval int, path string) Task {
	return &backupTask{interval, path}
}

type stopTask struct {
}

func (t *stopTask) Execute(lines chan string, stdin io.Writer) (bool, error) {
	fmt.Fprintln(stdin, "stop")
	return true, nil
}

// NewStopTask creates a new task that tells the minecraft to stop and exit gracefully.
func NewStopTask() Task {
	return &stopTask{}
}
