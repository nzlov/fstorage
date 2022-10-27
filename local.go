package fstorage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type localoss struct {
	path string
	lm   map[string]struct{}
}

func NewLocalOSS(path string) *localoss {
	return &localoss{
		path: path,
		lm:   map[string]struct{}{},
	}
}

func (m *localoss) Put(ctx context.Context, ext string, data []byte) (string, error) {
	return m.PutReader(ctx, ext, bytes.NewReader(data), int64(len(data)))
}

func (m *localoss) PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error) {
	name := sha(strconv.Itoa(int(time.Now().UnixNano()))) + "." + ext
	if _, ok := m.lm[name[:4]]; !ok {
		if err := os.MkdirAll(filepath.Join(m.path, name[:2], name[2:4]), 0700); err != nil {
			return "", err
		}
		m.lm[name[:4]] = struct{}{}
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

func (m *localoss) Get(ctx context.Context, name string) ([]byte, error) {
	r, err := m.GetReader(ctx, name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (m *localoss) GetReader(ctx context.Context, name string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(m.path, name))
}

func (m *localoss) Del(ctx context.Context, name string) error {
	return os.Remove(filepath.Join(m.path, name))
}
