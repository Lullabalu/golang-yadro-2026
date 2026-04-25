package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
    INSERT INTO comics (id, img_url)
    VALUES ($1, $2)
    ON CONFLICT (id) DO NOTHING
`, comics.ID, comics.URL)

	if err != nil {
		return err
	}

	words := comics.Words

	for _, word := range words {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO words(word)
		VALUES ($1)
		ON CONFLICT (word) DO NOTHING
		`, word)

		if err != nil {
			return err
		}

		var id int
		err = tx.QueryRowContext(ctx, `
			SELECT w.id 
			FROM words w
			WHERE w.word = $1
		`, word).Scan(&id)

		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
		INSERT INTO comic_words(comic_id, word_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
		`, comics.ID, id)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var WordsTotal int
	var WordsUnique int
	var ComicsFetched int

	err := db.conn.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM comics
	`).Scan(&ComicsFetched)

	if err != nil {
		return core.DBStats{}, err
	}

	err = db.conn.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM words
	`).Scan(&WordsUnique)

	if err != nil {
		return core.DBStats{}, err
	}

	err = db.conn.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM comic_words
	`).Scan(&WordsTotal)

	if err != nil {
		return core.DBStats{}, err
	}

	return core.DBStats{WordsTotal: WordsTotal, ComicsFetched: ComicsFetched, WordsUnique: WordsUnique}, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT id 
		FROM comics
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ids := make([]int, 0)

	for rows.Next() {
		var id int
		err := rows.Scan(&id)

		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	query := `
        TRUNCATE TABLE comic_words, words, comics RESTART IDENTITY;
    `
	_, err := db.conn.ExecContext(ctx, query)
	return err
}
