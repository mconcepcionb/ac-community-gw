package azerothsoap

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func envelopeResponse(result string) string {
	return xmlHeader + `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body>` +
		`<executeCommandResponse xmlns="urn:AC"><result>` + result + `</result></executeCommandResponse>` +
		`</soap:Body></soap:Envelope>`
}

const xmlHeader = `<?xml version="1.0" encoding="utf-8"?>`

func TestExecuteSuccess(t *testing.T) {
	var gotAuth bool
	var gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		gotAuth = ok && user == "admin" && pass == "secret"
		if r.Header.Get("SOAPAction") == "" {
			t.Error("missing SOAPAction header")
		}
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		_, _ = io.WriteString(w, envelopeResponse("Command executed."))
	}))
	defer server.Close()

	client, err := New(Config{URL: server.URL, Username: "admin", Password: "secret"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	result, err := client.Execute(context.Background(), ".server info")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result != "Command executed." {
		t.Fatalf("result = %q", result)
	}
	if !gotAuth {
		t.Fatal("basic auth not sent correctly")
	}
	if !strings.Contains(gotBody, "<commandString>.server info</commandString>") {
		t.Fatalf("command not serialized: %s", gotBody)
	}
}

func TestExecuteEscapesSpecialCharacters(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = io.WriteString(w, envelopeResponse("ok"))
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL})
	if _, err := client.Execute(context.Background(), `.announce <b>&"quoted"</b>`); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(gotBody, "&lt;b&gt;&amp;&#34;quoted&#34;&lt;/b&gt;") {
		t.Fatalf("special characters not escaped: %s", gotBody)
	}
}

func TestExecuteSOAPFault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, xmlHeader+`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><soap:Fault><faultcode>SOAP-ENV:Server</faultcode><faultstring>command failed</faultstring></soap:Fault></soap:Body></soap:Envelope>`)
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL})
	_, err := client.Execute(context.Background(), ".invalid")
	if !errors.Is(err, ErrSOAPFault) {
		t.Fatalf("expected ErrSOAPFault, got %v", err)
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Fatalf("fault detail missing: %v", err)
	}
}

func TestExecuteHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "nope")
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL})
	_, err := client.Execute(context.Background(), ".server info")
	if !errors.Is(err, ErrHTTPStatus) {
		t.Fatalf("expected ErrHTTPStatus, got %v", err)
	}
}

func TestExecuteProtocolError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "this is not xml")
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL})
	_, err := client.Execute(context.Background(), ".server info")
	if !errors.Is(err, ErrProtocol) {
		t.Fatalf("expected ErrProtocol, got %v", err)
	}
}

func TestExecuteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		_, _ = io.WriteString(w, envelopeResponse("late"))
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL, Timeout: 200 * time.Millisecond})
	_, err := client.Execute(context.Background(), ".server info")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "azerothsoap") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRequiresURL(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestExecuteMissingResponseElement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, xmlHeader+`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body></soap:Body></soap:Envelope>`)
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL})
	_, err := client.Execute(context.Background(), ".server info")
	if !errors.Is(err, ErrProtocol) {
		t.Fatalf("expected ErrProtocol, got %v", err)
	}
}

func TestExecuteHTTPStatusDoesNotLeakBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "upstream-secret-detail")
	}))
	defer server.Close()

	client, _ := New(Config{URL: server.URL})
	_, err := client.Execute(context.Background(), ".server info")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "upstream-secret-detail") {
		t.Fatalf("upstream body leaked into error: %v", err)
	}
}
