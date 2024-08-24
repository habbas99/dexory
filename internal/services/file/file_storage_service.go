package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileStorageService struct{}

func NewFileStorageService() *FileStorageService {
	return &FileStorageService{}
}

func (fs *FileStorageService) SaveFile(dirPath, fileName string, fileContent io.Reader) (*os.File, error) {
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory=%s, error: %w", dirPath, err)
	}

	filePath := filepath.Join(dirPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file=%s, error: %w", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to copy content to file=%s, error: %w", filePath, err)
	}

	return file, nil
}

func (fs *FileStorageService) CreateFile(dirPath, fileName string) (*os.File, error) {
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory=%s, error: %w", dirPath, err)
	}

	filePath := filepath.Join(dirPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file=%s, error: %w", filePath, err)
	}

	return file, nil
}
