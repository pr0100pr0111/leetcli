package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const endpoint = "https://leetcode.com/graphql"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchProfile(ctx context.Context, username string) ([]byte, error) {

	query := `
	query getUserProfile($username: String!) {
	  matchedUser(username: $username) {
	    username
	    profile {
	      ranking
	      reputation
	    }
	    submitStats {
	      acSubmissionNum {
	        difficulty
	        count
	      }
	    }
	    languageProblemCount {
	      languageName
	      problemsSolved
	    }
	  }
	}`

	body, _ := json.Marshal(map[string]interface{}{
		"query": query,
		"variables": map[string]string{
			"username": username,
		},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}