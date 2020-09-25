package entities

import "github.com/thethan/fdr-users/pkg/draft/entities"

type ImportPlayerStat struct {
	Guid      string `json:"guid"`
	PlayerKey string `json:"player_key"`
	Week      string `json:"week"`
	Season    string `json:"season"`
	// deprecated
	PlayerStats []entities.PlayerStat `json:"player_stats"`
}

type ImportPlayer struct {
	Guid   string `json:"guid"`
	GameID int    `json:"game_id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	// deprecated
}
