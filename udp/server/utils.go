package server

import (
	"fmt"
	"strings"
	"sync"
)

// Response represents a server response
type Response struct {
	StatusCode int
	Message    string
}

// ProcessDictCommand processes a dictionary command and returns a response
func ProcessDictCommand(command string, dict *Dictionary, mutex *sync.RWMutex) *Response {
	if command == "" {
		return &Response{
			StatusCode: 400,
			Message:    "Empty command",
		}
	}

	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return &Response{
			StatusCode: 400,
			Message:    "Invalid command format",
		}
	}

	method := strings.ToUpper(parts[0])

	switch method {
	case "LIST":
		return handleList(dict, mutex)

	case "LOOKUP":
		if len(parts) < 2 {
			return &Response{
				StatusCode: 400,
				Message:    "Usage: LOOKUP <term>",
			}
		}
		term := parts[1]
		return handleLookup(term, dict, mutex)

	case "INSERT":
		if len(parts) < 3 {
			return &Response{
				StatusCode: 400,
				Message:    "Usage: INSERT <term> <definition>",
			}
		}
		term := parts[1]
		// Join remaining parts as definition (allow spaces)
		definition := strings.Join(parts[2:], " ")
		return handleInsert(term, definition, dict, mutex)

	case "UPDATE":
		if len(parts) < 3 {
			return &Response{
				StatusCode: 400,
				Message:    "Usage: UPDATE <term> <new_definition>",
			}
		}
		term := parts[1]
		// Join remaining parts as definition (allow spaces)
		definition := strings.Join(parts[2:], " ")
		return handleUpdate(term, definition, dict, mutex)

	default:
		return &Response{
			StatusCode: 400,
			Message:    fmt.Sprintf("Unknown command: %s", method),
		}
	}
}

// handleList returns all terms in the dictionary
func handleList(dict *Dictionary, mutex *sync.RWMutex) *Response {
	mutex.RLock()
	defer mutex.RUnlock()

	terms := dict.List()
	if len(terms) == 0 {
		return &Response{
			StatusCode: 200,
			Message:    "[empty]",
		}
	}

	// Format as newline-separated list
	message := strings.Join(terms, "\n")
	return &Response{
		StatusCode: 200,
		Message:    message,
	}
}

// handleLookup searches for a term in the dictionary
func handleLookup(term string, dict *Dictionary, mutex *sync.RWMutex) *Response {
	mutex.RLock()
	defer mutex.RUnlock()

	definition, exists := dict.LookUp(term)
	if !exists {
		return &Response{
			StatusCode: 404,
			Message:    fmt.Sprintf("Term not found: %s", term),
		}
	}

	return &Response{
		StatusCode: 200,
		Message:    definition,
	}
}

// handleInsert inserts a new term into the dictionary
func handleInsert(term, definition string, dict *Dictionary, mutex *sync.RWMutex) *Response {
	term = strings.TrimSpace(term)
	definition = strings.TrimSpace(definition)

	if term == "" {
		return &Response{
			StatusCode: 400,
			Message:    "Term cannot be empty",
		}
	}

	if definition == "" {
		return &Response{
			StatusCode: 400,
			Message:    "Definition cannot be empty",
		}
	}

	mutex.Lock()
	defer mutex.Unlock()

	success := dict.Insert(term, definition)
	if !success {
		return &Response{
			StatusCode: 409,
			Message:    fmt.Sprintf("Term already exists: %s", term),
		}
	}

	return &Response{
		StatusCode: 201,
		Message:    fmt.Sprintf("Term inserted: %s", term),
	}
}

// handleUpdate updates an existing term in the dictionary
func handleUpdate(term, newDefinition string, dict *Dictionary, mutex *sync.RWMutex) *Response {
	term = strings.TrimSpace(term)
	newDefinition = strings.TrimSpace(newDefinition)

	if term == "" {
		return &Response{
			StatusCode: 400,
			Message:    "Term cannot be empty",
		}
	}

	if newDefinition == "" {
		return &Response{
			StatusCode: 400,
			Message:    "Definition cannot be empty",
		}
	}

	mutex.Lock()
	defer mutex.Unlock()

	success := dict.Update(term, newDefinition)
	if !success {
		return &Response{
			StatusCode: 404,
			Message:    fmt.Sprintf("Term not found: %s", term),
		}
	}

	return &Response{
		StatusCode: 200,
		Message:    fmt.Sprintf("Term updated: %s", term),
	}
}

// StatusCodeMessage returns HTTP-like status message for a code
func StatusCodeMessage(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 400:
		return "Bad Request"
	case 404:
		return "Not Found"
	case 409:
		return "Conflict"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}

// GetStatusEmoji returns an emoji for a status code
func GetStatusEmoji(code int) string {
	if code >= 200 && code < 300 {
		return "✅"
	} else if code >= 400 && code < 500 {
		return "⚠️"
	} else if code >= 500 {
		return "❌"
	}
	return "ℹ️"
}
