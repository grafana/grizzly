package grizzly

import (
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

type watch struct {
	path   string
	parent string
	isDir  bool
}

type Watcher struct {
	watcher     *fsnotify.Watcher
	watcherFunc func(string) error
	watches     []watch
}

func NewWatcher(watcherFunc func(path string) error) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	watcher := Watcher{
		watcher:     w,
		watcherFunc: watcherFunc,
	}
	return &watcher, nil
}

func (w *Watcher) Add(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		// `vim` renames and replaces, doesn't create a WRITE event. So we need to watch the whole dir and filter for our file
		parent := filepath.Dir(path)
		w.watches = append(w.watches, watch{path: path, parent: parent, isDir: false})
		err := w.watcher.Add(parent)
		if err != nil {
			return err
		}
	} else {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				w.watches = append(w.watches, watch{path: path, parent: path, isDir: true})
				return w.watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (w *Watcher) Watch() error {
	go func() {
		log.Info("Watching for changes")
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if !w.isFiltered(event.Name) {
						log.Info("Changes detected. Parsing")
						err := w.watcherFunc(event.Name)
						if err != nil {
							log.Error("error: ", err)
						}
					}
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				log.Error("error: ", err)
			}
		}
	}()
	return nil
}

func (w *Watcher) Wait() error {
	done := make(chan bool)
	<-done
	return nil
}

func (w *Watcher) isFiltered(path string) bool {
	parent := filepath.Dir(path)
	for _, watch := range w.watches {
		if parent == watch.parent {
			if watch.isDir || watch.path == path {
				return false
			}
		}
	}
	return true
}
