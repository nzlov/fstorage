package fstorage

import (
	"context"
	"time"
)

type testdb struct{}

func (testdb) Create(ctx context.Context, file File) error                            { return nil }
func (testdb) Get(ctx context.Context, id ...string) ([]File, error)                  { return nil, nil }
func (testdb) Del(ctx context.Context, id string) error                               { return nil }
func (testdb) Check(ctx context.Context, ids ...string) (bool, error)                 { return true, nil }
func (testdb) Use(ctx context.Context, who string, ids ...string) error               { return nil }
func (testdb) UseCheck(ctx context.Context, id string) (bool, error)                  { return false, nil }
func (testdb) UnUse(ctx context.Context, who string, ids ...string) error             { return nil }
func (testdb) Clean(ctx context.Context, t time.Time, cb func(id string) error) error { return nil }
