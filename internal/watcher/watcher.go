package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fw     *fsnotify.Watcher
	Events chan struct{}
	done   chan struct{}
}

func New(watchDirs []string) *Watcher {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		panic("vibeview: fsnotify: " + err.Error())
	}

	w := &Watcher{
		fw:     fw,
		Events: make(chan struct{}, 1),
		done:   make(chan struct{}),
	}

	for _, dir := range watchDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return nil
			}
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || (len(base) > 0 && base[0] == '.') {
				return filepath.SkipDir
			}
			fw.Add(path)
			return nil
		})
	}

	go w.loop()
	return w
}

func (w *Watcher) loop() {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	pending := false

	for {
		select {
		case <-w.fw.Events:
			timer.Reset(100 * time.Millisecond)
			pending = true
		case <-timer.C:
			if pending {
				select {
				case w.Events <- struct{}{}:
				default:
				}
				pending = false
			}
		case <-w.done:
			timer.Stop()
			return
		}
	}
}

func (w *Watcher) Close() {
	close(w.done)
	w.fw.Close()
}
