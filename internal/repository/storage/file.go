package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
)

type FileStorage struct {
	mu    sync.RWMutex
	path  string
	links []model.Link
}

func NewFileStorage(path string) (*FileStorage, error) {
	fs := &FileStorage{
		path:  path,
		links: make([]model.Link, 0),
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &fs.links); err != nil {
				return nil, err
			}
		}
	}

	return fs, nil
}

func (fs *FileStorage) Store(shortURL, originalURL string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	link := model.Link{
		UUID:        uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	fs.links = append(fs.links, link)
	return fs.save()
}

func (fs *FileStorage) GetByID(id string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for _, link := range fs.links {
		if link.ShortURL == id {
			return link.OriginalURL, nil
		}
	}

	return "", model.ErrNotFound
}

func (fs *FileStorage) save() error {
	data, err := json.Marshal(fs.links)
	if err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0644)
}
