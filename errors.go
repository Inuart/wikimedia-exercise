package shortdescription

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrUpstream        = errors.New("upstream error")
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInternal        = errors.New("internal error")
)

func responseError(r *http.Response) error {
	if r.StatusCode < http.StatusBadRequest { // a status code >= 400 is an error
		return nil
	}

	var err error

	switch r.StatusCode {
	case http.StatusNotFound:
		err = ErrNotFound
	case http.StatusBadRequest:
		err = ErrInternal // it's our fault
	case http.StatusBadGateway:
		err = ErrUpstream
	default:
		err = ErrUpstream
	}

	body, rerr := io.ReadAll(r.Body)
	if rerr != nil {
		return fmt.Errorf(
			"request failed with %w (also, body could not be read: %v)",
			err, rerr,
		)
	}

	if len(body) < 1 {
		body = []byte(http.StatusText(r.StatusCode))
	}

	return fmt.Errorf("request failed with %w: %s", err, body)
}
