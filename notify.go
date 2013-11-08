package peony

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
)

type Observer interface {
	Action()
	Path() string
}

type IgnoreObserver interface {
	Observer
	IgnoreDir(file os.FileInfo) bool
	IgnoreFile(file os.FileInfo) bool
}

type Notifier struct {
	watcher *fsnotify.Watcher
}

func NewNotifier() *Notifier {
	n := &Notifier{}
	return n
}

func (n *Notifier) Watch(o Observer) {
	var ignoreObserver IgnoreObserver = nil
	var isIgnoreObserver bool
	ignoreObserver, isIgnoreObserver = o.(IgnoreObserver)
	n.watcher.Watch(o.Path())
	filepath.Walk(o.Path(), func(path string, info os.FileInfo, err error) error {
		if isIgnoreObserver {
			if info.IsDir() && ignoreObserver.IgnoreDir(info) {
				return filepath.SkipDir
			}
			if ignoreObserver.IgnoreFile(info) {
				return
			}
		}
		n.watcher.Watch(path)
	})
}

func (n *Notifier) Notify() {

}
