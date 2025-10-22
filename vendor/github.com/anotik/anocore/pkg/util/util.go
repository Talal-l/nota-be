package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anotik/anocore/pkg/errs"
)

type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

// FindModuleRoot walks up the directory tree until it finds a go.mod file
func FindModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func GetFullURL(req *http.Request) string {
	scheme := req.URL.Scheme
	if scheme == "" {
		// Check if the request was made over HTTPS
		if req.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + req.Host + req.URL.String()
}

func FromJSON(r io.Reader, v any) error {
	if r == nil {
		return &errs.JSONError{
			Message: "reader is nil",
			Err:     nil,
		}
	}

	d, err := io.ReadAll(r)
	if err != nil {
		return &errs.UnexpectedError{
			Message: "failed to read json bytes",
			Err:     err,
		}
	}

	err = json.Unmarshal(d, &v)

	if err != nil {
		var unmarshalErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalErr) {
			jErr := &errs.JSONError{
				Message: "failed to decode JSON",
				Input:   string(d),
				Field:   &unmarshalErr.Field,
				Err:     err,
			}
			return jErr
		}

		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			jErr := &errs.JSONError{
				Message: "Invalid JSON syntax",
				Input:   string(d),
				Err:     err,
			}
			return jErr
		}
	}

	return nil
}

func ToJSON[T any](v T) ([]byte, error) {
	return json.Marshal(v)
}

func ToMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}

func FromMap(m map[string]any, v any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// Encode writes a JSON-encoded value of type T to the HTTP response writer.
// It sets the Content-Type header to "application/json", writes the provided
// HTTP status code, and encodes the value v as JSON to the response body.
// Returns an error if JSON encoding fails.
func Encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// Decode reads the JSON-encoded request body and decodes it into a value of type T.
// It returns the decoded value and any error encountered during decoding.
// The request body is expected to contain valid JSON that can be unmarshaled into type T.
func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// In this code, T has to implement the Validator interface,
// and the Valid method must return zero problems in order
// for the object to be considered successfully decoded.
//
// It’s safe to return nil for problems because we are going to
// check len(problems), which will be 0 for a nil map, but which won’t panic.
func DecodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}
