package service

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"leetcli/internal/domain"
)

type contestsResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		UserContestRanking *struct {
			Rating        float64 `json:"rating"`
			TopPercentage float64 `json:"topPercentage"`
		} `json:"userContestRanking"`
		UserContestRankingHistory []struct {
			Attended bool     `json:"attended"`
			Rating   *float64 `json:"rating"`
			Ranking  *int     `json:"ranking"`
			Contest  struct {
				Title     string `json:"title"`
				StartTime int64  `json:"startTime"`
			} `json:"contest"`
		} `json:"userContestRankingHistory"`
	} `json:"data"`
}

func (s *ProfileService) GetContests(ctx context.Context) (domain.ContestHistory, error) {
	raw, err := s.client.FetchContests(ctx, s.username)
	if err != nil {
		return domain.ContestHistory{}, err
	}

	var resp contestsResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return domain.ContestHistory{}, err
	}

	if len(resp.Errors) > 0 {
		return domain.ContestHistory{}, errors.New(resp.Errors[0].Message)
	}

	var history domain.ContestHistory
	if r := resp.Data.UserContestRanking; r != nil {
		history.HasRating = true
		history.Rating = r.Rating
		history.TopPercent = r.TopPercentage
	}

	for _, h := range resp.Data.UserContestRankingHistory {
		result := domain.ContestResult{
			Title:    h.Contest.Title,
			Time:     time.Unix(h.Contest.StartTime, 0),
			Attended: h.Attended,
		}
		if h.Rating != nil {
			result.HasRating = true
			result.Rating = *h.Rating
		}
		if h.Ranking != nil {
			result.Ranking = *h.Ranking
		}
		history.Results = append(history.Results, result)
	}

	sort.SliceStable(history.Results, func(i, j int) bool {
		return history.Results[i].Time.Before(history.Results[j].Time)
	})

	return history, nil
}
