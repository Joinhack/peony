package mole

import (
	"github.com/joinhack/peony"
	"go/build"
	"path/filepath"
	"testing"
)

func TestProcessSources(t *testing.T) {
	path, _ := filepath.Abs("../demos/login/app/")
	si, err := ProcessSources([]string{path})
	if err != nil {
		t.Log("error:", err)
		return
	}
	t.Log(getAlais(si))
}

func TestBuild(t *testing.T) {
	app := peony.NewApp(filepath.Join(build.Default.GOPATH, "src"), "github.com/joinhack/peony/demos/chat")
	app.DevMode = true
	ag, _ := NewAgent(app)
	ag.Run(":8080")
}
