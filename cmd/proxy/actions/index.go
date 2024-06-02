package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index"
	"github.com/gomods/athens/pkg/log"
	"github.com/sirupsen/logrus"
)

// indexHandler implements GET baseURL/index.
func indexHandler(index index.Indexer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		list, err := getIndexLines(r, index)
		if err != nil {
			log.EntryFromContext(ctx).SystemErr(err)
			http.Error(w, err.Error(), errors.Kind(err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		for _, meta := range list {
			if err = enc.Encode(meta); err != nil {
				log.EntryFromContext(ctx).SystemErr(err)
				fmt.Fprintln(w, err)
				return
			}
		}
	}
}

func getIndexLines(r *http.Request, index index.Indexer) ([]*index.Line, error) {
	const op errors.Op = "actions.IndexHandler"
	var (
		err   error
		limit = 2000
		since time.Time
	)
	if limitStr := r.FormValue("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return nil, errors.E(op, err, errors.KindBadRequest, logrus.InfoLevel)
		}
	}
	if sinceStr := r.FormValue("since"); sinceStr != "" {
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return nil, errors.E(op, err, errors.KindBadRequest, logrus.InfoLevel)
		}
	}
	list, err := index.Lines(r.Context(), since, limit)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return list, nil
}
