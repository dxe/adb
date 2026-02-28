package transport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/dxe/adb/model"
)

type QueryActivistResultJSON struct {
	Activists  []model.ActivistJSON    `json:"activists"`
	Pagination QueryActivistPagination `json:"pagination"`
}

type QueryActivistPagination struct {
	NextCursor string `json:"next_cursor"`
}

func ActivistsSearchHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	result, err := model.QueryActivists(authedUser, options, repo)
	if err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, QueryActivistResultJSON{
		Activists: model.BuildActivistJSONArray(result.Activists),
		Pagination: QueryActivistPagination{
			NextCursor: result.Pagination.NextCursor,
		},
	})
}
