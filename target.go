package ddp

import (
	"net/http"
)

// Target is an interface for proxy target
type Target interface {
	Do(*http.Request) (*http.Response, error)
}
