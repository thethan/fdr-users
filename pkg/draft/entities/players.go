package entities

type PlayerSeason struct {
	PlayerKey             string `json:"player_key" bson:"_id"`
	PlayerID              int    `json:"player_id"`
	Name                  PlayerName
	EditorialPlayerKey    string          `json:"editorial_player_key"`
	EditorialTeamKey      string          `json:"editorial_team_key"`
	EditorialTeamFullName string          `json:"editorial_team_full_name"`
	EditorialTeamAbbr     string          `json:"editorial_team_abbr"`
	ByeWeeks              PlayerByeWeeks `json:"bye_weeks"`
	UniformNumber         string          `json:"uniform_number"`
	DisplayPosition       string          `json:"display_position"`
	Headshot              PlayerHeadshot  `json:"headshot"`
	ImageURL              string          `json:"image_url"`
	IsUndroppable         bool          `json:"is_undroppable"`
	PositionType          string          `json:"position_type"`
	EligiblePositions     []string        `json:"eligible_positions"`

	Ranks       PlayerRanks `json:"ranks" bson:"ranks"`
	SeasonStats []PlayerStat `json:"season_stats" bson:"season_stats"`
	WeeklyStats SeasonWeekStats `json:"weekly_stats" bson:"weekly_stats"`
}

type PlayerRanks map[string]int

// sesaon -> week -> stats
type SeasonWeekStats map[int][]PlayerStat

type PlayerName struct {
	Full       string `json:"full"`
	First      string `json:"first"`
	Last       string `json:"last"`
	AsciiFirst string `json:"ascii_first"`
	AsciiLast  string `json:"ascii_last"`
}

type PlayerHeadshot struct {
	URL  string `json:"url"`
	Size string `json:"size"`
}

type PlayerByeWeeks struct {
	Week int `json:"week"`
}

type PlayerStat struct {
	StatID int `json:"stat_id" bson:"stat_id"`
	Value  float32 `json:"value" bson:"value"`
}
