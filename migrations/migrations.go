package migrations

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

//go:embed *.sql
var embedMigrations embed.FS

func Apply(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}
