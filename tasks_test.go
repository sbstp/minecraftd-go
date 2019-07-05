package main

import (
	"io/ioutil"
	"testing"
)

func TestLoadTask(t *testing.T) {
	lines := make(chan string, 100)
	lines <- `[21:12:57] [Server thread/INFO]: Time elapsed: 12173 ms`
	lines <- `[21:12:57] [Server thread/INFO]: Done (24.298s)! For help, type "help"`
	lines <- `[21:13:00] [Server thread/INFO]: Stopping server`
	close(lines)

	task := NewLoadTask()
	stop, err := task.Execute(lines, ioutil.Discard)

	if err != nil {
		t.Error(err)
	}

	if stop {
		t.Error("Load task should not issue a stop return value")
	}

	if <-lines != `[21:13:00] [Server thread/INFO]: Stopping server` {
		t.Error("Consumed too many lines")
	}
}

func TestLoadTaskFail(t *testing.T) {
	lines := make(chan string, 100)
	close(lines)

	task := NewLoadTask()
	stop, err := task.Execute(lines, ioutil.Discard)

	if err != ErrEOF {
		t.Error("Did not return an error")
	}

	if stop {
		t.Error("Load task should not issue a stop return value")
	}
}
