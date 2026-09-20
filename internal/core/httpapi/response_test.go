package httpapi

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

type decodeTarget struct {
	A string `json:"a"`
}

func TestDecodeJSONValid(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader(`{"a":"hello"}`))
	var dst decodeTarget
	if err := DecodeJSON(req, &dst); err != nil {
		t.Fatalf("DecodeJSON: %v", err)
	}
	if dst.A != "hello" {
		t.Fatalf("a = %q", dst.A)
	}
}

func TestDecodeJSONRejectsUnknownField(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader(`{"a":"x","b":1}`))
	var dst decodeTarget
	if err := DecodeJSON(req, &dst); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("error = %v, want ErrBadRequest", err)
	}
}

func TestDecodeJSONRejectsMultipleObjects(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader(`{"a":"x"}{"a":"y"}`))
	var dst decodeTarget
	if err := DecodeJSON(req, &dst); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("error = %v, want ErrBadRequest", err)
	}
}

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	body := `{"a":"` + strings.Repeat("x", maxRequestBody+1024) + `"}`
	req := httptest.NewRequest("POST", "/x", strings.NewReader(body))
	var dst decodeTarget
	if err := DecodeJSON(req, &dst); !errors.Is(err, ErrRequestTooLarge) {
		t.Fatalf("error = %v, want ErrRequestTooLarge", err)
	}
}

func TestDecodeJSONRejectsOversizedTailAfterValidPrefix(t *testing.T) {
	body := `{"a":"x"}` + strings.Repeat(" ", maxRequestBody+1024)
	req := httptest.NewRequest("POST", "/x", strings.NewReader(body))
	var dst decodeTarget
	if err := DecodeJSON(req, &dst); !errors.Is(err, ErrRequestTooLarge) {
		t.Fatalf("error = %v, want ErrRequestTooLarge", err)
	}
}

func TestDecodeJSONAcceptsBodyAtLimit(t *testing.T) {
	pad := strings.Repeat("x", maxRequestBody-len(`{"a":""}`))
	body := `{"a":"` + pad + `"}`
	if len(body) != maxRequestBody {
		t.Fatalf("test body length = %d, want %d", len(body), maxRequestBody)
	}
	req := httptest.NewRequest("POST", "/x", strings.NewReader(body))
	var dst decodeTarget
	if err := DecodeJSON(req, &dst); err != nil {
		t.Fatalf("DecodeJSON at limit: %v", err)
	}
}
