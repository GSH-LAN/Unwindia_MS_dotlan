package config

type DotlanConfig struct {
	UserId           uint   `json:"dotlanUserId"`
	TournamentFilter string `json:"dotlanTournamentFilter"`
	DefaultGame      string `json:"defaultGame"`
}
