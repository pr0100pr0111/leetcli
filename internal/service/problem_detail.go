package service

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"regexp"
	"strconv"
	"strings"

	"leetcli/internal/domain"
)

var (
	exampleHeadRe = regexp.MustCompile(`(?s)<strong class="example">Example (\d+):</strong>(.*?)</pre>`)
	preBlockRe    = regexp.MustCompile(`(?s)<pre>(.*?)</pre>`)
	brRe          = regexp.MustCompile(`(?i)<br\s*/?>`)
	blockEndRe    = regexp.MustCompile(`(?i)</(p|div|li|tr|h[1-6]|pre)>`)
	tagRe         = regexp.MustCompile(`<[^>]+>`)
)

type problemDetailResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		Question struct {
			QuestionID     string `json:"questionId"`
			FrontendID     string `json:"questionFrontendId"`
			Title          string `json:"title"`
			TitleSlug      string `json:"titleSlug"`
			Difficulty     string `json:"difficulty"`
			SampleTestCase string `json:"sampleTestCase"`
			MetaData       string `json:"metaData"`
			Content        string `json:"content"`
			CodeSnippets   []struct {
				Lang     string `json:"lang"`
				LangSlug string `json:"langSlug"`
				Code     string `json:"code"`
			} `json:"codeSnippets"`
		} `json:"question"`
	} `json:"data"`
}

type metaData struct {
	Name   string `json:"name"`
	Params []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"params"`
	Return struct {
		Type string `json:"type"`
	} `json:"return"`
}

func (s *ProfileService) GetProblemDetail(ctx context.Context, slug string) (domain.ProblemDetail, error) {
	raw, err := s.client.FetchProblemDetail(ctx, slug)
	if err != nil {
		return domain.ProblemDetail{}, err
	}

	var resp problemDetailResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return domain.ProblemDetail{}, err
	}

	if len(resp.Errors) > 0 {
		return domain.ProblemDetail{}, errors.New(resp.Errors[0].Message)
	}

	q := resp.Data.Question
	if q.TitleSlug == "" {
		return domain.ProblemDetail{}, errors.New("problem not found: " + slug)
	}

	detail := domain.ProblemDetail{
		Slug:        q.TitleSlug,
		Title:       q.Title,
		Difficulty:  q.Difficulty,
		SampleInput: q.SampleTestCase,
		ContentHTML: q.Content,
		Signature:   parseSignature(q.MetaData),
	}

	detail.ID, _ = strconv.Atoi(q.FrontendID)

	for _, sn := range q.CodeSnippets {
		detail.Snippets = append(detail.Snippets, domain.CodeSnippet{
			Lang:     sn.Lang,
			LangSlug: sn.LangSlug,
			Code:     sn.Code,
		})
	}

	detail.Examples = parseExamples(q.Content)

	return detail, nil
}

func parseSignature(raw string) domain.Signature {
	var meta metaData
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return domain.Signature{}
	}

	sig := domain.Signature{
		Name:   meta.Name,
		Return: meta.Return.Type,
	}
	for _, p := range meta.Params {
		sig.Params = append(sig.Params, domain.Param{Name: p.Name, Type: p.Type})
	}
	return sig
}

func parseExamples(content string) []domain.Example {
	var examples []domain.Example

	if blocks := exampleHeadRe.FindAllStringSubmatch(content, -1); len(blocks) > 0 {
		for _, b := range blocks {
			num, _ := strconv.Atoi(b[1])
			ex, ok := parseExampleText(stripHTML(b[2]))
			if ok {
				ex.Number = num
				examples = append(examples, ex)
			}
		}
		return examples
	}

	num := 0
	for _, b := range preBlockRe.FindAllStringSubmatch(content, -1) {
		text := stripHTML(b[1])
		if !strings.Contains(text, "Input:") {
			continue
		}
		num++
		ex, ok := parseExampleText(text)
		if ok {
			ex.Number = num
			examples = append(examples, ex)
		}
	}

	return examples
}

func parseExampleText(text string) (domain.Example, bool) {
	var (
		ex   domain.Example
		mode int
		in   []string
		out  []string
		exp  []string
	)

	for _, line := range strings.Split(text, "\n") {
		l := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(l, "Input:"):
			mode = 1
			if v := strings.TrimSpace(strings.TrimPrefix(l, "Input:")); v != "" {
				in = append(in, v)
			}
		case strings.HasPrefix(l, "Output:"):
			mode = 2
			if v := strings.TrimSpace(strings.TrimPrefix(l, "Output:")); v != "" {
				out = append(out, v)
			}
		case strings.HasPrefix(l, "Explanation:"):
			mode = 3
			if v := strings.TrimSpace(strings.TrimPrefix(l, "Explanation:")); v != "" {
				exp = append(exp, v)
			}
		default:
			if l == "" {
				continue
			}
			switch mode {
			case 1:
				in = append(in, l)
			case 2:
				out = append(out, l)
			case 3:
				exp = append(exp, l)
			}
		}
	}

	ex.Input = strings.Join(in, "\n")
	ex.Output = strings.Join(out, "\n")
	ex.Explanation = strings.Join(exp, "\n")

	return ex, ex.Input != "" || ex.Output != ""
}

func stripHTML(s string) string {
	s = brRe.ReplaceAllString(s, "\n")
	s = blockEndRe.ReplaceAllString(s, "\n")
	s = tagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)

	var out []string
	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n")
}
