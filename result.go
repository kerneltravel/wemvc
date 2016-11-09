package wemvc

import (
	"bytes"
	"net/http"
	"fmt"
)

type Result interface {
	ExecResult(w http.ResponseWriter, r *http.Request)
}

type FileResult struct {
	ContentType string
	FilePath    string
}

func (fr *FileResult) ExecResult(w http.ResponseWriter, r *http.Request){
	w.Header().Add("Content-Type", fr.ContentType)
	http.ServeFile(w, r, fr.FilePath)
}

type RedirectResult struct {
	RedirectUrl string
	StatusCode  int
}

func (rr *RedirectResult) ExecResult(w http.ResponseWriter, r *http.Request) {
	var statusCode = 301
	if rr.StatusCode != 301 {
		statusCode = 302
	}
	http.Redirect(w, r, rr.RedirectUrl, statusCode)
}

// Result define the action result struct
type ContentResult struct {
	Writer      *bytes.Buffer
	StatusCode  int
	ContentType string
	Encoding    string
	Headers     map[string]string
}

func (cr *ContentResult) Header() map[string]string {
	if cr.Headers == nil {
		cr.Headers = make(map[string]string)
	}
	return cr.Headers
}

func (cr *ContentResult) Write(data []byte) {
	if cr.Writer == nil {
		cr.Writer = &bytes.Buffer{}
	}
	cr.Writer.Write(data)
}

// GetOutput get the output bytes
func (cr *ContentResult) Output() []byte {
	if cr.Writer == nil {
		return nil
	}
	return cr.Writer.Bytes()
}

// ClearHeader clear the http header
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

// NewResult create a blank action result
func NewResult() *ContentResult {
	return &ContentResult{
		StatusCode:  200,
		ContentType: "text/html",
		Encoding:    "utf-8",
	}
}
