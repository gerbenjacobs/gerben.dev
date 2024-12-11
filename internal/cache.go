package internal

import (
	"errors"
	"io"
	"os"
	"time"
)

var (
	ErrCacheExpired = errors.New("cache expired")
	ErrCacheCreated = errors.New("cache created")
)

func GetCache(filePath string, expiry time.Duration) ([]byte, error) {
	info, err := os.Stat(filePath)
	switch {
	case os.IsNotExist(err):
		_, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		return nil, ErrCacheCreated
	case err != nil:
		return nil, err
	case info.ModTime().Before(time.Now().Add(-10 * expiry)):
		return nil, ErrCacheExpired
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func SetCache(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}
