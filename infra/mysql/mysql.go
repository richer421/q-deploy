package mysql

import (
	"fmt"

	"github.com/richer421/q-deploy/conf"
	"github.com/richer421/q-deploy/infra/mysql/dao"
	"github.com/richer421/q-deploy/infra/mysql/model"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg conf.MySQLConfig) error {
	// 先连接到 mysql 服务器（不指定数据库），创建数据库（如果不存在）
	createDB, err := gorm.Open(mysql.Open(cfg.DSNWithoutDB()), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlCreateDB, err := createDB.DB()
	if err != nil {
		return err
	}
	defer sqlCreateDB.Close()

	if err := createDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Database)).Error; err != nil {
		return err
	}

	// 连接到指定数据库
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	if err := DB.Use(otelgorm.NewPlugin()); err != nil {
		return err
	}

	dao.SetDefault(DB)
	return nil
}

func Migrate() error {
	if DB == nil {
		return gorm.ErrInvalidDB
	}
	return DB.AutoMigrate(
		&model.Release{},
	)
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
