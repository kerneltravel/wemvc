package wemvc

import (
	//"log"
	"net/http"
	"os"
	"sync"
)

// CtxHandler the error handler define
type CtxHandler func(*http.Request) *Result

// FilterFunc request filter func
type FilterFunc func(ctx Context)

// NewServer create a new server based on the server root dir
func NewServer(webRoot string) Server {
	return newServer(webRoot)
}

// WaitForExit in if there is two or more server running in single process, the WaitForExit function should be called to prevent the main function return immediately
//noinspection GoUnusedExportedFunction
func WaitForExit() {
	serverWaiting.Wait()
}

// App the application singleton
var (
	serverWaiting = sync.WaitGroup{}
)

// WorkingDir get the current working directory
func WorkingDir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}
