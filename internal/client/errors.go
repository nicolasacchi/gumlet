package client

import "fmt"

type APIError struct {
	StatusCode int
	Code       string
	Title      string
	Detail     string
	Hint       string

	// Kind classifies client-side guard errors raised before a request
	// (StatusCode stays 0). Currently: "write_locked" (confirm gate).
	Kind string
}

func (e *APIError) Error() string {
	// Client-side guard errors (no HTTP status) render from Detail/Hint.
	if e.Kind != "" && e.StatusCode == 0 {
		msg := e.Detail
		if msg == "" {
			msg = e.Kind
		}
		if e.Hint != "" {
			return fmt.Sprintf("%s — %s", msg, e.Hint)
		}
		return msg
	}
	if e.Detail != "" {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Detail)
	}
	if e.Title != "" {
		return fmt.Sprintf("%d: %s", e.StatusCode, e.Title)
	}
	return fmt.Sprintf("API error %d", e.StatusCode)
}

func (e *APIError) ExitCode() int {
	if e.Kind == "write_locked" {
		return 6 // refused for safety, not failed — matches the otx/stx write-gate contract
	}
	switch e.StatusCode {
	case 401, 403:
		return 3
	default:
		return 1
	}
}

func hintForError(statusCode int) string {
	switch statusCode {
	case 401:
		return "check your API key: --api-key flag, GUMLET_API_KEY env, or 'gumlet config list'"
	case 403:
		return "your API key may lack the required permissions"
	case 429:
		return "rate limited (1000 req/hour) — reduce request frequency or wait"
	}
	return ""
}
