package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// When Go is encoding a particular type to JSON, it looks to
// see if that type satifieds the json.Marshaler interface which
// has a MarshalJSON() method on it.
// We can satisfy this interface to do encode types exactly as we want to
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate string containing movie  runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// strconv.Quote wraps string in dobule quotes
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// Must be pointer receiver so that we modify actual value and not
// a copy
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// We expect that the incoming JSON value will be a string in the format
	// "<runtime> mins", we need to remove the double quotes
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split string to isolate part containing number
	parts := strings.Split(unquotedJSONValue, " ")

	// check parts of string is expected format
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// parse string into int32
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(i)

	return nil
}
