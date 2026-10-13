// Copyright 2026 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: 	MIT

package wfw

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/immanent-tech/go-syndication/validation"
)

func init() {
	validation.RegisterStructValidation(commentSubmissionCustomValidation, CommentSubmission{})
}

// NewCommentRequest builds the HTTP POST request the spec documents: Content-Type: text/xml, body containing the
// marshaled CommentSubmission.
func NewCommentRequest(endpoint string, c CommentSubmission) (*http.Request, error) {
	if err := validation.ValidateStruct(c); err != nil {
		return nil, err
	}
	body, err := xml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("marshaling comment submission: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(append([]byte(xml.Header), body...)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml")
	return req, nil
}

func commentSubmissionCustomValidation(sl validator.StructLevel) {
	c := sl.Current().Interface().(CommentSubmission)
	// Enforce "there must be either a title or a description."
	if c.Title == "" && c.Description == "" {
		sl.ReportError(
			c,
			"CommentSubmission",
			"CommentSubmission",
			"required",
			"must contain a title or a description",
		)
	}
}
