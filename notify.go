package peony

import (
	"github.com/howeyc/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Observer interface {
	Refresh() error
	ForceRefresh() bool
	Path() []string
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
	watchers  []*observerWatcher
	lastError Error
	mutex     *sync.Mutex
}

func NewNotifier() *Notifier {
	n := &Notifier{}
	n.watchers = []*observerWatcher{}
	n.mutex = &sync.Mutex{}
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
	for _, basePath := range o.Path() {
		abspath, err = filepath.Abs(basePath)
		if err != nil {
			ERROR.Println("create watcher error:", err)
			continue
		}
		if n.contain(abspath) {
			continue
		}
		var watcher *fsnotify.Watcher
		watcher, err = fsnotify.NewWatcher()
		//collect 100 events
		watcher.Event = make(chan *fsnotify.FileEvent, 200)
		//collect 10 error
		watcher.Error = make(chan error, 20)
		obsWatcher := &observerWatcher{o, watcher, abspath}
		n.watchers = append(n.watchers, obsWatcher)

		var ignoreObserver IgnoreObserver = nil
		var isIgnoreObserver bool
		ignoreObserver, isIgnoreObserver = o.(IgnoreObserver)
		watcher.Watch(abspath)
		err = filepath.Walk(abspath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
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
		if err != nil {
			ERROR.Println("watch error:", err)
		}
	}
}

func (n *Notifier) Notify() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	flush := false
	for _, obswatcher := range n.watchers {
		//get all event events store in chan obswatcher.watcher.Event
		for {
			select {
			case evt := <-obswatcher.watcher.Event:
				//ignore file name start with "." like ".xxx"
				if !strings.HasPrefix(filepath.Base(evt.Name), ".") {
					if ignoreObserver, ok := obswatcher.observer.(IgnoreObserver); ok {
						if ignoreObserver.IgnoreFile(evt.Name) {
							continue
						}
					}
					flush = true
				}
				continue
			case <-obswatcher.watcher.Error:
				continue
			default:
			}
			break
		}
		if obswatcher.observer.ForceRefresh() || flush {
			if err := obswatcher.observer.Refresh(); err != nil {
				return err
			}
		}
	}
	return nil
}

func GetNotifyFilter(svr *Server) Filter {
	return func(c *Controller, filter []Filter) {
		if err := svr.notifier.Notify(); err != nil {
			NewErrorRender(err).Apply(c)
			return
		}
		filter[0](c, filter[1:])
	}
}
