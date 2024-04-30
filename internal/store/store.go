// Package store содержит методы для работы с хранилищем - файлом.
package store

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/logging"
)

type configer interface {
	GetFileStoreName() string
	GetDBConnName() string
}

type storage struct {
	urlMap    map[string]string
	revertMap map[string]string
	fileName  string
	rw        sync.RWMutex
}

// New находит или создает файл, инициализирует Map - для хранения.
func New(ctx context.Context, c configer) (*storage, error) {
	fn := ""
	fp := ""
	if c.GetFileStoreName() != "" {
		fn = filepath.Base(c.GetFileStoreName())
		fp = filepath.Dir(c.GetFileStoreName())
	}
	res := storage{
		rw:        sync.RWMutex{},
		fileName:  fn,
		urlMap:    make(map[string]string),
		revertMap: make(map[string]string),
	}

	if fn != "" {
		if err := createIfNeeded(fp, fn); err != nil {
			return nil, fmt.Errorf("create file storage: %w", err)
		}

		if err := res.readFromFile(); err != nil {
			return nil, fmt.Errorf("read from file storage: %w", err)
		}
	}
	return &res, nil
}

// Ping не поддерживается для данного типа хранилища.
//
// Deprecated: не пддерживайется для хранилища - файла.
func (s *storage) Ping(ctx context.Context) error {
	return errors.New("unsupported store method")
}

// CloseDB не поддерживается для данного типа хранилища.
//
// Deprecated: не пддерживайется для хранилища - файла.
func (s *storage) CloseDB() error {
	return nil
}

// GetShortURL ищет по URL его сокращение.
func (s *storage) GetShortURL(ctx context.Context, url string) (string, error) {
	s.rw.RLock()
	generated, isFound := s.urlMap[url]
	s.rw.RUnlock()
	if !isFound {
		return "", nil
	}
	return generated, nil
}

// Add добавляет в файл пару URL - сокращение.
func (s *storage) Add(ctx context.Context, original, short string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	generated, isFound := s.urlMap[original]
	if isFound {
		return cutter.NewUniqueURLError(generated, fmt.Errorf("url already added"))
	}
	s.urlMap[original] = short
	s.revertMap[short] = original
	if s.fileName != "" {
		id := len(s.urlMap) + 1
		if err := writeItem(s.fileName, jsonobject.Item{ID: id, ShortURL: short, OriginalURL: original}); err != nil {
			return fmt.Errorf("store.add: write items: %w", err)
		}
	}
	return nil
}

// GetOriginalURL находит по переданному сокращению оригинальный URL.
func (s *storage) GetOriginalURL(ctx context.Context, value string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	res, isFound := s.revertMap[value]
	if !isFound {
		return "", fmt.Errorf("no data found in urlMap for value %s", value)
	}
	return res, nil
}

// UploadBatch загружает слайс BatchItem в файл.
func (s *storage) UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error) {
	for i := 0; i < len(batch); i++ {
		short, err := s.GetShortURL(ctx, batch[i].OriginalURL)
		if err != nil {
			return batch, fmt.Errorf("upload batch check: %w", err)
		}
		if short != "" {
			batch[i].ShortURL = short
		} else {
			s.Add(ctx, batch[i].OriginalURL, batch[i].ShortURL)
		}
		batch[i].OriginalURL = ""
	}
	return batch, nil
}

// GetUserURLs не поддерживается для данного типа хранилища.
//
// Deprecated: не пддерживайется для хранилища - файла.
func (s *storage) GetUserURLs(ctx context.Context) (jsonobject.Batch, error) {
	return nil, nil
}

// DeleteURLs не поддерживается для данного типа хранилища.
//
// Deprecated: не пддерживайется для хранилища - файла.
func (s *storage) DeleteURLs(ctx context.Context, userID string, ids []string) error {
	return errors.New("unsupported store method")
}

// readFromFile открывает файл на чтение.
// содержимое файла загружается в map.
func (s *storage) readFromFile() error {
	s.rw.RLock()
	defer s.rw.RUnlock()

	c, err := newConsumer(s.fileName)
	if err != nil {
		return fmt.Errorf("readFromFile: open file %w", err)
	}
	items, err := c.ReadItems()
	if err != nil {
		return fmt.Errorf("readFromFile: read file: %w", err)
	}
	for _, item := range items {
		s.urlMap[item.ShortURL] = item.OriginalURL
		s.revertMap[item.OriginalURL] = item.ShortURL
	}

	return nil
}

// Consumer открывает файл на чтение и читает из него.
type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func newConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("newConsumer: open file: %w", err)
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

// ReadItems читает каждую строку в jsonobject.Item.
// Возвращает слайс jsonobject.Item.
func (c *Consumer) ReadItems() ([]jsonobject.Item, error) {
	items := []jsonobject.Item{}
	for c.scanner.Scan() {
		data := c.scanner.Bytes()
		item := jsonobject.Item{}
		err := item.UnmarshalJSON(data)
		if err != nil {
			return nil, fmt.Errorf("unmarshal from file: %w", err)
		}
		items = append(items, item)
	}
	if err := c.scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file error: %w", err)
	}
	return items, nil
}

// writeItem открывает файл на запись и записывает 1 строку информации.
func writeItem(fname string, i jsonobject.Item) error {
	data, err := i.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("writeItem: open file: %w", err)
	}
	defer file.Close()
	data = append(data, '\n')
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("writeItem: write in file: %w", err)
	}
	return nil
}

// createIfNeeded находит файл с именем fileName по пути path.
// если файл или путь не найдены - создаст их.
func createIfNeeded(path string, fileName string) error {
	defer logging.Log.Sync()
	err := os.MkdirAll(path, 0750)
	if err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	logging.Log.Debugf("dir was created: %s ", path)
	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		file, errCreate := os.Create(fileName)
		if err1 := file.Close(); err1 != nil && errCreate == nil {
			errCreate = fmt.Errorf("create file: %w", err1)
		}
		if errCreate == nil {
			logging.Log.Debugf("file was created: %s (path %s)", fileName, path)
		}
		return errCreate
	} else {
		logging.Log.Debugf("file was found: %s (path %s)", fileName, path)
	}

	return err
}
