package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Macabare/url-shortener/internal/model"
	"github.com/Macabare/url-shortener/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
		create table if not exists url(
		  id integer primary key,
			original_url text not null,
			short_code text not null unique,
			access_count integer not null default(0),
			created_at text not null default CURRENT_TIMESTAMP,
			updated_at text not null default CURRENT_TIMESTAMP
		);
		create index if not exists idx_short_code on url(short_code);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if _, err := stmt.Exec(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(ctx context.Context, u model.URL) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare(`
		insert into url(original_url, short_code)
		values(?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, u.OriginalURL, u.ShortCode)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUrlExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last inserted id(%w)", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, shortCode string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("select original_url from url where short_code=?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var resUrl string
	err = stmt.QueryRowContext(ctx, shortCode).Scan(&resUrl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrUrlNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resUrl, nil
}

func (s *Storage) IncrementAccessCount(ctx context.Context, shortCode string) error {
	const op = "storage.sqlite.IncrementAccessCount"

	_, err := s.db.ExecContext(ctx, `
		update url
		set
			access_count = access_count + 1,
			updated_at = CURRENT_TIMESTAMP
		where short_code = ?
		`, shortCode)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURLStat(ctx context.Context, shortCode string) (model.URL, error) {
	const op = "storage.sqlite.GetURLStat"

	var (
		u         model.URL
		createdAt string
		updatedAt string
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT
			id,
			original_url,
			short_code,
			access_count,
			created_at,
			updated_at
		FROM url
		WHERE short_code = ?
	`, shortCode).Scan(
		&u.ID,
		&u.OriginalURL,
		&u.ShortCode,
		&u.AccessCount,
		&createdAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return model.URL{}, storage.ErrUrlNotFound
	}

	if err != nil {
		return model.URL{}, fmt.Errorf("%s: %w", op, err)
	}

	u.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		return model.URL{}, fmt.Errorf("%s: failed to parse created_at: %w", op, err)
	}

	u.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAt)
	if err != nil {
		return model.URL{}, fmt.Errorf("%s: failed to parse updated_at: %w", op, err)
	}

	return u, nil
}
