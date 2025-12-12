package fstorage

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestLocalOss(t *testing.T) {
	local := NewLocalOSS("./local")
	defer os.RemoveAll("./local")
	fs := NewFstorage(local, testdb{})

	path, err := fs.Put(context.Background(), "a.name", "txt", []byte("asd"))
	if err != nil {
		t.Fatal("local put error:", err)
	}
	data, err := os.ReadFile("./local/" + path)
	if err != nil {
		t.Fatal("local file path error:", err)
	}
	if string(data) != "asd" {
		t.Fatal("local file put data no match")
	}
	data, err = fs.Get(context.Background(), path)
	if err != nil {
		t.Fatal("local get error:", err)
	}
	if string(data) != "asd" {
		t.Fatal("local file get data no match")
	}
	if err = fs.Del(context.Background(), path); err != nil {
		t.Fatal("local del error:", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatal("local check file error:", err)
		}
	} else {
		fmt.Println(fi, err)
		t.Fatal("lcoal del failed")
	}
}
