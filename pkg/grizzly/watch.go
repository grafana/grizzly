package grizzly

import (
	"io/fs"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

type Watcher struct {
	watcher     *fsnotify.Watcher
	watcherFunc func(string) error
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

func (w *Watcher) Watch(path string) error {
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return w.watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		log.Info("Watching for changes")
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Changes detected. Parsing")
					err := w.watcherFunc(event.Name)
					if err != nil {
						log.Error("error: ", err)
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
