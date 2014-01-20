package mole

import (
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
