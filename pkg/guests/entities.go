package guests

import "time"

type Guest struct {
	Name             string    `json:"name" bson:"name"`
	Adults           int       `json:"adults"  bson:"adults"`
	Children         int       `json:"children"  bson:"children"`
	Email            string    `json:"email" bson:"email"`
	Attending        bool      `json:"attending" bson:"attending"`
	VeganOptionCount int       `json:"vegan_option_count" bson:"vegan_option_count"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
}
