package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type Spec struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

type Response struct {
	Status     int
	Headers    http.Header
	Body       []byte
	ParsedJSON any
}

func Do(spec Spec, timeout time.Duration) (*Response, error) {
	if spec.URL == "" {
		return nil, errors.New("request url is empty")
	}
	method := spec.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader
	if spec.Body != nil {
		switch v := spec.Body.(type) {
		case string:
			body = bytes.NewBufferString(v)
		case []byte:
			body = bytes.NewBuffer(v)
		default:
			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(encoded)
			if spec.Headers == nil {
				spec.Headers = map[string]string{}
			}
			if _, ok := spec.Headers["Content-Type"]; !ok {
				spec.Headers["Content-Type"] = "application/json"
			}
		}
	}
	req, err := http.NewRequest(method, spec.URL, body)
	if err != nil {
		return nil, err
	}
	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parsed any
	if len(b) > 0 {
		_ = json.Unmarshal(b, &parsed)
	}
	return &Response{
		Status:     resp.StatusCode,
		Headers:    resp.Header,
		Body:       b,
		ParsedJSON: parsed,
	}, nil
}
