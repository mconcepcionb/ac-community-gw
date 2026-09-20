// Package azerothsoap implements the AzerothCore CommandExecutor over SOAP.
//
// It is responsible for transport only: building and parsing SOAP XML, HTTP
// Basic authentication, timeouts and error translation. It has no knowledge of
// which commands exist or what they mean.
package azerothsoap

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

const (
	soapAction       = `"urn:AC#executeCommand"`
	maxResponseBytes = 1 << 20
)

var (
	// ErrSOAPFault is returned when AzerothCore responds with a SOAP fault.
	ErrSOAPFault = errors.New("azerothsoap: soap fault")
	// ErrHTTPStatus is returned for a non-200 HTTP response without a fault.
	ErrHTTPStatus = errors.New("azerothsoap: unexpected http status")
	// ErrProtocol is returned when the response cannot be parsed.
	ErrProtocol = errors.New("azerothsoap: invalid soap response")
)

// Config configures a SOAP client.
type Config struct {
	URL        string
	Username   string
	Password   string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Client is a transport-only AzerothCore SOAP client.
type Client struct {
	url      string
	username string
	password string
	client   *http.Client
}

var _ azerothcore.CommandExecutor = (*Client)(nil)

// New builds a SOAP client.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("azerothsoap: url must not be empty")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	} else if cfg.Timeout > 0 {
		clone := *httpClient
		if clone.Timeout == 0 {
			clone.Timeout = cfg.Timeout
		}
		httpClient = &clone
	}
	return &Client{
		url:      cfg.URL,
		username: cfg.Username,
		password: cfg.Password,
		client:   httpClient,
	}, nil
}

// Execute sends a single command and returns the textual result.
func (c *Client) Execute(ctx context.Context, command string) (string, error) {
	payload, err := buildEnvelope(command)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("azerothsoap: build request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)
	req.SetBasicAuth(c.username, c.password)
	if requestID := httpapi.RequestIDFromContext(ctx); requestID != "" {
		req.Header.Set("X-Request-Id", requestID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("azerothsoap: execute: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("azerothsoap: read response: %w", err)
	}

	envelope, parseErr := parseEnvelope(data)

	if resp.StatusCode != http.StatusOK {
		if parseErr == nil && envelope.Body.Fault != nil {
			return "", fmt.Errorf("%w: %s", ErrSOAPFault, envelope.Body.Fault)
		}
		return "", fmt.Errorf("%w: %s", ErrHTTPStatus, resp.Status)
	}
	if parseErr != nil {
		return "", fmt.Errorf("%w: %v", ErrProtocol, parseErr)
	}
	if envelope.Body.Fault != nil {
		return "", fmt.Errorf("%w: %s", ErrSOAPFault, envelope.Body.Fault)
	}
	if envelope.Body.ExecuteCommandResponse == nil {
		return "", fmt.Errorf("%w: missing executeCommandResponse", ErrProtocol)
	}
	return envelope.Body.ExecuteCommandResponse.Result, nil
}

func buildEnvelope(command string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">`)
	buf.WriteString(`<soap:Body><executeCommand xmlns="urn:AC"><commandString>`)
	if err := xml.EscapeText(&buf, []byte(command)); err != nil {
		return nil, fmt.Errorf("azerothsoap: escape command: %w", err)
	}
	buf.WriteString(`</commandString></executeCommand></soap:Body></soap:Envelope>`)
	return buf.Bytes(), nil
}

type soapEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    soapBody `xml:"Body"`
}

type soapBody struct {
	Fault                  *soapFault       `xml:"Fault"`
	ExecuteCommandResponse *executeResponse `xml:"executeCommandResponse"`
}

type executeResponse struct {
	Result string `xml:"result"`
}

type soapFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
	Detail      string `xml:"detail"`
}

func (f *soapFault) Error() string {
	if f.FaultCode == "" {
		return f.FaultString
	}
	return f.FaultCode + ": " + f.FaultString
}

func parseEnvelope(data []byte) (soapEnvelope, error) {
	var envelope soapEnvelope
	if err := xml.Unmarshal(data, &envelope); err != nil {
		return soapEnvelope{}, err
	}
	return envelope, nil
}
