package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"
)

// JSONBind parses the passed in request data in JSON format and serializes them
// in v. Returns an error if Content-Type does not match the format JSON
// are not supported or encoding error occurred parsing the data.
func JSONBind(r *http.Request, v interface{}) error {
	mediatype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch mediatype {
	case "application/json":
		break
	case "":
		return errors.New("unknown content-type")
	default:
		return fmt.Errorf("unsupported content-type %s", mediatype)
	}
	charset, ok := params["charset"]
	if !ok {
		charset = "UTF-8"
	}
	if strings.ToUpper(charset) != "UTF-8" {
		return fmt.Errorf("unsupported charset %s", charset)
	}
	// parse the data from the request
	return json.NewDecoder(r.Body).Decode(v)
}
