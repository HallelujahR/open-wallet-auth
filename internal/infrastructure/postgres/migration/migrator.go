package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrator executes versioned SQL migration files against PostgreSQL.
// Migrator 负责按版本执行 PostgreSQL SQL 迁移文件。
type Migrator struct {
	db  *sql.DB
	dir string
}

// NewMigrator creates a SQL migrator using a database handle and migrations directory.
// NewMigrator 使用数据库连接和迁移目录创建 SQL 迁移器。
func NewMigrator(db *sql.DB, dir string) *Migrator {
	return &Migrator{db: db, dir: dir}
}

// Up applies all pending up migrations in version order.
// Up 按版本顺序执行所有尚未应用的 up 迁移。
func (m *Migrator) Up(ctx context.Context) error {
	files, err := migrationFiles(m.dir, ".up.sql")
	if err != nil {
		return err
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return err
	}
	for _, file := range files {
		version := versionFromFile(file)
		if applied[version] {
			continue
		}
		if err := m.applyUp(ctx, version, file); err != nil {
			return err
		}
	}
	return nil
}

// Down rolls back the latest applied migration.
// Down 回滚最近一次已应用的迁移。
func (m *Migrator) Down(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	version, err := m.latestVersion(ctx)
	if err != nil {
		return err
	}
	if version == "" {
		return nil
	}
	file := filepath.Join(m.dir, version+".down.sql")
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("find down migration %s: %w", file, err)
	}
	return m.applyDown(ctx, version, file)
}

// ensureTable creates the migration bookkeeping table when missing.
// ensureTable 在缺失时创建迁移版本记录表。
func (m *Migrator) ensureTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(32) PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
	return err
}

// appliedVersions returns the set of already applied migration versions.
// appliedVersions 返回已经应用过的迁移版本集合。
func (m *Migrator) appliedVersions(ctx context.Context) (map[string]bool, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = true
	}
	return versions, rows.Err()
}

// latestVersion returns the latest applied migration version.
// latestVersion 返回最近一次应用的迁移版本。
func (m *Migrator) latestVersion(ctx context.Context) (string, error) {
	var version string
	err := m.db.QueryRowContext(ctx, "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return version, nil
}

// applyUp runs one up migration and records its version atomically.
// applyUp 在事务中执行单个 up 迁移并记录版本。
func (m *Migrator) applyUp(ctx context.Context, version string, file string) error {
	sqlText, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return m.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, string(sqlText)); err != nil {
			return fmt.Errorf("apply up migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", version); err != nil {
			return err
		}
		return nil
	})
}

// applyDown runs one down migration and removes its version atomically.
// applyDown 在事务中执行单个 down 迁移并删除版本记录。
func (m *Migrator) applyDown(ctx context.Context, version string, file string) error {
	sqlText, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return m.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, string(sqlText)); err != nil {
			return fmt.Errorf("apply down migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
			return err
		}
		return nil
	})
}

// withTx executes a migration operation in one transaction.
// withTx 在单个事务中执行迁移操作。
func (m *Migrator) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// migrationFiles returns sorted migration files with the requested suffix.
// migrationFiles 返回指定后缀的有序迁移文件列表。
func migrationFiles(dir string, suffix string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*"+suffix))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

// versionFromFile extracts a migration version from a file name.
// versionFromFile 从迁移文件名中提取版本号。
func versionFromFile(file string) string {
	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".up.sql")
	name = strings.TrimSuffix(name, ".down.sql")
	return name
}
