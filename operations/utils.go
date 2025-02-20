package operations

import (
	"encoding/json"
	"net/http"
	// -- import --
	// -- end --
)

func JSON(o interface{}, w http.ResponseWriter) {
	b, err := json.Marshal(o)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// -- code --
// -- end --
