package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dmad1989/urlcut/internal/myjsons"
)

type conf interface {
	GetFileStoreName() string
}

type storage struct {
	rw       sync.RWMutex
	fileName string
}

func New(c conf) (*storage, error) {
	fn := filepath.Base(c.GetFileStoreName())
	fp := filepath.Dir(c.GetFileStoreName())
	res := storage{
		rw:       sync.RWMutex{},
		fileName: fn,
	}
	if err := createIfNeeded(fp, fn); err != nil {
		return nil, fmt.Errorf("fail to create storage: %w", err)
	}
	return &res, nil
}

func (s *storage) Get(key string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	items, err := readItems(s.fileName)
	if err != nil {
		return "", fmt.Errorf("fail read items in Get: %w", err)
	}
	for _, item := range items {
		if item.ShortURL == key {
			return item.OriginalURL, nil
		}
	}
	return "", nil
}

func (s *storage) Add(key, value string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	items, err := readItems(s.fileName)
	if err != nil {
		return fmt.Errorf("fail read items in Add: %w", err)
	}
	id := len(items) + 1
	items = append(items, myjsons.StoreItem{ID: id, ShortURL: key, OriginalURL: value})

	if err := writeItems(s.fileName, items); err != nil {
		return fmt.Errorf("fail write items in Add: %w", err)
	}
	return nil
}

func (s *storage) GetKey(value string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	items, err := readItems(s.fileName)
	if err != nil {
		return "", fmt.Errorf("fail read items in GetByKey: %w", err)
	}
	for _, item := range items {
		if item.OriginalURL == value {
			return item.ShortURL, nil
		}
	}
	return "", fmt.Errorf("no data found in store for value %s", value)
}

func readItems(fname string) (myjsons.StoreItemSlice, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if len(b) == 0 {
		return nil, nil
	}

	items := myjsons.StoreItemSlice{}
	err = items.UnmarshalJSON(b)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal from file: %w", err)
	}
	return items, nil
}

func writeItems(fname string, items myjsons.StoreItemSlice) error {

	data, err := items.MarshalJSON()
	if err != nil {
		return fmt.Errorf("fail marshal items: %w", err)
	}

	err = os.WriteFile(fname, data, 066)
	if err != nil {
		return fmt.Errorf("failed to open file to write: %w", err)
	}

	return nil
}

func createIfNeeded(path string, file string) error {
	curPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("fail get curPath: %w", err)
	}

	newPath := curPath + path
	err = os.MkdirAll(newPath, 0750)
	if err != nil {
		return fmt.Errorf("fail mkdir: %w", err)
	}

	err = os.Chdir(newPath)
	if err != nil {
		return fmt.Errorf("fail chdir: %w", err)
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		file, err := os.Create(file)
		if err1 := file.Close(); err1 != nil && err == nil {
			err = fmt.Errorf("fail create file: %w", err1)
		}
		return err
	}

	return err
}
