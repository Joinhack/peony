package mole

import (
	"errors"
	"github.com/joinhack/peony"
	"io"
	"os"
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
	devMode := "false"
	if app.DevMode {
		devMode = "true"
	}
	appCmd.Cmd = exec.Command(appCmd.BinPath, "--bindAddr="+addr,
		"--importPath="+app.ImportPath,
		"--srcPath="+app.SourcePath,
		"--devMode="+devMode)
	return appCmd
}

type cmdOutput struct {
	io.Writer
	started chan bool
}

//intercept the stdout.
func (c *cmdOutput) Write(b []byte) (int, error) {
	if c.started != nil {
		if strings.Contains(string(b), "Server is running, listening on") {
			c.started <- true
			c.started = nil
		}
	}
	return c.Writer.Write(b)
}

var (
	TimeOut = errors.New("start app timeout")
	AppDied = errors.New("app died")
)

func (a *AppCmd) Start() error {
	output := &cmdOutput{os.Stdout, make(chan bool, 1)}
	a.Stdout = output
	a.Stderr = os.Stderr
	if err := a.Cmd.Start(); err != nil {
		return err
	}
	select {
	case <-a.waitChan():
		return AppDied
	case <-time.After(30 * time.Second):
		peony.ERROR.Println("start app timeout")
		a.Kill()
		return TimeOut
	case <-output.started:
		return nil
	}
}

//wait for program exit.
func (a *AppCmd) waitChan() <-chan bool {
	ch := make(chan bool, 1)
	go func() {
		if err := a.Wait(); err != nil {
			peony.ERROR.Println("wait error:", err)
		}
		ch <- true
	}()
	return ch
}

func (a *AppCmd) Kill() {
	if a.Cmd != nil && a.Process != nil && (a.ProcessState == nil || a.ProcessState.Exited()) {
		if err := a.Process.Kill(); err != nil {
			peony.ERROR.Println("kill app error:", err)
		}
	}
}
