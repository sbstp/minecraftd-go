package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type task interface {
	execute(lines chan string)
}

type loadTask struct {
}

func (l loadTask) execute(lines chan string) {
	line := <-lines
	fmt.Println("load", line)
}

type minecraftd struct {
	proc   *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	tasks  chan task
	lines  chan string
}

func (m *minecraftd) processOutputLine(line string) {
	fmt.Print("minecraft: ")
	fmt.Println(line)

	m.lines <- line
}

func (m *minecraftd) stdoutParser() {
	reader := bufio.NewReader(m.stdout)

	for {
		line, err := reader.ReadString(byte('\n'))

		if err != io.EOF && err != nil {
			log.Println("stdout error", err)
			break
		}

		line = strings.TrimSpace(line)
		m.processOutputLine(line)

		if err == io.EOF {
			break
		}
	}

	close(m.lines)
}

func (m *minecraftd) addTask(t task) {
	m.tasks <- t
}

func (m *minecraftd) taskExecutor() {
	for {
		select {
		case task := <-m.tasks:
			// When we receive a new task, pass the control flow to the task and it can use
			// the lines channel as it sees fit. Thus, it gains exclusive access to the lines.
			task.execute(m.lines)
		case line, ok := <-m.lines:
			if !ok {
				break
			}
			// No task is executing and we receive a line, just ignore it.
			fmt.Println("dumped", len(line), line)
		}
	}
}

func spawnMinecraftProcess() error {
	cmd := exec.Command("ls", "-l")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	m := minecraftd{
		cmd,
		stdin,
		stdout,
		make(chan task, 1000),
		make(chan string, 1000),
	}

	m.addTask(loadTask{})
	m.addTask(loadTask{})

	go m.stdoutParser()
	go m.taskExecutor()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := spawnMinecraftProcess(); err != nil {
		log.Fatal(err)
	}
}
