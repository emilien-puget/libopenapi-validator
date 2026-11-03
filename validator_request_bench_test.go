// Copyright 2023-2026 Princess Beef Heavy Industries, LLC / Dave Shanley
// SPDX-License-Identifier: MIT

package validator

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
)

// requestBenchSpec holds the happy-path shapes a gateway validates: a single
// path parameter, several parameters of mixed locations, and a JSON body.
const requestBenchSpec = `openapi: 3.1.0
info:
  title: bench
  version: 1.0.0
paths:
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: ok
  /orders/{orderId}/items/{itemId}:
    get:
      parameters:
        - name: orderId
          in: path
          required: true
          schema:
            type: integer
            minimum: 1
        - name: itemId
          in: path
          required: true
          schema:
            type: string
            pattern: '^[A-Z]{2}[0-9]{4}$'
        - name: status
          in: query
          schema:
            type: string
            enum: [open, closed]
        - name: X-Request-Id
          in: header
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: ok
  /burgers:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        '200':
          description: ok
`

func prepareRequestBenchValidator(b *testing.B) Validator {
	b.Helper()

	doc, err := libopenapi.NewDocument([]byte(requestBenchSpec))
	if err != nil {
		b.Fatalf("error creating document: %v", err)
	}

	v, errs := NewValidator(doc)
	if len(errs) > 0 {
		b.Fatalf("errors building validator: %v", errs)
	}
	b.Cleanup(v.Release)
	return v
}

func benchRequest(b *testing.B, method, target string, body io.Reader) *http.Request {
	b.Helper()

	req, err := http.NewRequest(method, target, body)
	if err != nil {
		b.Fatalf("error building request: %v", err)
	}
	return req
}

// BenchmarkValidateHttpRequestSync measures one valid GET with one uuid path
// parameter through the full synchronous entry point: routing plus path,
// header, cookie and query parameter validation.
func BenchmarkValidateHttpRequestSync(b *testing.B) {
	v := prepareRequestBenchValidator(b)
	req := benchRequest(b, http.MethodGet, "/users/3f2504e0-4f89-11d3-9a0c-0305e82c3301", nil)

	b.ReportAllocs()

	for b.Loop() {
		valid, validationErrors := v.ValidateHttpRequestSync(req)
		if !valid {
			b.Fatalf("request should validate, got %d errors", len(validationErrors))
		}
	}
}

// BenchmarkValidateHttpRequestSyncPathQueryHeader measures one valid GET that
// exercises routing, two path parameters, one query parameter and one header
// parameter.
func BenchmarkValidateHttpRequestSyncPathQueryHeader(b *testing.B) {
	v := prepareRequestBenchValidator(b)
	req := benchRequest(b, http.MethodGet, "/orders/12/items/AB1234?status=closed", nil)
	req.Header.Set("X-Request-Id", "3f2504e0-4f89-11d3-9a0c-0305e82c3301")

	b.ReportAllocs()

	for b.Loop() {
		valid, validationErrors := v.ValidateHttpRequestSync(req)
		if !valid {
			b.Fatalf("request should validate, got %d errors", len(validationErrors))
		}
	}
}

// BenchmarkValidateHttpRequestSyncJSONBody measures one valid POST with a JSON
// body. The request is rebuilt per iteration, because validation consumes the
// body reader; the repo benchmark pattern does the same.
func BenchmarkValidateHttpRequestSyncJSONBody(b *testing.B) {
	v := prepareRequestBenchValidator(b)
	body := []byte(`{"name":"cheeseburger"}`)

	b.ReportAllocs()

	for b.Loop() {
		req := benchRequest(b, http.MethodPost, "/burgers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		valid, validationErrors := v.ValidateHttpRequestSync(req)
		if !valid {
			b.Fatalf("request should validate, got %d errors", len(validationErrors))
		}
	}
}
