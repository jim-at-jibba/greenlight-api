package data

import "time"

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
