package sqlitecli

import "gorm.io/gorm"

func GetSqliteDB() *gorm.DB {
	return sqliteDB
}
