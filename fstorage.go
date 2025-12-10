package fstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
)

type FOss interface {
	// Put upload file
	PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error)
	// Get get file
	GetReader(ctx context.Context, name string) (io.ReadCloser, error)
	// Del delete file
	Del(ctx context.Context, name string) error
}

type FDB interface {
	// Create file record
	Create(ctx context.Context, name string) error
	// Del delete file
	Del(ctx context.Context, name string) error
	// Check check file exits
	Check(ctx context.Context, names ...string) (bool, error)
	// Use use file
	Use(ctx context.Context, who string, names ...string) error
	// UseCheck use file check
	UseCheck(ctx context.Context, name string) (bool, error)
	// UnUse unuse file
	UnUse(ctx context.Context, who string, names ...string) error
	// Clean clean file with not use file
	Clean(ctx context.Context, t time.Time, cb func(name string) error) error
}

type FStorage struct {
	oss FOss
	db  FDB
}

func NewFstorage(oss FOss, db FDB) *FStorage {
	return &FStorage{
		oss: oss,
		db:  db,
	}
}

// Put upload file
func (f *FStorage) Put(ctx context.Context, ext string, data []byte) (string, error) {
	name, err := f.oss.PutReader(ctx, ext, bytes.NewBuffer(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	return name, f.db.Create(ctx, name)
}

// PutReader upload file by reader
func (f *FStorage) PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error) {
	name, err := f.oss.PutReader(ctx, ext, r, l)
	if err != nil {
		return "", err
	}
	return name, f.db.Create(ctx, name)
}

// Get get file
func (f *FStorage) Get(ctx context.Context, name string) ([]byte, error) {
	r, err := f.GetReader(ctx, name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// GetReader get file reader
func (f *FStorage) GetReader(ctx context.Context, name string) (io.ReadCloser, error) {
	return f.oss.GetReader(ctx, name)
}

// Check check file
func (f *FStorage) Check(ctx context.Context, names ...string) (bool, error) {
	return f.db.Check(ctx, names...)
}

// Use use file
func (f *FStorage) Use(ctx context.Context, who string, names ...string) error {
	return f.db.Use(ctx, who, names...)
}

// UnUse unuse file
func (f *FStorage) UnUse(ctx context.Context, who string, names ...string) error {
	return f.db.UnUse(ctx, who, names...)
}

// Del delete file
func (f *FStorage) Del(ctx context.Context, names ...string) error {
	for _, v := range names {
		if u, err := f.db.UseCheck(ctx, v); err != nil {
			return err
		} else if u {
			return fmt.Errorf("%w:%s", ErrFileUsing, v)
		}
		if err := f.oss.Del(ctx, v); err != nil {
			return err
		}
		if err := f.db.Del(ctx, v); err != nil {
			return err
		}
	}
	return nil
}

func (f *FStorage) Clean(ctx context.Context, befor time.Time) error {
	return f.db.Clean(ctx, befor, func(name string) error {
		return f.oss.Del(ctx, name)
	})
}

func (f *FStorage) FDB(db FDB) *FStorage {
	return &FStorage{
		oss: f.oss,
		db:  db,
	}
}
