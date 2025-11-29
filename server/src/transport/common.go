package transport

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

// Temporarily make public for use by main package until all of its transport logic is
// migrated to this package.
func WriteJSON(w io.Writer, v interface{}) {
	writeJSON(w, v)
}

func writeJSON(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		log.Printf("Error writing JSON! %v", err.Error())
		//panic(err)
	}
}

/* Accepts a non-nil error, logs it, and sends an error response */
func sendErrorMessage(w http.ResponseWriter, status int, err error) {
	if err == nil {
		panic(errors.Wrap(err, "Cannot send error message if error is nil"))
	}
	log.Printf("ERROR: %+v\n", err.Error())

	w.WriteHeader(status)
	writeJSON(w, map[string]string{
		"status":  "error",
		"message": err.Error(),
	})
}
