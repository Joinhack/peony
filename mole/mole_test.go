package mole

import (
	"github.com/joinhack/peony"
	"go/build"
	"path/filepath"
	"testing"
)

func TestProcessSources(t *testing.T) {
	path, _ := filepath.Abs("../demos/login/app")
	si, err := ProcessSources([]string{path})
	if err != nil {
		t.Log("error:", err)
	}
	t.Log(getAlais(si))
}

func TestBuild(t *testing.T) {
	app := peony.NewApp(filepath.Join(build.Default.GOPATH, "src"), "github.com/joinhack/peony/demos/login")
	Build(app)
}
