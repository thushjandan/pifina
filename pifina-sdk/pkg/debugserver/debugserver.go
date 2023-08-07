// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package debugserver

import (
	"context"
	"net/http"
	_ "net/http/pprof" // Register the pprof handlers
	"time"
)

type DebugServer struct {
	ds *http.Server
}

func NewDebugServer(address string) *DebugServer {
	return &DebugServer{
		ds: &http.Server{
			Addr:    address,
			Handler: http.DefaultServeMux,
		},
	}
}

func (s *DebugServer) StartDebugServer() {
	go func() {
		s.ds.ListenAndServe()
	}()
}

func (s *DebugServer) ShutdownDebugServer() {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.ds.Shutdown(timeoutCtx)
}
