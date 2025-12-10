package main

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nzlov/fstorage"
	"github.com/nzlov/fstorage/dbgorm"
	"github.com/nzlov/fstorage/ossminio"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	endpoint := ""
	accessKeyID := ""
	secretAccessKey := ""
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		panic(err)
	}
	oss := ossminio.NewMiniOSS(minioClient, "demo", "debug")

	gdb, err := gorm.Open(sqlite.Open("gorm.db"))
	if err != nil {
		panic(err)
	}
	db, err := dbgorm.NewGormWithInit(gdb)
	if err != nil {
		panic(err)
	}

	fs := fstorage.NewFstorage(oss, db)
	ctx := context.Background()
	// name := "debug/2022/02/10/575d07359f559a49507021618d533ac52307320b.txt"
	name, err := fs.Put(ctx, "txt", []byte("asd"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Put:", name)
	//data, err := fs.Get(context.Background(), name)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Get:", name, ":", string(data))
	//fmt.Println("Use:", fs.Use(ctx, "test", name))
	// fmt.Println("Del:", fs.Del(ctx, name))
	fmt.Println("Clean:", fs.Clean(ctx, time.Now()))
}
