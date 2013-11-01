package mole

import (
	"path/filepath"
	"testing"
)

func TestProcessSources(t *testing.T) {
	path, _ := filepath.Abs("../demos/login")
	si, err := ProcessSources([]string{path})
	if err != nil {
		t.Log("error:", err)
	}
	t.Log(getAlais(si))
}
