package shortdescription

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (d Describer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// allow multiple origins / client websites
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if req.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	query := req.URL.Query()

	person := query.Get("person")
	if person == "" {
		http.Error(w, "the 'person' query parameter cannot be empty", http.StatusBadRequest)
		return
	}

	descr, err := d.ShortDescription(req.Context(), person, req.UserAgent())
	if err != nil {
		errCode := http.StatusInternalServerError
		if errors.Is(err, ErrUpstream) {
			errCode = http.StatusBadGateway
		}

		if errors.Is(err, ErrNotFound) {
			errCode = http.StatusNotFound
		}

		if errors.Is(err, ErrInvalidArgument) {
			errCode = http.StatusBadRequest
		}

		http.Error(w, err.Error(), errCode)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(descr); err != nil {
		http.Error(w,
			"error while encoding the short description: "+err.Error(),
			http.StatusInternalServerError,
		)
	}
}
