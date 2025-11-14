package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/mrhyman/shortner/internal/model"
)

type DBStorage struct {
	db *sqlx.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DBStorage{db}, nil
}

func (ds *DBStorage) Store(ctx context.Context, l model.Link) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO links (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		l.UUID, l.ShortURL, l.OriginalURL)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (ds *DBStorage) StoreBatch(ctx context.Context, ls []model.Link) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO links (uuid, short_url, original_url, correlation_id)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, link := range ls {
		_, err := stmt.ExecContext(ctx, link.UUID, link.ShortURL, link.OriginalURL, link.CorrelationID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (ds *DBStorage) GetByID(ctx context.Context, id string) (*model.Link, error) {
	var link model.Link

	err := ds.db.Get(
		&link,
		`SELECT * FROM links WHERE short_url LIKE '%' || $1`,
		id,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (ds *DBStorage) Ping() error {
	return ds.db.Ping()
}

func (ds *DBStorage) Close() error {
	return ds.db.Close()
}

func (ds *DBStorage) MigrateUp(migrationsDir, dsn string) error {
	driver, err := postgres.WithInstance(ds.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
