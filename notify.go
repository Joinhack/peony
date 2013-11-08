package peony

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
	"strings"
)

type Observer interface {
	Refresh() error
	Path() string
}

type IgnoreObserver interface {
	Observer
	IgnoreDir(file os.FileInfo) bool
	IgnoreFile(file string) bool
}

type observerWatcher struct {
	observer Observer
	watcher  *fsnotify.Watcher
	path     string
}

type Notifier struct {
	watchers []*observerWatcher
}

func NewNotifier() *Notifier {
	n := &Notifier{}
	n.watchers = []*observerWatcher{}
	return n
}

func (n *Notifier) contain(abspath string) bool {
	for _, obswatcher := range n.watchers {
		if obswatcher.path == abspath {
			return true
		}
	}
	return false
}

func (n *Notifier) Watch(o Observer) {
	var err error
	var abspath string
	abspath, err = filepath.Abs(o.Path())
	if err != nil {
		ERROR.Println("create watcher error:", err)
		return
	}
	if n.contain(abspath) {
		return
	}
	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	obsWatcher := &observerWatcher{o, watcher, abspath}
	n.watchers = append(n.watchers, obsWatcher)

	var ignoreObserver IgnoreObserver = nil
	var isIgnoreObserver bool
	ignoreObserver, isIgnoreObserver = o.(IgnoreObserver)
	watcher.Watch(abspath)
	filepath.Walk(abspath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if isIgnoreObserver {
			if info.IsDir() && ignoreObserver.IgnoreDir(info) {
				return filepath.SkipDir
			}
		}
		watcher.Watch(path)
		return nil
	})
}

func (n *Notifier) Notify() error {
	for _, obswatcher := range n.watchers {
		select {
		case evt := <-obswatcher.watcher.Event:
			//ignore file name start with "." like ".xxx"
			if !strings.HasPrefix(filepath.Base(evt.Name), ".") {
				if ignoreObserver, ok := obswatcher.observer.(IgnoreObserver); ok {
					if ignoreObserver.IgnoreFile(evt.Name) {
						continue
					}
				}
			}
		case err := <-obswatcher.watcher.Error:
			return err
		default:
			continue
		}
		if err := obswatcher.observer.Refresh(); err != nil {
			return nil
		}
	}
	return nil
}
