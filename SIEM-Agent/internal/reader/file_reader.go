package reader

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type RawEvent struct {
	Source    string
	Timestamp time.Time
	Data      string
}

type Reader struct {
	path string

	watcher *fsnotify.Watcher
	file    *os.File
	offset  int64

	events chan RawEvent
	errors chan error

	stop chan struct{}
	wg   sync.WaitGroup

	buffer string
}

func New(path string) (*Reader, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	r := &Reader{
		path:    path,
		watcher: w,
		events:  make(chan RawEvent, 100),
		errors:  make(chan error, 10),
		stop:    make(chan struct{}),
	}

	return r, nil
}

func (r *Reader) Events() <-chan RawEvent { return r.events }
func (r *Reader) Errors() <-chan error    { return r.errors }

func (r *Reader) Start() error {
	if err := r.openFile(true); err != nil {
		return err
	}

	if err := r.watcher.Add(r.path); err != nil {
		return err
	}

	r.wg.Add(1)
	go r.loop()

	return nil
}

func (r *Reader) Stop() {
	close(r.stop)
	r.wg.Wait()
	close(r.events)
	close(r.errors)
}

func (r *Reader) loop() {
	defer r.wg.Done()

	for {
		select {
		case <-r.stop:
			return

		case evt := <-r.watcher.Events:
			r.handleFsEvent(evt)

		case err := <-r.watcher.Errors:
			r.errors <- err
		}
	}
}

func (r *Reader) handleFsEvent(evt fsnotify.Event) {
	if evt.Op&fsnotify.Write == fsnotify.Write {
		r.readNew()
		return
	}

	if evt.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		if err := r.reopen(); err != nil {
			r.errors <- err
		}
	}
}

func (r *Reader) openFile(seekEnd bool) error {
	f, err := os.Open(r.path)
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}

	if seekEnd {
		r.offset = info.Size()
		_, err = f.Seek(r.offset, io.SeekStart)
		if err != nil {
			f.Close()
			return err
		}
	} else {
		r.offset = 0
	}

	r.file = f
	return nil
}

func (r *Reader) reopen() error {
	if r.file != nil {
		_ = r.file.Close()
	}

	for {
		select {
		case <-r.stop:
			return nil
		default:
			if err := r.openFile(false); err == nil {
				return nil
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (r *Reader) readNew() {
	if r.file == nil {
		r.errors <- errors.New("file is not open")
		return
	}

	reader := bufio.NewReader(r.file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.buffer += line
				return
			}
			r.errors <- err
			return
		}

		full := r.buffer + line
		r.buffer = ""

		r.offset += int64(len(line))

		r.events <- RawEvent{
			Source:    r.path,
			Timestamp: time.Now(),
			Data:      trimNewline(full),
		}
	}
}

func trimNewline(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
