package entities

type Message struct {
	Guid      string `json:"guid"`
	PlayerKey string `json:"player_key"`
	Week      int    `json:"week"`
}
