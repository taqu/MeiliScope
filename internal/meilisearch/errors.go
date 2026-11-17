package meilisearch

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"strings"
)

// ErrorKind classifies the type of error returned by the client.
type ErrorKind int

const (
	ErrKindNetwork           ErrorKind = iota // Generic network failure
	ErrKindDNS                                // DNS resolution failure
	ErrKindConnectionRefused                  // Connection refused
	ErrKindTimeout                            // Request timed out
	ErrKindCanceled                           // Request was canceled
	ErrKindTLS                                // TLS / certificate failure
	ErrKindAuth                               // HTTP 401 Unauthorized
	ErrKindForbidden                          // HTTP 403 Forbidden
	ErrKindNotFound                           // HTTP 404 Not Found
	ErrKindBadRequest                         // HTTP 400 Bad Request
	ErrKindRateLimit                          // HTTP 429 Too Many Requests
	ErrKindServer                             // HTTP 5xx Server Error
	ErrKindInvalidJSON                        // Invalid JSON in response
	ErrKindUnexpectedContent                  // Unexpected Content-Type
	ErrKindResponseTooLarge                   // Response body exceeded the limit
)

// APIError is a structured error returned by the Meilisearch client.
// Error messages are sanitized — credentials are never included.
type APIError struct {
	Kind   ErrorKind
	Status int    // HTTP status code; 0 for transport-level errors
	Code   string // Meilisearch error code, e.g. "index_not_found"
	Type   string // Meilisearch error type
	Msg    string // Human-readable message (no credentials)
	Link   string // Documentation link, when provided
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("meilisearch: %s (HTTP %d, code=%s)", e.Msg, e.Status, e.Code)
	}
	if e.Status > 0 {
		return fmt.Sprintf("meilisearch: %s (HTTP %d)", e.Msg, e.Status)
	}
	return "meilisearch: " + e.Msg
}

// IsNotFound reports whether err is a 404 Not Found error.
func IsNotFound(err error) bool { return kindOf(err) == ErrKindNotFound }

// IsAuth reports whether err is an authentication error (HTTP 401).
func IsAuth(err error) bool { return kindOf(err) == ErrKindAuth }

// IsForbidden reports whether err is an authorization error (HTTP 403).
func IsForbidden(err error) bool { return kindOf(err) == ErrKindForbidden }

// IsCanceled reports whether err represents a canceled request.
func IsCanceled(err error) bool { return kindOf(err) == ErrKindCanceled }

// IsTimeout reports whether err represents a timed-out request.
func IsTimeout(err error) bool { return kindOf(err) == ErrKindTimeout }

// IsBadRequest reports whether err is an HTTP 400 Bad Request error.
func IsBadRequest(err error) bool { return kindOf(err) == ErrKindBadRequest }

// IsRateLimit reports whether err is an HTTP 429 Too Many Requests error.
func IsRateLimit(err error) bool { return kindOf(err) == ErrKindRateLimit }

// IsServerError reports whether err is an HTTP 5xx server error.
func IsServerError(err error) bool { return kindOf(err) == ErrKindServer }

// IsInvalidJSON reports whether err indicates an invalid JSON response.
func IsInvalidJSON(err error) bool { return kindOf(err) == ErrKindInvalidJSON }

// IsUnexpectedContent reports whether err indicates an unexpected Content-Type.
func IsUnexpectedContent(err error) bool { return kindOf(err) == ErrKindUnexpectedContent }

// IsResponseTooLarge reports whether err indicates the response exceeded the size limit.
func IsResponseTooLarge(err error) bool { return kindOf(err) == ErrKindResponseTooLarge }

func kindOf(err error) ErrorKind {
	var e *APIError
	if errors.As(err, &e) {
		return e.Kind
	}
	return -1
}

// classifyTransportError maps a net/http transport error to an APIError.
// It never includes credential information in the message.
func classifyTransportError(err error) *APIError {
	if errors.Is(err, context.Canceled) {
		return &APIError{Kind: ErrKindCanceled, Msg: "request canceled"}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &APIError{Kind: ErrKindTimeout, Msg: "request timed out"}
	}

	// TLS errors
	var certInvalid *x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		return &APIError{Kind: ErrKindTLS, Msg: "TLS certificate is invalid"}
	}
	var hostnameMismatch *x509.HostnameError
	if errors.As(err, &hostnameMismatch) {
		return &APIError{Kind: ErrKindTLS, Msg: "TLS certificate hostname mismatch"}
	}
	var unknownAuth x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		return &APIError{Kind: ErrKindTLS, Msg: "TLS certificate authority not recognized"}
	}

	// DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsTimeout {
			return &APIError{Kind: ErrKindTimeout, Msg: "DNS lookup timed out"}
		}
		return &APIError{Kind: ErrKindDNS, Msg: "DNS resolution failed: host not found"}
	}

	// Network operation errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Timeout() {
			return &APIError{Kind: ErrKindTimeout, Msg: "connection timed out"}
		}
		if opErr.Err != nil {
			s := opErr.Err.Error()
			if strings.Contains(s, "refused") || strings.Contains(s, "actively refused") {
				return &APIError{Kind: ErrKindConnectionRefused, Msg: "connection refused"}
			}
		}
		return &APIError{Kind: ErrKindNetwork, Msg: "network error"}
	}

	// Generic timeout interface
	type timeouter interface{ Timeout() bool }
	if te, ok := err.(timeouter); ok && te.Timeout() {
		return &APIError{Kind: ErrKindTimeout, Msg: "request timed out"}
	}

	return &APIError{Kind: ErrKindNetwork, Msg: "network error"}
}

// apiErrorBody is the JSON structure of a Meilisearch error response.
type apiErrorBody struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Type    string `json:"type"`
	Link    string `json:"link"`
}

// httpStatusToAPIError converts an HTTP status code and optional body to an APIError.
func httpStatusToAPIError(status int, body *apiErrorBody) *APIError {
	e := &APIError{Status: status}
	if body != nil {
		e.Code = body.Code
		e.Type = body.Type
		e.Msg = body.Message
		e.Link = body.Link
	}

	switch {
	case status == 400:
		e.Kind = ErrKindBadRequest
		if e.Msg == "" {
			e.Msg = "bad request"
		}
	case status == 401:
		e.Kind = ErrKindAuth
		if e.Msg == "" {
			e.Msg = "authentication required"
		}
	case status == 403:
		e.Kind = ErrKindForbidden
		if e.Msg == "" {
			e.Msg = "access denied"
		}
	case status == 404:
		e.Kind = ErrKindNotFound
		if e.Msg == "" {
			e.Msg = "not found"
		}
	case status == 429:
		e.Kind = ErrKindRateLimit
		if e.Msg == "" {
			e.Msg = "rate limit exceeded"
		}
	case status >= 500:
		e.Kind = ErrKindServer
		if e.Msg == "" {
			e.Msg = fmt.Sprintf("server error (%d)", status)
		}
	default:
		e.Kind = ErrKindServer
		if e.Msg == "" {
			e.Msg = fmt.Sprintf("unexpected HTTP status %d", status)
		}
	}

	return e
}
