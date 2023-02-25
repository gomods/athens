package actions

import (
	"encoding/json"
	"net/http"

	"github.com/gomods/athens/pkg/build"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(build.Data())
}
