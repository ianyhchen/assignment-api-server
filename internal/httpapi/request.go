package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/ianyhchen/assignment-api-server/internal/task"
)

const maxRequestBodyBytes int64 = 1 << 20 // 1 MiB

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("request body exceeds %d byte limit", maxRequestBodyBytes)

		case errors.Is(err, io.EOF):
			return errors.New("request body is required")

		default:
			return errors.New("request body contains invalid JSON")
		}
	}

	var extraValue any

	if err := decoder.Decode(&extraValue); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON value")
	}

	return nil
}

func parseTaskID(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, task.ErrInvalidTaskID
	}

	return id, nil
}
