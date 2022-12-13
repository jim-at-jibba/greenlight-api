package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jim-at-jibba/greenlight/internal/validator"
	"github.com/lib/pq"
)

// All fields are Exported to allow them to be visible to `encoding/json`
type Movie struct {
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"-"` // hides value from json always
	Title    string    `json:"title"`
	Year     int32     `json:"year,omitempty"` // hides if field has no value
	Runtime  Runtime   `json:"runtime,omitempty"`
	Genres   []string  `json:"genres,omitempty"`
	Version  int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {

	// Use check() method to execute our validation
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	// using uique helper
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *sql.DB
}

// Takes a *Movie pointer meaning we are updating the values
// at the location the param points to with the returned values from the
// insert
func (m MovieModel) Insert(movie *Movie) error {
	query := `
  INSERT INTO movies (title, year, runtime, genres)
  VALUES ($1, $2, $3, $4)
  RETURNING id, created_at, version
  `

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreateAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	// Because id can never be negative why are we not using uint64 (unsigned).
	// 2 reasons:
	// 1. Postgres does not support unsigned ints. Its best to align your go
	// code and database types to avoid compatibility issues
	// 2. Go database package does not support integer values great than int64
	// which uint64 might be
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
  SELECT id, created_at, title, year, runtime, genres, version
  FROM movies
  WHERE id = $1
  `

	// Struct to hold returned data
	var movie Movie

	// 3 second timeout
	// context.Background - root context
	// Good article aboiut context
	// https://medium.com/@swapnildawange3650/golang-context-is-important-5dc6b6519866
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreateAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound

		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
  UPDATE movies
  SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
  WHERE id = $5 AND version = $6
  RETURNING version
  `

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
  DELETE FROM movies
  WHERE id = $1
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// (LOWER(title) = LOWER($1) OR $1 = '') = title = title or is skipped because its empty
	// @> is the postgres array contains function
	query := `
  SELECT id, created_at, title, year, runtime, genres, version
  FROM movies
  WHERE (LOWER(title) = LOWER($1) OR $1 = '')
  AND (genres @> $2 OR $2 = '{}')
  ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	movies := []*Movie{}

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.CreateAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}
