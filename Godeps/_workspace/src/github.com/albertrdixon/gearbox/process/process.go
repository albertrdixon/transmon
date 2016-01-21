package process

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/context"
)

type Process struct {
	*exec.Cmd
	attr      *syscall.SysProcAttr
	name, bin string
	args      []string
	c         context.Context
	out       []Writer
	stopC     chan struct{}
	er        error
}

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

func New(name, cmd string, out ...Writer) (*Process, error) {
	fields := strings.Fields(cmd)
	if len(fields) < 1 {
		return nil, errors.New("Bad command")
	}

	bin, er := exec.LookPath(fields[0])
	if er != nil {
		return nil, er
	}

	if out == nil || len(out) < 1 {
		out = []Writer{os.Stdout}
	}

	return &Process{
		name:  name,
		bin:   bin,
		args:  fields[1:],
		out:   out,
		stopC: make(chan struct{}, 1),
	}, nil
}

func (p *Process) String() string {
	return fmt.Sprintf("%s(pid=%d)", p.name, p.Pid())
}

func (p *Process) AddWriter(w Writer) {
	if p.out == nil {
		p.out = make([]Writer, 0, 1)
	}
	p.out = append(p.out, w)
}

func (p *Process) Pid() int {
	if p.Process != nil {
		return p.Process.Pid
	}
	return -1
}

func (p *Process) Exited() <-chan struct{} {
	return p.c.Done()
}

func (p *Process) Execute(ctx context.Context) error {
	p.Cmd = exec.Command(p.bin, p.args...)
	if p.attr != nil {
		p.Cmd.SysProcAttr = p.attr
	}

	sto, er := p.StdoutPipe()
	if er != nil {
		return er
	}
	ste, er := p.StderrPipe()
	if er != nil {
		return er
	}

	c, cancel := context.WithCancel(context.Background())
	p.c = c

	go stream(p, sto)
	go stream(p, ste)

	if er := p.Start(); er != nil {
		cancel()
		return er
	}

	go listen(p, ctx)
	go wait(p, cancel)

	return nil
}

func (p *Process) ExecuteAndRestart(ctx context.Context) {
	for {
		if er := p.Execute(ctx); er != nil {
			p.er = er
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-p.stopC:
			return
		case <-p.c.Done():
		}
	}
}

func (p *Process) Stop() {
	p.stopC <- struct{}{}
	close(p.stopC)
}

func (p *Process) SetUser(uid, gid uint32) *Process {
	p.attr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
	return p
}

func (p *Process) Release() error {
	if p.Process != nil {
		return p.Process.Release()
	}
	return nil
}

func (p *Process) Dead() bool {
	return p.ProcessState != nil && p.ProcessState.Exited()
}

func (p *Process) Kill() error {
	if !p.Dead() && p.Process != nil {
		return p.Process.Kill()
	}
	return nil
}

func (p *Process) Signal(sig os.Signal) error {
	if p.Process != nil {
		return p.Process.Signal(sig)
	}
	return nil
}

func (p *Process) Term() error {
	if !p.Dead() {
		if er := p.Signal(syscall.SIGTERM); er != nil {
			return er
		}
		time.Sleep(20 * time.Millisecond)
		return p.Kill()
	}
	return nil
}

func stream(p *Process, r Reader) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		select {
		case <-p.c.Done():
			return
		default:
			for i := range p.out {
				fmt.Fprintf(p.out[i], "[%s] %s\n", p.name, s.Text())
			}
		}
	}
}

func listen(p *Process, ctx context.Context) {
	select {
	case <-p.c.Done():
		return
	case <-ctx.Done():
		p.Kill()
	case <-p.stopC:
		p.Term()
	}
}

func wait(p *Process, cancel context.CancelFunc) {
	p.Wait()
	cancel()
}
