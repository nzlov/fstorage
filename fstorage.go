package fstorage

import (
	"context"
	"io"
	"time"
)

type FOss interface {
	Put(ctx context.Context, ext string, data []byte) (string, error)
	PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error)
	Get(ctx context.Context, name string) ([]byte, error)
	GetReader(ctx context.Context, name string) (io.Reader, error)
	Del(ctx context.Context, name string) error
}

type FDB interface {
	Create(ctx context.Context, name string) error
	Del(ctx context.Context, name string) error
	Check(ctx context.Context, names ...string) (bool, error)
	Use(ctx context.Context, who string, names ...string) error
	UnUse(ctx context.Context, who string, names ...string) error
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

func (f *FStorage) Put(ctx context.Context, ext string, data []byte) (string, error) {
	name, err := f.oss.Put(ctx, ext, data)
	if err != nil {
		return "", err
	}
	return name, f.db.Create(ctx, name)
}

func (f *FStorage) PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error) {
	name, err := f.oss.PutReader(ctx, ext, r, l)
	if err != nil {
		return "", err
	}
	return name, f.db.Create(ctx, name)
}

func (f *FStorage) Get(ctx context.Context, name string) ([]byte, error) {
	return f.oss.Get(ctx, name)
}

func (f *FStorage) GetReader(ctx context.Context, name string) (io.Reader, error) {
	return f.oss.GetReader(ctx, name)
}

func (f *FStorage) Check(ctx context.Context, names ...string) (bool, error) {
	return f.db.Check(ctx, names...)
}

func (f *FStorage) Use(ctx context.Context, who string, names ...string) error {
	return f.db.Use(ctx, who, names...)
}

func (f *FStorage) UnUse(ctx context.Context, who string, names ...string) error {
	return f.db.UnUse(ctx, who, names...)
}

func (f *FStorage) Del(ctx context.Context, names ...string) error {
	for _, v := range names {
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
