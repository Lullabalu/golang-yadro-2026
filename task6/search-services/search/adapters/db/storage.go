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

func (db *DB) GetIndexRows(ctx context.Context) ([]core.IndexRow, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT c.id, c.img_url, w.word
		FROM comics c
		JOIN comic_words cw ON cw.comic_id = c.id
		JOIN words w ON w.id = cw.word_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexRows := make([]core.IndexRow, 0)
	for rows.Next() {
		var indexRow core.IndexRow
		if err = rows.Scan(&indexRow.ComicID, &indexRow.URL, &indexRow.Word); err != nil {
			return nil, err
		}
		indexRows = append(indexRows, indexRow)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return indexRows, nil
}

func (db *DB) Get(ctx context.Context, id int) (core.ComicsInfo, error) {
	var comicsInfo core.ComicsInfo
	if err := db.conn.QueryRowContext(ctx, `
		SELECT c.id, c.img_url
		FROM comics c
		WHERE c.id = $1
	`, id).Scan(&comicsInfo.ID, &comicsInfo.URL); err != nil {
		return core.ComicsInfo{}, err
	}

	return comicsInfo, nil
}
