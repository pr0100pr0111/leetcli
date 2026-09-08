package api

import (
	"context"
)

func (c *Client) FetchContests(ctx context.Context, username string) ([]byte, error) {
	query := `
	query userContestRanking($username: String!) {
	  userContestRanking(username: $username) {
	    rating
	    topPercentage
	  }
	  userContestRankingHistory(username: $username) {
	    attended
	    rating
	    ranking
	    contest {
	      title
	      startTime
	    }
	  }
	}`
	return c.doQuery(ctx, query, map[string]string{"username": username})
}
