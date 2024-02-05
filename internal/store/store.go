package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/dmad1989/urlcut/internal/myjsons"
)

type conf interface {
	GetFileStoreName() string
}

type storage struct {
	rw        sync.RWMutex
	urlMap    map[string]string
	revertMap map[string]string
	fileName  string
}

func New(c conf) (*storage, error) {
	fn := filepath.Base(c.GetFileStoreName())
	fp := filepath.Dir(c.GetFileStoreName())
	res := storage{
		rw:        sync.RWMutex{},
		fileName:  fn,
		urlMap:    make(map[string]string),
		revertMap: make(map[string]string),
	}

	if err := createIfNeeded(fp, fn); err != nil {
		return nil, fmt.Errorf("fail to create storage: %w", err)
	}

	if err := res.readFromFile(); err != nil {
		return nil, fmt.Errorf("fail to read from storage: %w", err)
	}

	return &res, nil
}

func (s *storage) Get(key string) (string, error) {
	s.rw.RLock()
	generated, isFound := s.urlMap[key]
	s.rw.RUnlock()
	if !isFound {
		return "", nil
	}
	return generated, nil
}

func (s *storage) Add(key, value string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.urlMap[key] = value
	s.revertMap[value] = key
	id := len(s.urlMap) + 1

	if err := writeItem(s.fileName, myjsons.StoreItem{ID: id, ShortURL: key, OriginalURL: value}); err != nil {
		return fmt.Errorf("fail write items in Add: %w", err)
	}
	return nil
}

func (s *storage) GetKey(value string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	res, isFound := s.revertMap[value]
	if !isFound {
		return "", fmt.Errorf("no data found in urlMap for value %s", value)
	}
	return res, nil
}

func (s *storage) readFromFile() error {
	s.rw.RLock()
	defer s.rw.RUnlock()

	c, err := NewConsumer(s.fileName)
	if err != nil {
		return fmt.Errorf("failed to open file for read: %w", err)
	}
	items, err := c.ReadItems()
	if err != nil {
		return fmt.Errorf("failed to read from file: %w", err)
	}
	for _, item := range items {
		s.urlMap[item.ShortURL] = item.OriginalURL
		s.revertMap[item.OriginalURL] = item.ShortURL
	}

	return nil
}

type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadItems() ([]myjsons.StoreItem, error) {
	items := []myjsons.StoreItem{}
	for c.scanner.Scan() {
		data := c.scanner.Bytes()
		item := myjsons.StoreItem{}
		err := item.UnmarshalJSON(data)
		if err != nil {
			return nil, fmt.Errorf("failed unmarshal from file: %w", err)
		}
		items = append(items, item)
	}
	if err := c.scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file error: %v", err)
	}
	return items, nil
}

func writeItem(fname string, item myjsons.StoreItem) error {
	data, err := item.MarshalJSON()
	if err != nil {
		return fmt.Errorf("fail marshal item: %w", err)
	}

	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file to write: %w", err)
	}
	defer file.Close()
	data = append(data, '\n')
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write in file: %w", err)
	}
	return nil
}

func createIfNeeded(path string, fileName string) error {
	err := os.MkdirAll(path, 0750)
	if err != nil {
		return fmt.Errorf("fail mkdir: %w", err)
	}
	logging.Log.Sugar().Debug("dir was created: %s ", path)
	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("fail chdir: %w", err)
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err1 := file.Close(); err1 != nil && err == nil {
			err = fmt.Errorf("fail create file: %w", err1)
		}
		if err == nil {
			logging.Log.Sugar().Debug("file was created: %s (path %s)", fileName, path)
		}
		return err
	} else {
		logging.Log.Sugar().Debug("file was found: %s (path %s)", fileName, path)
	}

	return err
}
