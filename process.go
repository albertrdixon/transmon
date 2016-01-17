package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/albertrdixon/gearbox/logger"

	"golang.org/x/net/context"
)

type command struct {
	*exec.Cmd
	name string
	done chan struct{}
	err  chan error
}

func newCommand(name, cmd string) (*command, error) {
	list := strings.Fields(cmd)
	if len(list) < 1 {
		return nil, errors.New("Bad command")
	}

	path, er := exec.LookPath(list[0])
	if er != nil {
		return nil, er
	}
	return &command{
		Cmd:  exec.Command(path, list[1:]...),
		name: name,
		done: make(chan struct{}, 1),
	}, nil
}

func (c *command) Execute(ctx context.Context) error {
	sto, er := c.StdoutPipe()
	if er != nil {
		return er
	}
	ste, er := c.StderrPipe()
	if er != nil {
		return er
	}

	go stream(c.name, sto, ctx)
	go stream(c.name, ste, ctx)

	if er := c.Start(); er != nil {
		return er
	}
	go wait(c)

	go func() {
		select {
		case <-c.done:
			return
		case <-ctx.Done():
			c.Stop()
			c.Process.Release()
		}
	}()

	return nil
}

func (c *command) SetUser(uid, gid int) {
	c.Cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}
}

func (c *command) Stop() error {
	if c.ProcessState != nil && !c.ProcessState.Exited() {
		if er := c.Process.Kill(); er != nil {
			return er
		}
	}
	return nil
}

func stream(name string, r io.Reader, c context.Context) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		select {
		case <-c.Done():
			return
		default:
			fmt.Printf("[%s] %s\n", name, s.Text())
		}
	}
}

func wait(c *command) {
	if er := c.Wait(); er != nil {
		logger.Errorf("%s: %v", c.name, er)
	}
	c.done <- struct{}{}
	close(c.done)
}
