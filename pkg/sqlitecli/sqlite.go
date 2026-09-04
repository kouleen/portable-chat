package sqlitecli

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/logger"
	mdl "github.com/kouleen/portable-chat/pkg/model"
	"github.com/kouleen/portable-chat/scripts"
	"gorm.io/gorm"
)

var sqliteDB *gorm.DB

func init() {
	gormLog := logger.NewZapLogger()
	dbPath := resolveProjectPath("portable-chat.db")
	// 连接 SQLite（文件不存在会自动创建）
	db, err := gorm.Open(sqlite.Open(dbPath+"?_loc=Local&parseTime=true&_journal_mode=WAL&_cache_size=-20000"), &gorm.Config{
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

	if err = sqliteDB.AutoMigrate(&mdl.IdAlloc{}, &model.Authorization{}, &model.ChatMessage{}); err != nil {
		log.Fatal("Init sqlite table ❌, %w", err)
	}
	initSQL := getInitSQL()
	if err = sqliteDB.Exec(initSQL).Error; err != nil {
		log.Fatalf("Run init sql failed ❌: %v", err)
	}
	log.Println("Connected to sqlite database ✅")
}

func resolveProjectPath(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return name
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, name)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join(wd, name)
		}
		dir = parent
	}
}

func getInitSQL() string {
	file, err := scripts.GetFs().ReadFile("init.sql")
	if err != nil {
		log.Fatal(err)
	}
	return string(file)
}

// execMultiSQL 按分号分割，批量执行多条SQL，忽略空语句
func execMultiSQL(db *gorm.DB, sql string) error {
	parts := strings.Split(sql, ";")
	for _, stmt := range parts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}
