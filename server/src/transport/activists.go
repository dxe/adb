package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dxe/adb/model"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
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

func ActivistGetHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, db *sqlx.DB) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	activistID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, http.StatusBadRequest, fmt.Errorf("invalid activist id %s: %w", rawID, err))
		return
	}

	activist, err := model.GetActivistJSONForUser(db, authedUser, model.GetActivistOptions{ID: activistID})
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			sendErrorMessage(w, http.StatusNotFound, fmt.Errorf("no activist found with id %d", activistID))
		} else if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, map[string]any{
		"activist": activist,
	})
}
