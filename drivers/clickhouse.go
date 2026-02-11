package drivers

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	gomigrator "github.com/nfwGytautas/go-migrator"
)

// clickhouseDriver is a driver for the migrator using the ClickHouse native interface
type clickhouseDriver struct {
	dsn string
	db  clickhouse.Conn
}

// NewClickhouseDriver creates a new clickhouse driver
func NewClickhouseDriver(connString string) *clickhouseDriver {
	return &clickhouseDriver{dsn: connString}
}

func (d *clickhouseDriver) Connect(ctx context.Context) (err error) {
	opts, err := clickhouse.ParseDSN(d.dsn)
	if err != nil {
		return fmt.Errorf("failed to parse clickhouse DSN: %w", err)
	}

	d.db, err = clickhouse.Open(opts)
	if err != nil {
		return err
	}

	return d.db.Ping(ctx)
}

func (d *clickhouseDriver) CreateMigrationsTable(ctx context.Context) (err error) {
	const tableSchema = `
	CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (
		id        Int32,
		name      String,
		applied_at DateTime
	) ENGINE = MergeTree()
	ORDER BY id
	`

	return d.db.Exec(ctx, tableSchema)
}

func (d *clickhouseDriver) GetCurrentVersion(ctx context.Context) (version int, err error) {
	const query = `SELECT COALESCE(max(id), 0) FROM ` + migrationsTable

	row := d.db.QueryRow(ctx, query)
	err = row.Scan(&version)
	return
}

func (d *clickhouseDriver) ApplyMigration(ctx context.Context, migration gomigrator.Migration) error {
	const query = `INSERT INTO ` + migrationsTable + ` (id, name, applied_at) VALUES (?, ?, ?)`

	// Apply the migration (no transaction support in ClickHouse)
	err := d.db.Exec(ctx, migration.MigrationSQL)
	if err != nil {
		return fmt.Errorf("failed to apply migration (%s): %w", migration.Name, err)
	}

	// Log the migration
	err = d.db.Exec(
		ctx,
		query,
		migration.Version,
		migration.Name,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to log migration: %w", err)
	}

	return nil
}

func (d *clickhouseDriver) Close(_ context.Context) error {
	return d.db.Close()
}
