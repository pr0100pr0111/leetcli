package domain

type Topic struct {
	Name string
	Slug string
}

type DailyChallenge struct {
	Date       string
	Slug       string
	ID         int
	Title      string
	Difficulty string
	Topics     []Topic
}
