package service

import (
	"context"
	"encoding/json"
	"testing"

	"leetcli/internal/api"
)

type mockClient struct {
	response         []byte
	err              error
	problemsResponse []byte
	problemsErr      error
	problemsLimit    int
	problemsSkip     int
}

func (m *mockClient) FetchProfile(ctx context.Context, username string) ([]byte, error) {
	return m.response, m.err
}

func (m *mockClient) FetchProblems(ctx context.Context, limit, skip int) ([]byte, error) {
	if m.problemsErr != nil {
		return nil, m.problemsErr
	}
	m.problemsLimit = limit
	m.problemsSkip = skip
	return m.problemsResponse, nil
}

func TestProfileService_GetProfile(t *testing.T) {
	rawResponse := `{
		"data": {
			"matchedUser": {
				"username": "testuser",
				"profile": {
					"ranking": 1000,
					"reputation": 500,
					"userSlug": "testuser"
				},
				"submitStats": {
					"acSubmissionNum": [
						{"difficulty": "Easy", "count": 100},
						{"difficulty": "Medium", "count": 50},
						{"difficulty": "Hard", "count": 10},
						{"difficulty": "All", "count": 160}
					]
				},
				"languageProblemCount": [
					{"languageName": "Go", "problemsSolved": 80},
					{"languageName": "Python", "problemsSolved": 50},
					{"languageName": "JavaScript", "problemsSolved": 30}
				],
				"tagProblemCounts": {
					"advanced": [
						{"tagName": "Dynamic Programming", "problemsSolved": 25},
						{"tagName": "Graph", "problemsSolved": 15}
					],
					"intermediate": [
						{"tagName": "Binary Search", "problemsSolved": 30}
					],
					"fundamental": [
						{"tagName": "Array", "problemsSolved": 50},
						{"tagName": "String", "problemsSolved": 40}
					]
				}
			}
		}
	}`

	client := &mockClient{response: []byte(rawResponse)}
	service := NewProfileService(client, "testuser")

	profile, err := service.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile returned error: %v", err)
	}

	if profile.Username != "testuser" {
		t.Errorf("Username = %q, want %q", profile.Username, "testuser")
	}
	if profile.Rank != 1000 {
		t.Errorf("Rank = %d, want %d", profile.Rank, 1000)
	}
	if profile.Reputation != 500 {
		t.Errorf("Reputation = %d, want %d", profile.Reputation, 500)
	}
	if profile.Difficulty.Easy != 100 {
		t.Errorf("Easy = %d, want %d", profile.Difficulty.Easy, 100)
	}
	if profile.Difficulty.Medium != 50 {
		t.Errorf("Medium = %d, want %d", profile.Difficulty.Medium, 50)
	}
	if profile.Difficulty.Hard != 10 {
		t.Errorf("Hard = %d, want %d", profile.Difficulty.Hard, 10)
	}
	if profile.Difficulty.Total != 160 {
		t.Errorf("Total = %d, want %d", profile.Difficulty.Total, 160)
	}
	if profile.LeetCodeTotals.Easy != 800 {
		t.Errorf("LeetCodeTotals.Easy = %d, want %d", profile.LeetCodeTotals.Easy, 800)
	}
	if profile.LeetCodeTotals.Medium != 1600 {
		t.Errorf("LeetCodeTotals.Medium = %d, want %d", profile.LeetCodeTotals.Medium, 1600)
	}
	if profile.LeetCodeTotals.Hard != 700 {
		t.Errorf("LeetCodeTotals.Hard = %d, want %d", profile.LeetCodeTotals.Hard, 700)
	}
	if len(profile.Languages) != 3 {
		t.Errorf("Languages count = %d, want 3", len(profile.Languages))
	}
	if profile.Languages[0].Name != "Go" || profile.Languages[0].Count != 80 {
		t.Errorf("Top language = %+v, want {Go 80}", profile.Languages[0])
	}
	if len(profile.Skills) != 5 {
		t.Errorf("Skills count = %d, want 5", len(profile.Skills))
	}
	if profile.Skills[0].Name != "Array" || profile.Skills[0].Count != 50 {
		t.Errorf("Top skill = %+v, want {Array 50}", profile.Skills[0])
	}
}

func TestProfileService_GetProblems(t *testing.T) {
	rawResponse := `{
		"data": {
			"problemsetQuestionList": {
				"total": 3,
				"questions": [
					{"frontendQuestionId": "1", "title": "Two Sum", "titleSlug": "two-sum", "difficulty": "Easy", "status": null},
					{"frontendQuestionId": "2", "title": "Add Two Numbers", "titleSlug": "add-two-numbers", "difficulty": "Medium", "status": "Solved"},
					{"frontendQuestionId": "3", "title": "Longest Substring", "titleSlug": "longest-substring", "difficulty": "Medium", "status": null}
				]
			}
		}
	}`

	client := &mockClient{problemsResponse: []byte(rawResponse)}
	service := NewProfileService(client, "testuser")

	problems, total, err := service.GetProblems(context.Background(), 100, 0)
	if err != nil {
		t.Fatalf("GetProblems returned error: %v", err)
	}

	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(problems) != 3 {
		t.Fatalf("problems count = %d, want 3", len(problems))
	}
	if problems[0].ID != 1 || problems[0].Title != "Two Sum" || problems[0].Difficulty != "Easy" {
		t.Errorf("first problem = %+v", problems[0])
	}
	if problems[1].Status != "Solved" {
		t.Errorf("second problem status = %q, want %q", problems[1].Status, "Solved")
	}
	if client.problemsLimit != 100 || client.problemsSkip != 0 {
		t.Errorf("FetchProblems called with limit=%d skip=%d, want limit=100 skip=0", client.problemsLimit, client.problemsSkip)
	}
}

func TestProfileService_GetProblems_GraphQLError(t *testing.T) {
	rawResponse := `{"errors":[{"message":"Variable \"$username\" is never used"}]}`

	client := &mockClient{problemsResponse: []byte(rawResponse)}
	service := NewProfileService(client, "testuser")

	_, _, err := service.GetProblems(context.Background(), 100, 0)
	if err == nil {
		t.Fatal("expected error for GraphQL errors array, got nil")
	}
}

func TestProfileService_GetProfile_Error(t *testing.T) {
	client := &mockClient{err: api.ErrNetwork}
	service := NewProfileService(client, "testuser")

	_, err := service.GetProfile(context.Background())
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProfileService_GetProfile_InvalidJSON(t *testing.T) {
	client := &mockClient{response: []byte("invalid json")}
	service := NewProfileService(client, "testuser")

	_, err := service.GetProfile(context.Background())
	if err == nil {
		t.Fatal("Expected JSON unmarshal error, got nil")
	}
}

func TestGraphQLResponseStructure(t *testing.T) {
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"matchedUser": map[string]interface{}{
				"username": "test",
				"profile": map[string]interface{}{
					"ranking":    100,
					"reputation": 200,
				},
				"submitStats": map[string]interface{}{
					"acSubmissionNum": []map[string]interface{}{
						{"difficulty": "Easy", "count": 10},
					},
				},
				"languageProblemCount": []map[string]interface{}{
					{"languageName": "Go", "problemsSolved": 5},
				},
				"tagProblemCounts": map[string]interface{}{
					"advanced":     []map[string]interface{}{},
					"intermediate": []map[string]interface{}{},
					"fundamental":  []map[string]interface{}{},
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal test response: %v", err)
	}

	client := &mockClient{response: data}
	service := NewProfileService(client, "test")

	profile, err := service.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if profile.Username != "test" {
		t.Errorf("Username mismatch")
	}
	if len(profile.Languages) != 1 {
		t.Errorf("Expected 1 language")
	}
	if len(profile.Skills) != 0 {
		t.Errorf("Expected 0 skills")
	}
	if profile.LeetCodeTotals.Easy != 800 {
		t.Errorf("Expected LeetCodeTotals.Easy = 800")
	}
}
