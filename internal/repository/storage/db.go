package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	if _, err := ds.db.ExecContext(ctx,
		"INSERT INTO links (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		l.UUID, l.ShortURL, l.OriginalURL,
	); err != nil {
		return ds.convertPgError(ctx, l, err)
	}
	return nil
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

func (ds *DBStorage) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	var link model.Link

	err := ds.db.Get(
		&link,
		`SELECT * FROM links WHERE short_url LIKE '%' || $1`,
		shortURL,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (ds *DBStorage) GetByOriginalURL(ctx context.Context, originURL string) (*model.Link, error) {
	var link model.Link

	err := ds.db.Get(
		&link,
		`SELECT * FROM links WHERE original_url = $1`,
		originURL,
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

func (ds *DBStorage) GetUserLinks(ctx context.Context, userID string) ([]model.Link, error) {
	return nil, nil
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

func (ds *DBStorage) convertPgError(ctx context.Context, l model.Link, err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		if pgErr.Constraint == "idx_links_original_url" {
			existing, getErr := ds.GetByOriginalURL(ctx, l.OriginalURL)
			if getErr == nil {
				return model.NewAlreadyExistsError(existing.ShortURL, err)
			}
		}
	}

	return err
}
