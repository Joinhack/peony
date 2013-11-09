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
	*exec.Cmd
	BinPath string
	Addr    string
}

func NewAppCmd(app *peony.App, binPath, addr string) *AppCmd {
	appCmd := &AppCmd{BinPath: binPath, Addr: addr}
	appCmd.Cmd = exec.Command(appCmd.BinPath, "--bindAddr="+addr,
		"--importPath="+app.ImportPath,
		"--srcPath="+app.SourcePath)
	return appCmd
}

func (a *AppCmd) Run() error {
	return a.Cmd.Run()
}

type cmdOutput struct {
	io.Writer
	started chan bool
}

func (c *cmdOutput) Write(b []byte) (int, error) {
	if c.started != nil {
		if strings.Contains(string(b), "Server is running") {
			c.started <- true
		}
	}
	return c.Write(b)
}

var (
	TimeOut = errors.New("start app timeout")
	AppDied = errors.New("app died")
)

func (a *AppCmd) Start() error {
	output := &cmdOutput{a.Stdout, make(chan bool, 1)}
	a.Stdout = output
	if err := a.Cmd.Start(); err != nil {
		return err
	}
	wChan := a.waitChan()
	select {
	case <-wChan:
		return AppDied
	case <-time.After(5 * time.Second):
		peony.ERROR.Println("start app timeout")
		a.Kill()
		return TimeOut
	case <-output.started:
		return nil
	}
}

func (a *AppCmd) waitChan() <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		a.Wait()
		ch <- true
	}()
	return ch
}

func (a *AppCmd) Kill() {
	if a.Cmd != nil && (a.ProcessState != nil && a.ProcessState.Exited()) {
		if err := a.Process.Kill(); err != nil {
			peony.ERROR.Println("kill app error:", err)
		}
	}
}
