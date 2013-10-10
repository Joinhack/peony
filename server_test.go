package peony

import (
	"testing"
)

func TestServer(t *testing.T) {
	svr := NewServer()
	svr.Run()
}
