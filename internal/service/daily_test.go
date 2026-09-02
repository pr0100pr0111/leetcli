package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const dailyChallengeResponse = `{"data":{"activeDailyCodingChallengeQuestion":{"date":"2026-10-07","link":"/problems/remove-invalid-parentheses/","question":{"questionFrontendId":"301","title":"Remove Invalid Parentheses","titleSlug":"remove-invalid-parentheses","difficulty":"Hard","topicTags":[{"name":"String","slug":"string"},{"name":"Backtracking","slug":"backtracking"},{"name":"Breadth-First Search","slug":"breadth-first-search"}]}}}}`

func TestGetDailyChallenge(t *testing.T) {
	client := &mockClient{dailyResponse: []byte(dailyChallengeResponse)}
	svc := NewProfileService(client, "testuser")

	d, err := svc.GetDailyChallenge(context.Background())
	if err != nil {
		t.Fatalf("GetDailyChallenge: %v", err)
	}
	if d.Date != "2026-10-07" {
		t.Errorf("Date = %q", d.Date)
	}
	if d.Slug != "remove-invalid-parentheses" {
		t.Errorf("Slug = %q", d.Slug)
	}
	if d.ID != 301 {
		t.Errorf("ID = %d", d.ID)
	}
	if d.Title != "Remove Invalid Parentheses" {
		t.Errorf("Title = %q", d.Title)
	}
	if d.Difficulty != "Hard" {
		t.Errorf("Difficulty = %q", d.Difficulty)
	}
	if len(d.Topics) != 3 {
		t.Fatalf("topics = %v", d.Topics)
	}
	if d.Topics[0].Name != "String" || d.Topics[0].Slug != "string" {
		t.Errorf("topic[0] = %+v", d.Topics[0])
	}
}

func TestGetDailyChallengeErrors(t *testing.T) {
	cases := []struct {
		name    string
		client  *mockClient
		wantSub string
	}{
		{"graphql error", &mockClient{dailyResponse: []byte(`{"errors":[{"message":"query blew up"}]}`)}, "query blew up"},
		{"empty data", &mockClient{dailyResponse: []byte(`{"data":{}}`)}, "not available"},
		{"invalid json", &mockClient{dailyResponse: []byte("nope")}, "invalid"},
		{"network", &mockClient{dailyErr: errors.New("net down")}, "net down"},
	}
	for _, c := range cases {
		svc := NewProfileService(c.client, "testuser")
		_, err := svc.GetDailyChallenge(context.Background())
		if err == nil {
			t.Errorf("%s: expected error", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.wantSub) {
			t.Errorf("%s: err = %q, want substring %q", c.name, err, c.wantSub)
		}
	}
}
