package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"leetcli/internal/domain"
)

type dailyResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Date     string `json:"date"`
			Link     string `json:"link"`
			Question struct {
				FrontendID string `json:"questionFrontendId"`
				Title      string `json:"title"`
				TitleSlug  string `json:"titleSlug"`
				Difficulty string `json:"difficulty"`
				TopicTags  []struct {
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"topicTags"`
			} `json:"question"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

func (s *ProfileService) GetDailyChallenge(ctx context.Context) (domain.DailyChallenge, error) {
	raw, err := s.client.FetchDailyChallenge(ctx)
	if err != nil {
		return domain.DailyChallenge{}, err
	}

	var resp dailyResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return domain.DailyChallenge{}, err
	}

	if len(resp.Errors) > 0 {
		return domain.DailyChallenge{}, errors.New(resp.Errors[0].Message)
	}

	q := resp.Data.ActiveDailyCodingChallengeQuestion
	if q.Question.TitleSlug == "" {
		return domain.DailyChallenge{}, errors.New("daily challenge is not available")
	}

	challenge := domain.DailyChallenge{
		Date:       q.Date,
		Slug:       q.Question.TitleSlug,
		Title:      q.Question.Title,
		Difficulty: q.Question.Difficulty,
	}
	challenge.ID, _ = strconv.Atoi(q.Question.FrontendID)
	for _, t := range q.Question.TopicTags {
		challenge.Topics = append(challenge.Topics, domain.Topic{Name: t.Name, Slug: t.Slug})
	}

	return challenge, nil
}
