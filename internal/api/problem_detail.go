package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func (c *Client) FetchProblemDetail(ctx context.Context, titleSlug string) ([]byte, error) {
	query := `
	query questionDetail($titleSlug: String!) {
	  question(titleSlug: $titleSlug) {
	    questionId
	    questionFrontendId
	    title
	    titleSlug
	    difficulty
	    sampleTestCase
	    metaData
	    codeSnippets {
	      lang
	      langSlug
	      code
	    }
	    content
	  }
	}`

	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": map[string]interface{}{"titleSlug": titleSlug},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return io.ReadAll(resp.Body)
}
