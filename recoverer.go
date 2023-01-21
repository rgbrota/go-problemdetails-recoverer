package problemdetailsrecoverer

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	problemdetails "github.com/rgbrota/go-problemdetails"
)

type ResponseFormat uint8

const (
	JSON ResponseFormat = iota
	XML
)

const (
	contentTypeHeaderName   = "Content-Type"
	contentTypeProblemXML   = "application/problem+xml"
	contentTypeProblemJSON  = "application/problem+json"
	internalServerErrorType = "https://www.rfc-editor.org/rfc/rfc7231#section-6.6.1"
)

// RecovererConfig specifies the config for the recoverer middleware.
type RecovererConfig struct {
	// LogFunc is a function for custom logging. It defaults to nil.
	LogFunc func(err error, stack []byte)
	// LogAllStack enables logging the whole stack trace. It defaults to true.
	LogAllStack bool
	// ResponseFormat specifies the format of the output. It defaults to JSON.
	ResponseFormat ResponseFormat
	// ProblemDetailsType specifies the Type value of the Problem Details instance
	// that will be created. It defaults to the RFC url of the Internal Server Error status.
	ProblemDetailsType string
}

var defaultRecoverConfig = RecovererConfig{
	LogFunc:            nil,
	LogAllStack:        true,
	ResponseFormat:     JSON,
	ProblemDetailsType: internalServerErrorType,
}

// New creates a middleware which recovers from panics and returns a HTTP status
// Internal Server Error with a Problem Details body following the RFC7807 specification.
func New(next http.Handler) http.Handler {
	return NewWithConfig(next, defaultRecoverConfig)
}

// NewWithConfig creates a new Recoverer middleware with the given config.
func NewWithConfig(next http.Handler, config RecovererConfig) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				if r == http.ErrAbortHandler {
					panic(r)
				}

				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				pd := problemdetails.New(config.ProblemDetailsType, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, "", "")

				if config.ResponseFormat == XML {
					w.Header().Set(contentTypeHeaderName, contentTypeProblemXML)
					w.WriteHeader(http.StatusInternalServerError)
					xml.NewEncoder(w).Encode(pd)
				} else {
					w.Header().Set(contentTypeHeaderName, contentTypeProblemJSON)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(pd)
				}

				var stack []byte

				if config.LogAllStack {
					stack = debug.Stack()
				}

				if config.LogFunc != nil {
					config.LogFunc(err, stack)
				} else {
					log.Printf("[PANIC]: %v %s\n", err, stack)
				}
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
