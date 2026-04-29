package db

import (
	"context"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/search/core"
)

type DB struct {
	conn *sqlx.DB
}

func New(address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)

	if err != nil {
		log.Fatalf("Errors with connection to DB")
		return nil, err
	}

	return &DB{
		conn: db,
	}, nil
}

func (db *DB) Search(ctx context.Context, words []string, limit int) ([]core.ComicsInfo, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT c.id, c.img_url
		FROM comics c
		JOIN comic_words cw ON cw.comic_id = c.id
		JOIN words w ON w.id = cw.word_id
		WHERE w.word = ANY($1)
		GROUP BY c.id, c.img_url
		ORDER BY COUNT(DISTINCT w.word) DESC, c.id
		LIMIT $2
	`, words, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comics := make([]core.ComicsInfo, 0)

	for rows.Next() {
		var c core.ComicsInfo
		if err := rows.Scan(&c.ID, &c.URL); err != nil {
			return nil, err
		}
		comics = append(comics, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comics, nil
}
