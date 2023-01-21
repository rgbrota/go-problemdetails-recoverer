package problemdetailsrecoverer

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	problemdetails "github.com/rgbrota/go-problemdetails"
	"github.com/stretchr/testify/assert"
)

func panicingHandler(http.ResponseWriter, *http.Request) {
	panic("lorem ipsum")
}

func panicingAbortHandler(http.ResponseWriter, *http.Request) {
	panic(http.ErrAbortHandler)
}

func TestNew(t *testing.T) {
	expected := generateExpectedResponse(JSON, internalServerErrorType)

	ts := httptest.NewServer(New(http.HandlerFunc(panicingHandler)))
	defer ts.Close()

	res, _ := http.Get(fmt.Sprintf("%s/", ts.URL))

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, contentTypeProblemJSON, res.Header.Get(contentTypeHeaderName))
	assert.Equal(t, expected, string(body))
}

func TestNewWithConfig(t *testing.T) {
	type args struct {
		responseFormat     ResponseFormat
		problemDetailsType string
	}
	type test struct {
		name     string
		args     args
		expected string
	}

	tests := []test{
		{
			name: "json_default_type",
			args: args{
				responseFormat:     JSON,
				problemDetailsType: internalServerErrorType,
			},
			expected: generateExpectedResponse(JSON, internalServerErrorType),
		},
		{
			name: "json_custom_type",
			args: args{
				responseFormat:     JSON,
				problemDetailsType: "unexpected-error",
			},
			expected: generateExpectedResponse(JSON, "unexpected-error"),
		},
		{
			name: "xml_default_type",
			args: args{
				responseFormat:     XML,
				problemDetailsType: internalServerErrorType,
			},
			expected: generateExpectedResponse(XML, internalServerErrorType),
		},
		{
			name: "xml_custom_type",
			args: args{
				responseFormat:     XML,
				problemDetailsType: "unexpected-error",
			},
			expected: generateExpectedResponse(XML, "unexpected-error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var contentType string

			if test.args.responseFormat == XML {
				contentType = contentTypeProblemXML
			} else {
				contentType = contentTypeProblemJSON
			}

			config := RecovererConfig{
				LogFunc:            nil,
				LogAllStack:        false,
				ResponseFormat:     test.args.responseFormat,
				ProblemDetailsType: test.args.problemDetailsType,
			}

			ts := httptest.NewServer(NewWithConfig(http.HandlerFunc(panicingHandler), config))
			defer ts.Close()

			res, _ := http.Get(fmt.Sprintf("%s/", ts.URL))

			body, _ := io.ReadAll(res.Body)
			res.Body.Close()

			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
			assert.Equal(t, contentType, res.Header.Get(contentTypeHeaderName))
			assert.Equal(t, test.expected, string(body))
		})
	}
}

func TestRecovererLogs(t *testing.T) {
	type args struct {
		logFunc     func(err error, stack []byte)
		logAllStack bool
	}
	type test struct {
		name     string
		args     args
		expected string
	}

	tests := []test{
		{
			name: "logFunc_no_stack",
			args: args{
				logFunc: func(err error, stack []byte) {
					log.Printf("[LogFunc] test %v", err)
				},
				logAllStack: false,
			},
			expected: "[LogFunc] test lorem ipsum",
		},
		{
			name: "logFunc_all_stack",
			args: args{
				logFunc: func(err error, stack []byte) {
					log.Printf("[LogFunc] stack test %v %s", err, stack)
				},
				logAllStack: true,
			},
			expected: "[LogFunc] stack test lorem ipsum goroutine",
		},
		{
			name: "log_all_stack",
			args: args{
				logFunc:     nil,
				logAllStack: true,
			},
			expected: "[PANIC]: lorem ipsum goroutine",
		},
		{
			name: "log_no_stack",
			args: args{
				logFunc:     nil,
				logAllStack: false,
			},
			expected: "[PANIC]: lorem ipsum",
		},
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := RecovererConfig{
				LogFunc:            test.args.logFunc,
				LogAllStack:        test.args.logAllStack,
				ResponseFormat:     JSON,
				ProblemDetailsType: internalServerErrorType,
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/", nil)

			h := NewWithConfig(http.HandlerFunc(panicingHandler), config)

			h.ServeHTTP(w, req)

			assert.Contains(t, buf.String(), test.expected)

			buf.Reset()
		})
	}
}

func TestRecovererAbortHandler(t *testing.T) {
	defer func() {
		r := recover()
		assert.Equal(t, http.ErrAbortHandler, r)
	}()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	h := New(http.HandlerFunc(panicingAbortHandler))

	h.ServeHTTP(w, req)
}

func generateExpectedResponse(format ResponseFormat, problemDetailsType string) string {
	buf := new(bytes.Buffer)
	pd := problemdetails.New(problemDetailsType, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError, "", "")

	if format == XML {
		xml.NewEncoder(buf).Encode(pd)
	} else {
		json.NewEncoder(buf).Encode(pd)
	}

	return buf.String()
}
