package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func (s *PifinaHttpServer) HandleProxyRequest(rw http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		http.Error(rw, "Missing endpoint query param", http.StatusBadRequest)
		return
	}

	endpointDetail := s.ed.Get(endpoint)
	if endpointDetail == nil {
		http.Error(rw, "Endpoint unknown", http.StatusBadRequest)
		return
	}

	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	r.RequestURI = ""

	delHopHeaders(r.Header)

	// Proxy request => overwrite destination
	r.URL.Scheme = "http"
	r.URL.Host = fmt.Sprintf("%s:%d", endpointDetail.Address.String(), endpointDetail.Port)
	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	s.logger.Info("Proxying API request to controller", "remoteAddr", r.RemoteAddr, "method", r.Method, "url", r.URL)

	resp, err := client.Do(r)
	if err != nil {
		http.Error(rw, "Server Error", http.StatusInternalServerError)
		s.logger.Error("Proxy API request failed", "remoteAddr", r.RemoteAddr, "url", r.URL, "err", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		s.logger.Warn("Proxy API request returned a non-ok status code", "remoteAddr", r.RemoteAddr, "status", resp.StatusCode, "url", r.URL)
	}

	delHopHeaders(resp.Header)

	copyHeader(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
}
