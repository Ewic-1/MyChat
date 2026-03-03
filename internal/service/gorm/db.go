package gorm

import (
	"database/sql"
	"fmt"
	"mychat_server/internal/config"
	"mychat_server/internal/dao"
	"mychat_server/internal/model"
	"strings"

	"gorm.io/driver/mysql"
	gormlib "gorm.io/gorm"
)

func InitDB() error {
	cfg := config.GetConfig()
	if err := ensureDatabase(cfg); err != nil {
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MysqlConfig.User,
		cfg.MysqlConfig.Password,
		cfg.MysqlConfig.Host,
		cfg.MysqlConfig.Port,
		cfg.MysqlConfig.DatabaseName,
	)

	db, err := gormlib.Open(mysql.Open(dsn), &gormlib.Config{})
	if err != nil {
		return fmt.Errorf("connect mysql failed: %w", err)
	}

	dao.SetDB(db)
	if err := autoMigrate(db); err != nil {
		return err
	}
	return nil
}

func ensureDatabase(cfg *config.Config) error {
	dbName := strings.ReplaceAll(cfg.MysqlConfig.DatabaseName, "`", "")
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MysqlConfig.User,
		cfg.MysqlConfig.Password,
		cfg.MysqlConfig.Host,
		cfg.MysqlConfig.Port,
	)

	rawDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("open mysql admin connection failed: %w", err)
	}
	defer rawDB.Close()

	if _, err = rawDB.Exec("CREATE DATABASE IF NOT EXISTS `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"); err != nil {
		return fmt.Errorf("create database failed: %w", err)
	}

	return nil
}

func autoMigrate(db *gormlib.DB) error {
	if err := db.AutoMigrate(
		&model.UserInfo{},
		&model.UserContact{},
		&model.Session{},
		&model.Message{},
		&model.GroupInfo{},
		&model.ContactApply{},
	); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}
	return nil
}
