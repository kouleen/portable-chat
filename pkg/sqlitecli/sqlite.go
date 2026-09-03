package sqlitecli

import (
	"log"

	"github.com/glebarez/sqlite"
	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/logger"
	"gorm.io/gorm"
)

var sqliteDB *gorm.DB

func init() {
	gormLog := logger.NewZapLogger()
	// 连接 SQLite（文件不存在会自动创建）
	db, err := gorm.Open(sqlite.Open("portable-chat.db?_loc=Local&parseTime=true&_journal_mode=WAL&_cache_size=-20000"), &gorm.Config{
		Logger:      gormLog,
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatal("Connected to sqlite database ❌")
	}
	pingdb, err := db.DB()
	if err != nil {
		log.Fatal("Connected to sqlite database ❌")
	}
	// SQLite 连接池配置（轻量即可）
	pingdb.SetMaxOpenConns(1)    // 最大打开连接
	pingdb.SetMaxIdleConns(1)    // 最大空闲连接
	pingdb.SetConnMaxLifetime(0) // 连接永不过期
	sqliteDB = db

	if err = sqliteDB.AutoMigrate(&model.Authorization{}, &model.CharContact{}, &model.CharHistory{}); err != nil {
		log.Fatal("Init sqlite table ❌, %w", err)
	}
	log.Println("Connected to sqlite database ✅")
}
