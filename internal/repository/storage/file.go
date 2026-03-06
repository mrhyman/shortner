package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

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

func (fs *FileStorage) Store(ctx context.Context, link model.Link) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.links = append(fs.links, link)
	return fs.save()
}

func (fs *FileStorage) StoreBatch(ctx context.Context, ls []model.Link) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.links = append(fs.links, ls...)
	return fs.save()
}

func (fs *FileStorage) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	for _, link := range fs.links {
		if link.ShortURL == shortURL {
			return &link, nil
		}
	}

	return nil, model.ErrNotFound
}

func (fs *FileStorage) GetByUserID(ctx context.Context, userID string) ([]model.Link, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var links []model.Link

	for _, link := range fs.links {
		if link.UserID == userID {
			links = append(links, link)
		}
	}

	return links, nil
}

func (fs *FileStorage) DeleteUserLinksByID(ctx context.Context, links []string) error {
	userID, ok := ctx.Value(model.UserIDKey).(string)
	if !ok {
		return model.ErrUnknownUser
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	toDelete := make(map[string]struct{}, len(links))
	for _, l := range links {
		toDelete[l] = struct{}{}
	}

	for i := range fs.links {
		if fs.links[i].UserID != userID {
			continue
		}

		if _, ok := toDelete[fs.links[i].ShortURL]; ok {
			fs.links[i].IsDeleted = true
		}
	}

	return fs.save()
}

func (fs *FileStorage) CountURLs(ctx context.Context) (int, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	count := 0
	for _, link := range fs.links {
		if !link.IsDeleted {
			count++
		}
	}

	return count, nil
}

func (fs *FileStorage) CountUsers(ctx context.Context) (int, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	users := make(map[string]struct{})
	for _, link := range fs.links {
		if link.IsDeleted {
			continue
		}
		users[link.UserID] = struct{}{}
	}

	return len(users), nil
}

func (fs *FileStorage) Ping() error {
	_, err := os.Stat(fs.path)
	return err
}

func (fs *FileStorage) Close() error {
	return nil
}

func (fs *FileStorage) save() error {
	data, err := json.Marshal(fs.links)
	if err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0644)
}
