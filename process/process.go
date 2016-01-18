package process

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

type Process struct {
	*exec.Cmd
	name string
	done chan struct{}
	err  chan error
}

func New(name, cmd string) (*Process, error) {
	list := strings.Fields(cmd)
	if len(list) < 1 {
		return nil, errors.New("Bad command")
	}

	path, er := exec.LookPath(list[0])
	if er != nil {
		return nil, er
	}
	return &Process{
		Cmd:  exec.Command(path, list[1:]...),
		name: name,
		done: make(chan struct{}, 1),
	}, nil
}

func (p *Process) Execute(ctx context.Context) error {
	sto, er := p.StdoutPipe()
	if er != nil {
		return er
	}
	ste, er := p.StderrPipe()
	if er != nil {
		return er
	}

	go stream(p.name, sto, ctx)
	go stream(p.name, ste, ctx)

	if er := p.Start(); er != nil {
		return er
	}
	go wait(p)

	go func() {
		select {
		case <-p.done:
			return
		case <-ctx.Done():
			p.Stop()
			p.Process.Release()
		}
	}()

	return nil
}

func (p *Process) SetUser(uid, gid int) {
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}
}

func (p *Process) Stop() error {
	if p.ProcessState != nil && !p.ProcessState.Exited() {
		if er := p.Process.Kill(); er != nil {
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

func wait(p *Process) {
	if er := p.Wait(); er != nil {
		logger.Errorf("%s: %v", p.name, er)
	}
	p.done <- struct{}{}
	close(p.done)
}
