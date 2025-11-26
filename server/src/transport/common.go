package transport

import (
	"encoding/json"
	"io"
	"log"

	"github.com/pkg/errors"
)

// Temporarily make public for use by main package until all of its transport logic is
// migrated to this package.
func WriteJSON(w io.Writer, v interface{}) {
	writeJSON(w, v)
}

// Temporarily make public for use by main package until all of its transport logic is
// migrated to this package.
func SendErrorMessage(w io.Writer, err error) {
	sendErrorMessage(w, err)
}

func writeJSON(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		log.Printf("Error writing JSON! %v", err.Error())
		//panic(err)
	}
}

/* Accepts a non-nil error and sends an error response */
func sendErrorMessage(w io.Writer, err error) {
	if err == nil {
		panic(errors.Wrap(err, "Cannot send error message if error is nil"))
	}
	log.Printf("ERROR: %+v\n", err.Error())
	writeJSON(w, map[string]string{
		"status":  "error",
		"message": err.Error(),
	})
}
