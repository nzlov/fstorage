package ossminio

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
)

type minioss struct {
	c      *minio.Client
	prefix string
	bucket string
}

func NewMiniOSS(c *minio.Client, bucket, prefix string) *minioss {
	return &minioss{
		c:      c,
		bucket: bucket,
		prefix: prefix,
	}
}

func (m *minioss) PutReader(ctx context.Context, ext string, r io.Reader, l int64) (string, error) {
	t := time.Now()
	n := filepath.Join(m.prefix, t.Format("2006/01/02"), sha(strconv.Itoa(int(t.UnixNano())))+"."+ext)
	_, err := m.c.PutObject(ctx, m.bucket, n, r, l, minio.PutObjectOptions{})
	return n, err
}

func (m *minioss) GetReader(ctx context.Context, name string) (io.ReadCloser, error) {
	return m.c.GetObject(ctx, m.bucket, name, minio.GetObjectOptions{})
}

func (m *minioss) Del(ctx context.Context, name string) error {
	return m.c.RemoveObject(ctx, m.bucket, name, minio.RemoveObjectOptions{})
}

func sha(s string) string {
	m := sha1.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}
