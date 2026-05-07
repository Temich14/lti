package http

import (
	"net"
	"net/http"
	"time"
)

func NewHTTPClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
}
func doRequestWithRetry(client *http.Client, req *http.Request, retries int) (*http.Response, error) {

	var resp *http.Response
	var err error

	for i := 0; i < retries; i++ {

		resp, err = client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
			continue
		}

		// Retry on rate limits and transient server errors.
		// 429 is common for LMS/NRPS throttling.
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		resp.Body.Close()

		// Respect Retry-After if present, otherwise fallback to incremental backoff.
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			// Retry-After can be seconds or an HTTP date; here we handle seconds.
			if sec, parseErr := time.ParseDuration(retryAfter + "s"); parseErr == nil {
				time.Sleep(sec)
			} else {
				time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
			}
		} else {
			time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
		}
	}

	return resp, err
}
