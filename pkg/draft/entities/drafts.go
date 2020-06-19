package entities

import "time"

type Draft struct {
	ID           string        `json:"id" bson:"_id"`
	League       League        `json:"league" bson:"leagues"`
	DraftResults []DraftResult `json:"draft_results" bson:"draft_results"`
}

//
//type DraftResult struct {
//	PlayerKey    PlayerSeason `json:"player" bson:"players"`
//	LeagueKey string       `json:"league_key" bson:"-"`
//	Cost      string       `json:"cost" bson:"cost"`
//	TeamKey      TeamKey        `json:"team" bson:"team"`
//	Round     int          `json:"round" bson:"round"`
//	Pick  int          `json:"position" bson:"position"`
//	Timestamp time.Time    `json:"timestamp" bson:"timestamp"`
//}

type DraftResult struct {
	UserGUID  string    `json:"user_guid" bson:"user_guid"`
	PlayerKey string    `json:"player_key" bson:"player_key"`
	PlayerID  int       `json:"player_key" bson:"player_id"`
	LeagueKey string    `json:"league_key" bson:"league_key"`
	TeamKey   string    `json:"team_key" bson:"team_key"`
	Round     int       `json:"round" bson:"round"`
	Pick      int       `json:"pick" bson:"pick"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	GameID    int       `json:"game_id" bson:"game_id"`

	Player    []*PlayerSeason `json:"player" bson:"player,omitempty"`
	//League    League          `json:"league" bson:"leagues,omitempty"`
}
