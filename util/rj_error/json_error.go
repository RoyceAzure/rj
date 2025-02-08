package rj_error

import (
	"errors"
	"io"
	"strings"
)

func JsonDecodeErrToMsg(err error) string {
	var msg string
	switch {
	case errors.Is(err, io.EOF):
		msg = "Request body is empty"
	case strings.Contains(err.Error(), "parsing time"):
		msg = "Invalid time format. Expected format: YYYY-MM-DDTHH:mm:ssZ"
	case strings.Contains(err.Error(), "cannot unmarshal"):
		msg = "Invalid data format in request"
	default:
		msg = "Invalid request payload"
	}
	return msg
}
