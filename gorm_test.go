package fstorage

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGorm(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		t.Fatal("gorm open error:", err)
	}
	fdb, err := NewGormWithInit(db)
	if err != nil {
		t.Fatal("gorm init error:", err)
	}
	name := "a"
	who := "who"
	if err := fdb.Create(context.Background(), name); err != nil {
		t.Fatal("gorm create error:", err)
	}
	{
		has, err := fdb.Check(context.Background(), name)
		if err != nil {
			t.Fatal("gorm check error:", err)
		}
		if !has {
			t.Fatal("gorm create error: no match name")
		}
	}
	{
		has, err := fdb.Check(context.Background(), "no")
		if err != nil {
			t.Fatal("gorm check error:", err)
		}
		if has {
			t.Fatal("gorm check error: find don't exist")
		}
	}
	{
		if err := fdb.Use(context.Background(), who, name); err != nil {
			t.Fatal("gorm use error:", err)
		}
		gfs := GormFileUse{}
		if err := db.Where("who = ?", who).First(&gfs).Error; err != nil {
			t.Fatal("gorm find error:", err)
		}
		if gfs.Who != who || gfs.Name != name {
			t.Fatal("gorm use error: who or name no match")
		}
	}
	{
		if err := fdb.UnUse(context.Background(), who); err != nil {
			t.Fatal("gorm unuse error:", err)
		}
		gfs := GormFileUse{}
		if err := db.Where("who = ?", who).First(&gfs).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				t.Fatal("gorm find error:", err)
			}
		} else {
			t.Fatal("gorm unuse no delete")
		}
	}
	{
		if err := fdb.Del(context.Background(), name); err != nil {
			t.Fatal("gorm del error:", err)
		}
		n := int64(0)
		if err := db.Model(new(GormFile)).Count(&n).Error; err != nil {
			t.Fatal("gorm del count query error:", err)
		}
		if n != 0 {
			t.Fatal("gorm del count check error")
		}
		if err := db.Model(new(GormFileUse)).Count(&n).Error; err != nil {
			t.Fatal("gorm del count query error:", err)
		}
		if n != 0 {
			t.Fatal("gorm del count check error")
		}
	}

	{
		if err := fdb.Create(context.Background(), name); err != nil {
			t.Fatal("gorm create error:", err)
		}
		if err := fdb.Create(context.Background(), name+"2"); err != nil {
			t.Fatal("gorm create error:", err)
		}
		if err := fdb.Create(context.Background(), name+"3"); err != nil {
			t.Fatal("gorm create error:", err)
		}
		if err := fdb.Use(context.Background(), who, name); err != nil {
			t.Fatal("gorm use error:", err)
		}
		if err := fdb.Clean(context.Background(), time.Now().AddDate(0, 0, 1), func(name string) error {
			t.Log("clean:", name)
			return nil
		}); err != nil {
			t.Fatal("gorm clean error:", err)
		}
		gfs := []GormFile{}
		if err := db.Find(&gfs).Error; err != nil {
			t.Fatal("gorm find query error:", err)
		}
		if len(gfs) != 1 {
			t.Fatal("gorm clean file error: no clean")
		}
		if gfs[0].Name != name {
			t.Fatal("gorm clean file error: clean error data")
		}
	}
}
