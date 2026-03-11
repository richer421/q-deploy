package testutil

import (
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMockDB 创建一个 mock 的 GORM DB 和 sqlmock 实例。
// 调用者通过 sqlmock 设置期望，通过 *gorm.DB 执行查询。
func NewMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}
