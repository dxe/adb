package transport

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dxe/adb/model"
)

func ActivistsSearchHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	result, err := model.QueryActivists(authedUser, options, repo)
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, result)
}
