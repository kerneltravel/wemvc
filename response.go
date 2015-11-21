package wemvc

import (
	"bytes"
	"io"
)

type Response interface {
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

type response struct {
	resFile     string
	redUrl      string
	statusCode  int
	contentType string
	encoding    string
	headers     map[string]string
	writer      bytes.Buffer
}

func (this *response) SetStatusCode(code int) {
	this.statusCode = code
}

func (this *response) GetStatusCode() int {
	return this.statusCode
}

func (this *response) SetContentType(c string) {
	this.contentType = c
}

func (this *response) GetContentType() string {
	return this.contentType
}

func (this *response) SetEncoding(e string) {
	this.encoding = e
}

func (this *response) GetEncoding() string {
	return this.encoding
}

func (this *response) SetHeader(key, value string) {
	if this.headers == nil {
		this.headers = make(map[string]string)
	}
	this.headers[key] = value
}

func (this *response) GetHeaders() map[string]string {
	if this.headers == nil {
		this.headers = make(map[string]string)
	}
	return this.headers
}

func (this *response) Write(data []byte) {
	this.writer.Write(data)
}

func (this *response) GetOutput() []byte {
	return this.writer.Bytes()
}

func (this *response) ClearHeader() {
	this.headers = make(map[string]string)
}

func (this *response) ClearOutput() {
	this.writer = bytes.Buffer{}
}

func (this *response) Clear() {
	this.ClearHeader()
	this.ClearOutput()
}

func (this *response) GetWriter() io.Writer {
	return &this.writer
}

func NewResponse() Response {
	return &response{
		statusCode:  200,
		contentType: "text/html",
		encoding:    "utf-8",
	}
}
