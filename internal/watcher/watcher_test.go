package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsChange(t *testing.T) {
	dir := t.TempDir()
	w := New([]string{dir})
	defer w.Close()

	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "test.css"), []byte("body{}"), 0644)

	select {
	case <-w.Events:
		// ok
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for file change")
	}
}

func TestWatcherDebounce(t *testing.T) {
	dir := t.TempDir()
	w := New([]string{dir})
	defer w.Close()

	time.Sleep(50 * time.Millisecond)

	// rapid writes
	for i := 0; i < 20; i++ {
		os.WriteFile(filepath.Join(dir, "test.css"), []byte("body{}"), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	count := 0
	timeout := time.After(2 * time.Second)
loop:
	for {
		select {
		case <-w.Events:
			count++
		case <-timeout:
			break loop
		}
	}
	if count == 0 {
		t.Error("expected at least 1 event")
	}
	if count > 3 {
		t.Errorf("expected <=3 debounced events, got %d", count)
	}
}

func TestWatcherSkipsNodeModules(t *testing.T) {
	dir := t.TempDir()
	nm := filepath.Join(dir, "node_modules")
	os.MkdirAll(nm, 0755)

	w := New([]string{dir})
	defer w.Close()

	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(nm, "lib.js"), []byte("x"), 0644)

	select {
	case <-w.Events:
		t.Error("should not fire event for node_modules")
	case <-time.After(500 * time.Millisecond):
		// expected
	}
}
