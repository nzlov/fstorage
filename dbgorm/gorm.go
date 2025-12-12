package dbgorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nzlov/fstorage"
	"gorm.io/gorm"
)

type GormFile struct {
	gorm.Model

	UID  string `gorm:"index"`
	Name string
	Path string
}

type GormFileUse struct {
	gorm.Model

	UID string `gorm:"index"`
	Who string `gorm:"index"`
}

type gormdb struct {
	db *gorm.DB
}

func NewGormWithInit(db *gorm.DB) (*gormdb, error) {
	if err := db.AutoMigrate(new(GormFile), new(GormFileUse)); err != nil {
		return nil, err
	}

	return &gormdb{db}, nil
}

func NewGorm(db *gorm.DB) *gormdb {
	return &gormdb{db}
}

func (g *gormdb) Create(ctx context.Context, file fstorage.File) error {
	return g.db.Create(&GormFile{
		UID:  file.ID,
		Name: file.Filename,
		Path: file.Path,
	}).Error
}

func (g *gormdb) Get(ctx context.Context, ids ...string) ([]fstorage.File, error) {
	fs := []GormFile{}
	if err := g.db.Where("uid in (?)", ids).Find(&fs).Error; err != nil {
		return nil, err
	}

	files := make([]fstorage.File, len(fs))
	for i, v := range fs {
		files[i] = fstorage.File{
			ID:       v.UID,
			Filename: v.Name,
			Path:     v.Path,
		}
	}
	return files, nil
}

func (g *gormdb) Del(ctx context.Context, id string) error {
	if err := g.db.Unscoped().Where("uid = ?", id).Delete(new(GormFileUse)).Error; err != nil {
		return err
	}

	return g.db.Unscoped().Where("uid = ?", id).Delete(new(GormFile)).Error
}

func (g *gormdb) Check(ctx context.Context, ids ...string) (bool, error) {
	n := int64(0)
	if err := g.db.Model(new(GormFile)).Where("uid in (?)", ids).Count(&n).Error; err != nil {
		return false, err
	}
	return int(n) == len(ids), nil
}

func (g *gormdb) Use(ctx context.Context, who string, ids ...string) error {
	for _, v := range ids {
		if err := g.db.Create(&GormFileUse{
			UID: v,
			Who: who,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (g *gormdb) UseCheck(ctx context.Context, id string) (bool, error) {
	n := int64(0)
	if err := g.db.Model(new(GormFileUse)).Where("uid = ?", id).Count(&n).Error; err != nil {
		return false, err
	}
	return int(n) > 0, nil
}

func (g *gormdb) UnUse(ctx context.Context, who string, ids ...string) error {
	ss := []string{}
	vs := []any{}

	if len(ids) != 0 {
		ss = append(ss, "uid in (?)")
		vs = append(vs, ids)
	}
	if who != "" {
		ss = append(ss, "who = ?")
		vs = append(vs, who)
	}
	if len(ss) == 0 {
		return fmt.Errorf("fstorage unuse no params")
	}
	return g.db.Where(strings.Join(ss, " and "), vs...).Unscoped().Delete([]GormFileUse{}).Error
}

func (g *gormdb) Clean(ctx context.Context, t time.Time, cb func(path string) error) error {
	for {
		fs := []GormFile{}
		if err := g.db.Where(`uid not in (select uid from gorm_file_uses) and created_at < ?`, t).Find(&fs).Error; err != nil {
			return err
		}
		if len(fs) == 0 {
			break
		}
		for _, v := range fs {
			if err := cb(v.Path); err != nil {
				return err
			}
			if err := g.db.Delete(&v).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
