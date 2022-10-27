package fstorage

import (
	"context"
	"time"
)

type testdb struct{}

func (testdb) Create(ctx context.Context, name string) error                            { return nil }
func (testdb) Del(ctx context.Context, name string) error                               { return nil }
func (testdb) Check(ctx context.Context, names ...string) (bool, error)                 { return true, nil }
func (testdb) Use(ctx context.Context, who string, names ...string) error               { return nil }
func (testdb) UnUse(ctx context.Context, who string, names ...string) error             { return nil }
func (testdb) Clean(ctx context.Context, t time.Time, cb func(name string) error) error { return nil }
