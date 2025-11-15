package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type HTTPRequest struct {
	Method string // LIST, LOOKUP, INSERT, UPDATE, etc.
	Path   string // O termo ou recurso
	Body   string // Corpo da requisição (para INSERT/UPDATE)
}

func (r HTTPRequest) String() string {
	if r.Body != "" {
		return fmt.Sprintf("%s /%s\r\nBody: %s\r\n\r\n", r.Method, r.Path, r.Body)
	}
	return fmt.Sprintf("%s /%s\r\n\r\n", r.Method, r.Path)
}

func (r HTTPRequest) Bytes() []byte {
	return []byte(r.String())
}

type HTTPResponse struct {
	StatusCode int
	Message    string
}

func (r HTTPResponse) String() string {
	return fmt.Sprintf("%d %s: %s", r.StatusCode, http.StatusText(r.StatusCode), r.Message)
}

func (r HTTPResponse) Bytes() []byte {
	return []byte(r.String())
}

func ParseHTTPRequest(data []byte) (*HTTPRequest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid request format")
	}

	lines := bytes.Split(data, []byte("\r\n"))
	if len(lines) < 1 {
		return nil, fmt.Errorf("invalid request format")
	}

	parts := bytes.Fields(lines[0])
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid request line format")
	}

	method := strings.ToUpper(string(parts[0]))
	path := strings.TrimPrefix(string(parts[1]), "/")

	request := &HTTPRequest{
		Method: method,
		Path:   path,
	}

	for _, line := range lines[1:] {
		if bytes.HasPrefix(line, []byte("Body: ")) {
			request.Body = string(bytes.TrimPrefix(line, []byte("Body: ")))
			break
		}
	}

	return request, nil
}

func GetEmoji(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "\u2705"
	} else if statusCode >= 300 && statusCode < 400 {
		return "\u26a0\ufe0f"
	} else if statusCode >= 400 {
		return "\u274c"
	} else {
		return "\u2139\ufe0f"
	}
}
