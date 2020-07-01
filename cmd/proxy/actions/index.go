package actions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index"
)

// indexHandler implements GET baseURL/index
func indexHandler(index index.Indexer) http.HandlerFunc {
	const op errors.Op = "actions.IndexHandler"
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err   error
			limit int
			since time.Time
		)
		if limitStr := r.FormValue("limit"); limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if sinceStr := r.FormValue("since"); sinceStr != "" {
			since, err = time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if limit <= 0 {
			limit = 2000
		}
		list, err := index.Lines(r.Context(), since, limit)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		enc := json.NewEncoder(w)
		for _, meta := range list {
			if err = enc.Encode(meta); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
	}
}
