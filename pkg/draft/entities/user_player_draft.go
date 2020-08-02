package entities

type UserPlayerPreference struct {
	LeagueKey  string   `json:"league_key" bson:"league_key"`
	GameID     string   `json:"game_id" bson:"game_id"`
	UserID     string   `json:"user_guid" bson:"user_guid"`
	DoNotDraft []string `json:"dnd" bson:"dnd"`
	Preference []string `json:"pref" bson:"pref"`
	Available  []string `json:"ap" bson:"ap"`

	Positions map[string][]string `json:"positions" bson:"positions"`
}
