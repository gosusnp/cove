package store

type Exercise struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Progression *string `json:"progression"`
}
