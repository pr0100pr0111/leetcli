package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const contestsFixture = `{"data":{"userContestRanking":{"rating":1353.198,"topPercentage":92.55},"userContestRankingHistory":[{"attended":true,"rating":1411.755,"ranking":9091,"contest":{"title":"Weekly Contest 212","startTime":1603593000}},{"attended":false,"ranking":null,"rating":null,"contest":{"title":"Biweekly Contest 70","startTime":1610000000}},{"attended":true,"rating":1353.198,"ranking":11056,"contest":{"title":"Weekly Contest 296","startTime":1654396200}}]}}`

func TestGetContests(t *testing.T) {
	client := &mockClient{contestsResponse: []byte(contestsFixture)}
	svc := NewProfileService(client, "radewoosh")

	h, err := svc.GetContests(context.Background())
	if err != nil {
		t.Fatalf("GetContests: %v", err)
	}

	if client.contestsUsername != "radewoosh" {
		t.Errorf("username = %q", client.contestsUsername)
	}
	if !h.HasRating || h.Rating != 1353.198 || h.TopPercent != 92.55 {
		t.Errorf("summary = %+v", h)
	}
	if len(h.Results) != 3 {
		t.Fatalf("results = %d", len(h.Results))
	}
	if h.Results[0].Title != "Weekly Contest 212" || h.Results[0].Ranking != 9091 {
		t.Errorf("result[0] = %+v", h.Results[0])
	}
	if !h.Results[0].HasRating || h.Results[0].Rating != 1411.755 {
		t.Errorf("result[0] rating = %+v", h.Results[0])
	}
	if h.Results[1].Attended || h.Results[1].HasRating || h.Results[1].Ranking != 0 {
		t.Errorf("result[1] should be without rating: %+v", h.Results[1])
	}
	if got := h.Results[1].Time.Year(); got != 2021 {
		t.Errorf("result[1] year = %d", got)
	}
	if h.Results[2].Title != "Weekly Contest 296" {
		t.Errorf("result[2] = %+v", h.Results[2])
	}
	if h.AttendedCount() != 2 {
		t.Errorf("AttendedCount = %d", h.AttendedCount())
	}
}

func TestGetContestsEmpty(t *testing.T) {
	client := &mockClient{contestsResponse: []byte(`{"data":{"userContestRanking":null,"userContestRankingHistory":[]}}`)}
	svc := NewProfileService(client, "nobody")

	h, err := svc.GetContests(context.Background())
	if err != nil {
		t.Fatalf("GetContests: %v", err)
	}
	if h.HasRating || len(h.Results) != 0 {
		t.Errorf("empty history = %+v", h)
	}
}

func TestGetContestsUnsorted(t *testing.T) {
	client := &mockClient{contestsResponse: []byte(`{"data":{"userContestRanking":null,"userContestRankingHistory":[{"attended":true,"rating":100.5,"ranking":5,"contest":{"title":"B","startTime":2000000000}},{"attended":true,"rating":90.5,"ranking":6,"contest":{"title":"A","startTime":1000000000}}]}}`)}
	svc := NewProfileService(client, "someone")

	h, err := svc.GetContests(context.Background())
	if err != nil {
		t.Fatalf("GetContests: %v", err)
	}
	if h.Results[0].Title != "A" || h.Results[1].Title != "B" {
		t.Errorf("results not sorted by time: %q, %q", h.Results[0].Title, h.Results[1].Title)
	}
}

func TestGetContestsErrors(t *testing.T) {
	cases := []struct {
		name    string
		client  *mockClient
		wantSub string
	}{
		{"graphql error", &mockClient{contestsResponse: []byte(`{"errors":[{"message":"rating blew up"}]}`)}, "rating blew up"},
		{"invalid json", &mockClient{contestsResponse: []byte("nope")}, "invalid"},
		{"network", &mockClient{contestsErr: errors.New("net down")}, "net down"},
	}
	for _, c := range cases {
		svc := NewProfileService(c.client, "someone")
		_, err := svc.GetContests(context.Background())
		if err == nil {
			t.Errorf("%s: expected error", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.wantSub) {
			t.Errorf("%s: err = %q, want substring %q", c.name, err, c.wantSub)
		}
	}
}
