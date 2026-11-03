// Copyright 2023-2026 Princess Beef Heavy Industries, LLC / Dave Shanley
// SPDX-License-Identifier: MIT

package validator

import (
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
)

// requestBenchSpec is a minimal spec with one templated path and one uuid path
// parameter, the shape that gateway request validation sees on every call.
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
`

func prepareRequestBenchValidator(b *testing.B) (Validator, *http.Request) {
	b.Helper()

	doc, err := libopenapi.NewDocument([]byte(requestBenchSpec))
	if err != nil {
		b.Fatalf("error creating document: %v", err)
	}

	v, errs := NewValidator(doc)
	if len(errs) > 0 {
		b.Fatalf("errors building validator: %v", errs)
	}

	req, err := http.NewRequest(http.MethodGet, "/users/3f2504e0-4f89-11d3-9a0c-0305e82c3301", nil)
	if err != nil {
		b.Fatalf("error building request: %v", err)
	}
	return v, req
}

// BenchmarkValidateHttpRequestSync measures one valid request through the full
// synchronous entry point: routing plus path, header, cookie and query
// parameter validation.
func BenchmarkValidateHttpRequestSync(b *testing.B) {
	v, req := prepareRequestBenchValidator(b)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		valid, validationErrors := v.ValidateHttpRequestSync(req)
		if !valid {
			b.Fatalf("request should validate, got %d errors", len(validationErrors))
		}
	}
}
