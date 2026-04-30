package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/config"
	"github.com/open-wallet-auth/open-wallet-auth/internal/infrastructure/postgres/migration"
)

// main loads configuration and runs the requested database migration command.
// main 读取配置并执行指定的数据库迁移命令。
func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	dir := flag.String("dir", "./migrations", "directory containing migration SQL files")
	timeout := flag.Duration("timeout", 30*time.Second, "migration timeout")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fatalf("load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fatalf("ping database: %v", err)
	}

	migrator := migration.NewMigrator(db, *dir)
	switch *direction {
	case "up":
		err = migrator.Up(ctx)
	case "down":
		err = migrator.Down(ctx)
	default:
		err = fmt.Errorf("unsupported migration direction %q", *direction)
	}
	if err != nil {
		fatalf("run migration: %v", err)
	}
}

// fatalf writes an error to stderr and exits with failure status.
// fatalf 输出错误到 stderr 并以失败状态退出。
func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
