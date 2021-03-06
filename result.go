package wemvc

import (
	"bytes"
	"fmt"
	"net/http"
)

// Result the request result interface
type Result interface {
	ExecResult(w http.ResponseWriter, r *http.Request)
}

// FileResult the file result
type FileResult struct {
	ContentType string
	FilePath    string
}

// ExecResult execute the result
func (fr *FileResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	if len(fr.ContentType) > 0 {
		w.Header().Add("Content-Type", fr.ContentType)
	}
	http.ServeFile(w, r, fr.FilePath)
}

// RedirectResult the redirect result
type RedirectResult struct {
	RedirectURL string
	StatusCode  int
}

// ExecResult execute the redirect result
func (rr *RedirectResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	var statusCode = 301
	if rr.StatusCode != 301 {
		statusCode = 302
	}
	http.Redirect(w, r, rr.RedirectURL, statusCode)
}

// ContentResult define the action result struct
type ContentResult struct {
	Writer      *bytes.Buffer
	StatusCode  int
	ContentType string
	Encoding    string
	Headers     map[string]string
}

// Header get the content result header
func (cr *ContentResult) Header() map[string]string {
	if cr.Headers == nil {
		cr.Headers = make(map[string]string)
	}
	return cr.Headers
}

// Write write content bytes into content result
func (cr *ContentResult) Write(data []byte) {
	if cr.Writer == nil {
		cr.Writer = &bytes.Buffer{}
	}
	cr.Writer.Write(data)
}

// Output get the output bytes from content result
func (cr *ContentResult) Output() []byte {
	if cr.Writer == nil {
		return nil
	}
	return cr.Writer.Bytes()
}

// ClearHeader clear the http header in content result
func (cr *ContentResult) ClearHeader() {
	cr.Headers = nil
}

// ClearOutput clear the output buffer
func (cr *ContentResult) ClearOutput() {
	cr.Writer = nil
}

// Clear clear the http headers and output buffer
func (cr *ContentResult) Clear() {
	cr.ClearHeader()
	cr.ClearOutput()
}

// ExecResult execute the content result
func (cr *ContentResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	if cr.Headers != nil {
		for k, v := range cr.Headers {
			if k == "Content-Type" {
				continue
			}
			w.Header().Add(k, v)
		}
	}
	if len(cr.ContentType) > 0 {
		encoding := cr.Encoding
		if len(encoding) == 0 {
			encoding = "utf-8"
		}
		contentType := fmt.Sprintf("%s;charset=%s", cr.ContentType, encoding)
		w.Header().Add("Content-Type", contentType)
	}
	if cr.StatusCode != 200 {
		w.WriteHeader(cr.StatusCode)
	}
	output := cr.Output()
	if len(output) > 0 {
		w.Write(output)
	}
}

// NewResult create a blank content result
func NewResult() *ContentResult {
	return &ContentResult{
		StatusCode:  200,
		ContentType: "text/html",
		Encoding:    "utf-8",
	}
}
