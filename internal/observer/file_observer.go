package observer

import (
	"encoding/json"
	"os"
	"sync"
)

type FileObserver struct {
	mu   sync.Mutex
	file *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileObserver{file: f}, nil
}

func (o *FileObserver) OnRequest(event Event) {
	o.mu.Lock()
	defer o.mu.Unlock()

	data, _ := json.Marshal(event)
	o.file.Write(append(data, '\n'))
}

func (o *FileObserver) Close() error {
	return o.file.Close()
}
