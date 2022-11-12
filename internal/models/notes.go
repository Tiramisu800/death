package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Note struct {
	ID       int
	Fullname string
	HowDie   string
	Created  time.Time
	Die      time.Time
}

// Define a NoteModel type which wraps a sql.DB connection pool.
type NoteModel struct {
	DB *pgxpool.Pool
}

func (m *NoteModel) Insert(name string, howdie string, die int) (int, error) {
	q := "INSERT INTO victims (fullname, howdie, created, die) VALUES($1, $2, NOW(), (NOW() + INTERVAL '1 SECONDS' * $3)) RETURNING id"

	row := m.DB.QueryRow(context.Background(),
		q,
		name, howdie, die)

	var id uint64

	err := row.Scan(&id)
	if err != nil {
		fmt.Printf("Unable to INSERT: %v\n", err)
		return 0, err
	}

	fmt.Println(id)

	return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *NoteModel) Get(id int) (*Note, error) {
	row := m.DB.QueryRow(context.Background(),
		"SELECT id, fullname, howdie, created, die FROM victims WHERE die > NOW() AND id = $1", id)
	s := &Note{}
	err := row.Scan(&s.ID, &s.Fullname, &s.HowDie, &s.Created, &s.Die)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *NoteModel) Latest() ([]*Note, error) {
	query := `SELECT id, fullname, howdie, created, die FROM victims
			WHERE die > NOW() ORDER BY id DESC LIMIT 10`
	rows, err := m.DB.Query(context.Background(), query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notes := []*Note{}

	for rows.Next() {
		s := &Note{}

		err = rows.Scan(&s.ID, &s.Fullname, &s.HowDie, &s.Created, &s.Die)

		if err != nil {
			return nil, err
		}

		notes = append(notes, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}
