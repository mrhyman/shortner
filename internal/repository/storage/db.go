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
	DB *sqlx.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	DB, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	DB.SetMaxOpenConns(10)
	DB.SetConnMaxLifetime(time.Hour)

	if err := DB.Ping(); err != nil {
		return nil, err
	}

	return &DBStorage{DB}, nil
}

func (ds *DBStorage) Store(ctx context.Context, l model.Link) error {
	if _, err := ds.DB.ExecContext(ctx,
		"INSERT INTO links (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
		l.UUID, l.ShortURL, l.OriginalURL, l.UserID,
	); err != nil {
		return ds.convertPgError(ctx, l, err)
	}
	return nil
}

func (ds *DBStorage) StoreBatch(ctx context.Context, ls []model.Link) error {
	tx, err := ds.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO links (uuid, short_url, original_url, correlation_id, user_id)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, link := range ls {
		_, err := stmt.ExecContext(ctx, link.UUID, link.ShortURL, link.OriginalURL, link.CorrelationID, link.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (ds *DBStorage) GetByShortURL(ctx context.Context, shortURL string) (*model.Link, error) {
	var link model.Link

	err := ds.DB.Get(
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

	err := ds.DB.Get(
		&link,
		`SELECT * FROM links WHERE original_url = $1`,
		originURL,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (ds *DBStorage) GetByUserID(ctx context.Context, userID string) ([]model.Link, error) {
	var links []model.Link

	err := ds.DB.Select(
		&links,
		`SELECT * FROM links WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (ds *DBStorage) DeleteUserLinksByID(ctx context.Context, links []string) error {
	userID, ok := ctx.Value(model.UserIDKey).(string)
	if !ok {
		return model.ErrUnknownUser
	}

	tx, err := ds.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE links SET is_deleted=true 
		WHERE short_url = ANY($1)
		AND user_id = $2
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, pq.Array(links), userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ds *DBStorage) CountURLs(ctx context.Context) (int, error) {
	var count int

	err := ds.DB.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM links WHERE is_deleted = false`,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (ds *DBStorage) CountUsers(ctx context.Context) (int, error) {
	var count int

	err := ds.DB.GetContext(
		ctx,
		&count,
		`SELECT COUNT(DISTINCT user_id) FROM links WHERE is_deleted = false`,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (ds *DBStorage) Ping() error {
	return ds.DB.Ping()
}

func (ds *DBStorage) Close() error {
	return ds.DB.Close()
}

func (ds *DBStorage) MigrateUp(migrationsDir, dsn string) error {
	driver, err := postgres.WithInstance(ds.DB.DB, &postgres.Config{})
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
