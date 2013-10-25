package mole

import (
	"testing"
)

func TestProcessSources(t *testing.T) {
	_, err := ProcessSources([]string{"../demos/login"})
	if err != nil {
		t.Log("error:", err)
	}
}
