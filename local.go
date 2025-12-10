package fstorage

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type localoss struct {
	path string
}

func NewLocalOSS(path string) *localoss {
	return &localoss{
		path: path,
	}
}

func (m *localoss) PutReader(
	ctx context.Context,
	ext string,
	r io.Reader,
	l int64,
) (string, error) {
	name := sha(strconv.Itoa(int(time.Now().UnixNano())))
	if strings.HasPrefix(ext, ".") {
		name += ext
	} else {
		name += "." + ext
	}

	if err := os.MkdirAll(filepath.Join(m.path, name[:2], name[2:4]), 0o700); err != nil {
		return "", err
	}

	n := filepath.Join(name[:2], name[2:4], name)

	file, err := os.Create(filepath.Join(m.path, n))
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, r); err != nil {
		return "", err
	}
	return n, nil
}

func (m *localoss) GetReader(
	ctx context.Context,
	name string,
) (io.ReadCloser, error) {
	return os.Open(filepath.Join(m.path, name))
}

func (m *localoss) Del(ctx context.Context, name string) error {
	return os.Remove(filepath.Join(m.path, name))
}

func sha(s string) string {
	m := sha1.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}
