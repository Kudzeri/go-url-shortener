package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: Add migrations
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
	    id INTEGER PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// TODO: implement Alias validation
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	//stmt, err := s.db.Prepare("INSERT INTO url(url,alias) VALUES(?, ?)")
	//if err != nil {
	//	return 0, fmt.Errorf("%s: %w", op, err)
	//}
	//defer stmt.Close()

	res, err := s.db.Exec(
		"INSERT INTO url(url, alias) VALUES(?, ?)",
		urlToSave, alias,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			err = storage.ErrAlreadyExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"
	//stmt, err := s.db.QueryRow("SELECT alias, url FROM url WHERE alias = ?").Scan(&url)
	//if err != nil {
	//	return "", fmt.Errorf("%s: %w", op, err)
	//})
	//defer stmt.Close()

	var url string
	err := s.db.QueryRow("SELECT url FROM url WHERE alias = ?", alias).Scan(&url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteUrl(alias string) error {
	const op = "storage.sqlite.DeleteUrl"

	_, err := s.db.Exec("DELETE FROM url WHERE alias = ?", alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
