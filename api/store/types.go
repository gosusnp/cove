package store

type Exercise struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Progression *string `json:"progression"`
}

type Program struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProgramSet struct {
	ID                  int64   `json:"id"`
	ProgramID           int64   `json:"program_id"`
	Name                *string `json:"name"`
	Rounds              int     `json:"rounds"`
	IntraSetRestSeconds *int    `json:"intra_set_rest_seconds"`
	SortOrder           *int    `json:"sort_order"`
}
