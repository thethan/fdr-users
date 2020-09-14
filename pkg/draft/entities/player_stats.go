package entities

type PlayerStats struct {
	PlayerID string `json:"player_key" bson:"player_key"`

	Season map[int]map[int][]PlayerStat `json:"stats_by_season" bson:"stats_by_season"`
}
