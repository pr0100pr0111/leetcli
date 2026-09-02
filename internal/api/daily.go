package api

import (
	"context"
)

func (c *Client) FetchDailyChallenge(ctx context.Context) ([]byte, error) {
	query := `
	query activeDailyCodingChallengeQuestion {
	  activeDailyCodingChallengeQuestion {
	    date
	    link
	    question {
	      questionFrontendId
	      title
	      titleSlug
	      difficulty
	      topicTags {
	        name
	        slug
	      }
	    }
	  }
	}`
	return c.doQuery(ctx, query, map[string]string{})
}
