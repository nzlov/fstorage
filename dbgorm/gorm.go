package dbgorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type GormFile struct {
	gorm.Model

	Name string `gorm:"index"`
}

type GormFileUse struct {
	gorm.Model

	Name string `gorm:"index"`
	Who  string `gorm:"index"`
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

func (g *gormdb) Create(ctx context.Context, name string) error {
	return g.db.Create(&GormFile{Name: name}).Error
}

func (g *gormdb) Del(ctx context.Context, name string) error {
	if err := g.db.Unscoped().Where("name = ?", name).Delete(new(GormFileUse)).Error; err != nil {
		return err
	}

	return g.db.Unscoped().Where("name = ?", name).Delete(new(GormFile)).Error
}

func (g *gormdb) Check(ctx context.Context, names ...string) (bool, error) {
	n := int64(0)
	if err := g.db.Model(new(GormFile)).Where("name in (?)", names).Count(&n).Error; err != nil {
		return false, err
	}
	return int(n) == len(names), nil
}

func (g *gormdb) Use(ctx context.Context, who string, names ...string) error {
	for _, v := range names {
		if err := g.db.Create(&GormFileUse{
			Name: v,
			Who:  who,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (g *gormdb) UseCheck(ctx context.Context, name string) (bool, error) {
	n := int64(0)
	if err := g.db.Model(new(GormFileUse)).Where("name = ?", name).Count(&n).Error; err != nil {
		return false, err
	}
	return int(n) > 0, nil
}

func (g *gormdb) UnUse(ctx context.Context, who string, names ...string) error {
	ss := []string{}
	vs := []any{}

	if len(names) != 0 {
		ss = append(ss, "name in (?)")
		vs = append(vs, names)
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

func (g *gormdb) Clean(ctx context.Context, t time.Time, cb func(name string) error) error {
	for {
		fs := []GormFile{}
		if err := g.db.Where(`name not in (select name from gorm_file_uses) and created_at < ?`, t).Find(&fs).Error; err != nil {
			return err
		}
		if len(fs) == 0 {
			break
		}
		for _, v := range fs {
			if err := cb(v.Name); err != nil {
				return err
			}
			if err := g.db.Delete(&v).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
