package data

import (
	"fmt"
	"strconv"
)

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
