package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type Config struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		Host: env("MYSQL_HOST", "127.0.0.1"),
		Port: env("MYSQL_PORT", "3306"),
		User: env("MYSQL_USER", "root"),
		Pass: env("MYSQL_PASS", "root"),
		Name: env("MYSQL_DB", "go_products"),
	}
}

func (c Config) DSN(denganDB bool) string {
	nama := ""
	if denganDB {
		nama = c.Name
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Local",
		c.User, c.Pass, c.Host, c.Port, nama)
}

func ConnectDB() error {
	cfg := Load()

	bootstrap, err := sql.Open("mysql", cfg.DSN(false))
	if err != nil {
		return err
	}
	defer bootstrap.Close()

	if err := bootstrap.Ping(); err != nil {
		return fmt.Errorf("gagal connect MySQL di %s:%s sebagai '%s' (%w), set password dengan: set MYSQL_PASS=passwordmu", cfg.Host, cfg.Port, cfg.User, err)
	}

	stmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Name)
	if _, err := bootstrap.Exec(stmt); err != nil {
		return fmt.Errorf("gagal membuat database '%s': %w", cfg.Name, err)
	}

	db, err := sql.Open("mysql", cfg.DSN(true))
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return err
	}

	DB = db
	return nil
}
