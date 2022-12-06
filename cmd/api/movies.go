package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jim-at-jibba/greenlight/internal/data"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// this struct is our target decode destination
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestHandler(w, r, err)
		return
	}

	// dump contents
	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	// if param could not be converted or is less than 1 we know there is an
	// issue and return 404
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:       id,
		CreateAt: time.Now(),
		Title:    "Casablanca",
		Runtime:  102,
		Genres:   []string{"drame", "romance", "war"},
		Version:  1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
