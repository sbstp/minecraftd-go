package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type minecraftd struct {
	proc   *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	tasks  chan Task
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

func (m *minecraftd) addTask(t Task) {
	m.tasks <- t
}

func (m *minecraftd) taskExecutor() {
	for {
		select {
		case task := <-m.tasks:
			// When we receive a new task, pass the control flow to the task and it can use
			// the lines channel as it sees fit. Thus, it gains exclusive access to the lines.
			stop, err := task.Execute(m.lines, m.stdin)
			if err != nil {
				log.Println("Error during task execution:", err)
				break
			}
			if stop {
				log.Println("Task has asked to stop the task executor")
				break
			}
		case line, ok := <-m.lines:
			if !ok {
				break
			}
			// No task is executing and we receive a line, just ignore it.
			_ = line
		}
	}
}

func (m *minecraftd) signalHandler() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	log.Println("Caught interrupt, queuing stop task.")
	m.addTask(NewStopTask())
}

func spawnMinecraftProcess() error {
	cmd := exec.Command("java", "-jar", "server.jar", "nogui")
	cmd.Dir = "world"
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

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
		make(chan Task, 1000),
		make(chan string, 1000),
	}

	m.addTask(NewLoadTask())

	go m.stdoutParser()
	go m.taskExecutor()
	go m.signalHandler()

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
