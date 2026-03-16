package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	SaveFile(filename string, data io.Reader) (string, error)
	GetFile(path string) (io.ReadCloser, error)
	DeleteFile(path string) error
}

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) SaveFile(filename string, data io.Reader) (string, error) {
	fullPath := filepath.Join(s.basePath, filename)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	_, err = io.Copy(dst, data)
	if err != nil {
		os.Remove(fullPath)
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) GetFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (s *LocalStorage) DeleteFile(path string) error {
	return os.Remove(path)
}
