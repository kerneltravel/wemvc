package wemvc

import (
	"bytes"
	"io"
)

// ActionResult action result interface
type ActionResult interface {
	SetStatusCode(int)
	GetStatusCode() int

	SetContentType(string)
	GetContentType() string

	SetEncoding(string)
	GetEncoding() string

	SetHeader(string, string)
	GetHeaders() map[string]string

	Write([]byte)
	GetWriter() io.Writer
	GetOutput() []byte

	ClearHeader()
	ClearOutput()
	Clear()
}

type actionResult struct {
	resFile     string
	redURL      string
	statusCode  int
	contentType string
	encoding    string
	headers     map[string]string
	writer      bytes.Buffer
}

func (ares *actionResult) SetStatusCode(code int) {
	ares.statusCode = code
}

func (ares *actionResult) GetStatusCode() int {
	return ares.statusCode
}

func (ares *actionResult) SetContentType(c string) {
	ares.contentType = c
}

func (ares *actionResult) GetContentType() string {
	return ares.contentType
}

func (ares *actionResult) SetEncoding(e string) {
	ares.encoding = e
}

func (ares *actionResult) GetEncoding() string {
	return ares.encoding
}

func (ares *actionResult) SetHeader(key, value string) {
	if ares.headers == nil {
		ares.headers = make(map[string]string)
	}
	ares.headers[key] = value
}

func (ares *actionResult) GetHeaders() map[string]string {
	if ares.headers == nil {
		ares.headers = make(map[string]string)
	}
	return ares.headers
}

func (ares *actionResult) Write(data []byte) {
	ares.writer.Write(data)
}

func (ares *actionResult) GetOutput() []byte {
	return ares.writer.Bytes()
}

func (ares *actionResult) ClearHeader() {
	ares.headers = make(map[string]string)
}

func (ares *actionResult) ClearOutput() {
	ares.writer = bytes.Buffer{}
}

func (ares *actionResult) Clear() {
	ares.ClearHeader()
	ares.ClearOutput()
}

func (ares *actionResult) GetWriter() io.Writer {
	return &ares.writer
}

// NewActionResult create a blank action result
func NewActionResult() ActionResult {
	return &actionResult{
		statusCode:  200,
		contentType: "text/html",
		encoding:    "utf-8",
	}
}
