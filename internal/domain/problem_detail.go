package domain

type CodeSnippet struct {
	Lang     string
	LangSlug string
	Code     string
}

type Param struct {
	Name string
	Type string
}

type Signature struct {
	Name   string
	Params []Param
	Return string
}

type Example struct {
	Number      int
	Input       string
	Output      string
	Explanation string
}

type ProblemDetail struct {
	ID          int
	Slug        string
	Title       string
	Difficulty  string
	Signature   Signature
	Snippets    []CodeSnippet
	SampleInput string
	Examples    []Example
	ContentHTML string
}

func (d ProblemDetail) Snippet(langSlug string) (CodeSnippet, bool) {
	for _, s := range d.Snippets {
		if s.LangSlug == langSlug {
			return s, true
		}
	}
	return CodeSnippet{}, false
}
