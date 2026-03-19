package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"nabr/config"
)

type Response struct {
	StatusCode int
	Body       []byte
}

var pathParamRe = regexp.MustCompile(`\{(\w+)\}`)

func ExtractPathParams(url string) []string {
	matches := pathParamRe.FindAllStringSubmatch(url, -1)
	params := make([]string, 0, len(matches))
	for _, m := range matches {
		params = append(params, m[1])
	}
	return params
}

func Execute(cmd config.Command, params map[string]string) (*Response, error) {
	url := cmd.URL
	for key, val := range params {
		url = strings.ReplaceAll(url, "{"+key+"}", val)
	}

	var bodyReader io.Reader
	if cmd.Body != "" {
		bodyReader = strings.NewReader(cmd.Body)
	}

	req, err := http.NewRequest(cmd.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	for k, v := range cmd.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return &Response{StatusCode: resp.StatusCode, Body: body}, nil
}

func FormatJSON(body []byte, raw bool) string {
	if raw {
		return string(body)
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return string(body)
	}
	return buf.String()
}
