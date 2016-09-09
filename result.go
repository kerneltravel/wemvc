package wemvc

import (
	"bytes"
	"io"
)

// Result action result interface
type Result interface {
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

type result struct {
	resFile     string
	redURL      string
	statusCode  int
	contentType string
	encoding    string
	headers     map[string]string
	writer      bytes.Buffer
}

func (ares *result) SetStatusCode(code int) {
	ares.statusCode = code
}

func (ares *result) GetStatusCode() int {
	return ares.statusCode
}

func (ares *result) SetContentType(c string) {
	ares.contentType = c
}

func (ares *result) GetContentType() string {
	return ares.contentType
}

func (ares *result) SetEncoding(e string) {
	ares.encoding = e
}

func (ares *result) GetEncoding() string {
	return ares.encoding
}

func (ares *result) SetHeader(key, value string) {
	if ares.headers == nil {
		ares.headers = make(map[string]string)
	}
	ares.headers[key] = value
}

func (ares *result) GetHeaders() map[string]string {
	if ares.headers == nil {
		ares.headers = make(map[string]string)
	}
	return ares.headers
}

func (ares *result) Write(data []byte) {
	ares.writer.Write(data)
}

func (ares *result) GetOutput() []byte {
	return ares.writer.Bytes()
}

func (ares *result) ClearHeader() {
	ares.headers = nil
}

func (ares *result) ClearOutput() {
	ares.writer = bytes.Buffer{}
}

func (ares *result) Clear() {
	ares.ClearHeader()
	ares.ClearOutput()
}

func (ares *result) GetWriter() io.Writer {
	return &ares.writer
}

// NewResult create a blank action result
func NewResult() Result {
	return &result{
		statusCode:  200,
		contentType: "text/html",
		encoding:    "utf-8",
	}
}
