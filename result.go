package wemvc

import (
	"bytes"
	"io"
)

// Result define the action result struct
type Result struct {
	respFile    string
	redURL      string

	Writer      *bytes.Buffer
	StatusCode  int
	ContentType string
	Encoding    string
	Headers     map[string]string
}

func (res *Result) Write(data []byte) {
	if res.Writer == nil {
		res.Writer = &bytes.Buffer{}
	}
	res.Writer.Write(data)
}

func (res *Result) GetOutput() []byte {
	return res.Writer.Bytes()
}

func (res *Result) ClearHeader() {
	res.Headers = nil
}

func (res *Result) ClearOutput() {
	res.Writer = nil
}

func (res *Result) Clear() {
	res.ClearHeader()
	res.ClearOutput()
}

func (res *Result) GetWriter() io.Writer {
	return res.Writer
}

// NewResult create a blank action result
func NewResult() *Result {
	return &Result{
		StatusCode:  200,
		ContentType: "text/html",
		Encoding:    "utf-8",
		Headers:     map[string]string{},
		Writer:      &bytes.Buffer{},
	}
}
