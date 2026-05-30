package domain

type Difficulty struct {
	Easy   int
	Medium int
	Hard   int
	Total  int
}

type LeetCodeTotals struct {
	Easy   int
	Medium int
	Hard   int
}

type Language struct {
	Name  string
	Count int
}

type Skill struct {
	Name  string
	Count int
}

type Profile struct {
	Username   string
	Rank       int
	Reputation int
	Streak     int

	Difficulty     Difficulty
	LeetCodeTotals LeetCodeTotals
	Languages      []Language
	Skills         []Skill
}
