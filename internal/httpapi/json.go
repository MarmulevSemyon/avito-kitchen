package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

func decodeJSON(
	r *http.Request,
	target any,
) error {
	decoder := json.NewDecoder(
		io.LimitReader(r.Body, 1<<20),
	)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New(
				"request body must contain one JSON object",
			)
		}

		return err
	}

	return nil
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func parsePositiveID(
	value string,
) (int64, error) {
	id, err := strconv.ParseInt(
		value,
		10,
		64,
	)
	if err != nil || id <= 0 {
		return 0, errors.New(
			"id must be a positive integer",
		)
	}

	return id, nil
}
