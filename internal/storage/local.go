package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

func (s *LocalStorage) SaveOriginal(id, filename string, data []byte) (string, error) {
	dir := filepath.Join(s.basePath, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create image dir: %w", err)
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return path, nil
}

func (s *LocalStorage) Delete(path string) error {
	return os.RemoveAll(path)
}
