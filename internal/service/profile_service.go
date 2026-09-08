package service

import (
	"context"
	"encoding/json"
	"errors"
	"leetcli/internal/domain"
	"sort"
	"strconv"
)

type ProfileFetcher interface {
	FetchProfile(ctx context.Context, username string) ([]byte, error)
	FetchProblems(ctx context.Context, limit, skip int) ([]byte, error)
	FetchProblemDetail(ctx context.Context, titleSlug string) ([]byte, error)
	FetchDailyChallenge(ctx context.Context) ([]byte, error)
	FetchContests(ctx context.Context, username string) ([]byte, error)
}

type ProfileService struct {
	client   ProfileFetcher
	username string
}

func NewProfileService(client ProfileFetcher, username string) *ProfileService {
	return &ProfileService{client: client, username: username}
}

func (s *ProfileService) GetProfile(ctx context.Context) (domain.Profile, error) {

	raw, err := s.client.FetchProfile(ctx, s.username)
	if err != nil {
		return domain.Profile{}, err
	}

	var response struct {
		Data struct {
			MatchedUser struct {
				Username string `json:"username"`
				Profile  struct {
					Ranking    int    `json:"ranking"`
					Reputation int    `json:"reputation"`
					UserSlug   string `json:"userSlug"`
				} `json:"profile"`
				SubmitStats struct {
					AC []struct {
						Difficulty string `json:"difficulty"`
						Count      int    `json:"count"`
					} `json:"acSubmissionNum"`
				} `json:"submitStats"`
				LanguageStats []struct {
					Name  string `json:"languageName"`
					Count int    `json:"problemsSolved"`
				} `json:"languageProblemCount"`
				TagProblemCounts struct {
					Advanced []struct {
						TagName        string `json:"tagName"`
						ProblemsSolved int    `json:"problemsSolved"`
					} `json:"advanced"`
					Intermediate []struct {
						TagName        string `json:"tagName"`
						ProblemsSolved int    `json:"problemsSolved"`
					} `json:"intermediate"`
					Fundamental []struct {
						TagName        string `json:"tagName"`
						ProblemsSolved int    `json:"problemsSolved"`
					} `json:"fundamental"`
				} `json:"tagProblemCounts"`
			} `json:"matchedUser"`
		} `json:"data"`
	}

	if err := json.Unmarshal(raw, &response); err != nil {
		return domain.Profile{}, err
	}

	user := response.Data.MatchedUser

	profile := domain.Profile{
		Username:   user.Username,
		Rank:       user.Profile.Ranking,
		Reputation: user.Profile.Reputation,
		LeetCodeTotals: domain.LeetCodeTotals{
			Easy:   800,
			Medium: 1600,
			Hard:   700,
		},
	}

	for _, d := range user.SubmitStats.AC {
		switch d.Difficulty {
		case "Easy":
			profile.Difficulty.Easy = d.Count
		case "Medium":
			profile.Difficulty.Medium = d.Count
		case "Hard":
			profile.Difficulty.Hard = d.Count
		case "All":
			profile.Difficulty.Total = d.Count
		}
	}

	for _, l := range user.LanguageStats {
		profile.Languages = append(profile.Languages,
			domain.Language{Name: l.Name, Count: l.Count})
	}

	sort.Slice(profile.Languages, func(i, j int) bool {
		return profile.Languages[i].Count > profile.Languages[j].Count
	})

	skillMap := make(map[string]int)
	for _, s := range user.TagProblemCounts.Advanced {
		skillMap[s.TagName] += s.ProblemsSolved
	}
	for _, s := range user.TagProblemCounts.Intermediate {
		skillMap[s.TagName] += s.ProblemsSolved
	}
	for _, s := range user.TagProblemCounts.Fundamental {
		skillMap[s.TagName] += s.ProblemsSolved
	}

	for name, count := range skillMap {
		profile.Skills = append(profile.Skills, domain.Skill{Name: name, Count: count})
	}

	sort.Slice(profile.Skills, func(i, j int) bool {
		return profile.Skills[i].Count > profile.Skills[j].Count
	})

	return profile, nil
}

func (s *ProfileService) GetProblems(ctx context.Context, limit, skip int) ([]domain.Problem, int, error) {
	raw, err := s.client.FetchProblems(ctx, limit, skip)
	if err != nil {
		return nil, 0, err
	}

	var response struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Data struct {
			ProblemsetQuestionList struct {
				Total     int `json:"total"`
				Questions []struct {
					FrontendQuestionID string `json:"frontendQuestionId"`
					Title              string `json:"title"`
					TitleSlug          string `json:"titleSlug"`
					Difficulty         string `json:"difficulty"`
					Status             string `json:"status"`
				} `json:"questions"`
			} `json:"problemsetQuestionList"`
		} `json:"data"`
	}

	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, 0, err
	}

	if len(response.Errors) > 0 {
		return nil, 0, errors.New(response.Errors[0].Message)
	}

	questions := response.Data.ProblemsetQuestionList.Questions
	problems := make([]domain.Problem, len(questions))
	for i, q := range questions {
		id, _ := strconv.Atoi(q.FrontendQuestionID)
		problems[i] = domain.Problem{
			ID:         id,
			Title:      q.Title,
			Difficulty: q.Difficulty,
			Slug:       q.TitleSlug,
			Status:     q.Status,
		}
	}

	return problems, response.Data.ProblemsetQuestionList.Total, nil
}
