package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// any interpolated values are stored in the request context
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	// if param could not be converted or is less than 1 we know there is an
	// issue and return 404
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "show the details of the movie %d", id)
}
