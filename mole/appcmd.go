package mole

import (
	"errors"
	"github.com/joinhack/peony"
	"io"
	"os/exec"
	"strings"
	"time"
)

type AppCmd struct {
	BinPath string
	Addr    string
	cmd     *exec.Cmd
}

func NewAppCmd(app *peony.App, binPath, addr string) *AppCmd {
	appCmd := &AppCmd{BinPath: binPath, Addr: addr}
	appCmd.cmd = exec.Command(appCmd.BinPath, "--bindAddr="+addr,
		"--importPath="+app.ImportPath,
		"--srcPath="+app.SourcePath)
	return appCmd
}

func (a *AppCmd) Run() error {
	return a.cmd.Run()
}

type cmdOutput struct {
	io.Writer
	started chan bool
}

func (c *cmdOutput) Write(b []byte) (int, error) {
	if c.started != nil {
		if strings.Contains(string(b), "Server is running") {
			c.started <- true
			c.started = nil
		}
	}
	return c.Write(b)
}

var (
	TimeOut = errors.New("start app timeout")
	AppDied = errors.New("app died")
)

func (a *AppCmd) Start() error {
	output := &cmdOutput{a.cmd.Stdout, make(chan bool, 1)}
	a.cmd.Stdout = output
	if err := a.cmd.Start(); err != nil {
		return err
	}
	select {
	case <-a.wait():
		return AppDied
	case <-output.started:
		return nil
	case <-time.After(30 * time.Second):
		a.Kill()
		peony.ERROR.Println("start app timeout")
		return TimeOut
	}
}

func (a *AppCmd) wait() <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		a.cmd.Wait()
		ch <- true
	}()
	return ch
}

func (a *AppCmd) Kill() {
	cmd := a.cmd
	if cmd != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited()) {
		if err := cmd.Process.Kill(); err != nil {
			peony.ERROR.Println("kill app error:", err)
		}
	}
}
