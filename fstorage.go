package fstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FOss interface {
	// Put upload file
	PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error)
	// Get get file
	GetReader(ctx context.Context, name string) (io.ReadCloser, error)
	// Del delete file
	Del(ctx context.Context, name string) error
}

type File struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

type FDB interface {
	// Create file record
	Create(ctx context.Context, file File) error
	// Get file
	Get(ctx context.Context, ids ...string) ([]File, error)
	// Del delete file
	Del(ctx context.Context, id string) error
	// Check check file exits
	Check(ctx context.Context, ids ...string) (bool, error)
	// Use use file
	Use(ctx context.Context, who string, ids ...string) error
	// UseCheck use file check
	UseCheck(ctx context.Context, id string) (bool, error)
	// UnUse unuse file
	UnUse(ctx context.Context, who string, ids ...string) error
	// Clean clean file with not use file
	Clean(ctx context.Context, t time.Time, cb func(path string) error) error
}

type fstorageCtxKey string

var _fstorageCtxKey fstorageCtxKey = "@nzlov@fstorage"

func For(ctx context.Context) *FStorage {
	return ctx.Value(_fstorageCtxKey).(*FStorage)
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

func (f *FStorage) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, _fstorageCtxKey, f)
}

// Put upload file. return id
func (f *FStorage) Put(ctx context.Context, name, ext string, data []byte) (*File, error) {
	return f.PutReader(ctx, name, ext, bytes.NewBuffer(data), int64(len(data)))
}

// PutReader upload file by reader. return id
func (f *FStorage) PutReader(ctx context.Context, name, ext string, r io.Reader, l int64) (*File, error) {
	path, err := f.oss.PutReader(ctx, ext, r, l)
	if err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	ids := strings.ReplaceAll(id.String(), "-", "")
	file := File{
		ID:       ids,
		Filename: name,
		Path:     path,
	}
	return &file, f.db.Create(ctx, file)
}

// Get get file by id
func (f *FStorage) Get(ctx context.Context, id string) ([]byte, error) {
	r, err := f.GetReader(ctx, id)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// GetFile get file by id
func (f *FStorage) GetFile(ctx context.Context, ids ...string) ([]File, error) {
	return f.db.Get(ctx, ids...)
}

// GetReader get file reader by id
func (f *FStorage) GetReader(ctx context.Context, id string) (io.ReadCloser, error) {
	return f.oss.GetReader(ctx, id)
}

// Check check file by id list
func (f *FStorage) Check(ctx context.Context, ids ...string) (bool, error) {
	return f.db.Check(ctx, ids...)
}

// Use use file by id list
func (f *FStorage) Use(ctx context.Context, who string, ids ...string) error {
	return f.db.Use(ctx, who, ids...)
}

// UnUse unuse file by id list
func (f *FStorage) UnUse(ctx context.Context, who string, ids ...string) error {
	return f.db.UnUse(ctx, who, ids...)
}

// Del delete file by id list
func (f *FStorage) Del(ctx context.Context, ids ...string) error {
	for _, v := range ids {
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

// Clean clean unuse file with last update time
func (f *FStorage) Clean(ctx context.Context, befor time.Time) error {
	return f.db.Clean(ctx, befor, func(path string) error {
		return f.oss.Del(ctx, path)
	})
}

func (f *FStorage) FDB(db FDB) *FStorage {
	return &FStorage{
		oss: f.oss,
		db:  db,
	}
}
