package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
)

type User struct {
	Email          string `json:"email" bson:"_id"`
	Name           string `json:"name"`
	ManagerID      string `json:"manager_id"`
	Nickname       string `json:"nickname"`
	Guid           string `json:"guid"`
	Uid            string `json:"uid"`
	IsCommissioner bool   `json:"is_commissioner"`
	IsCurrentLogin bool   `json:"is_current_login"`
	ImageURL       string `json:"image_url"`
	Teams          []Team
	Commissioned   []primitive.ObjectID
}

type LeagueGroup struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	FirstLeagueID int                `bson:"first_league_id" json:"first_league_id"`
	Leagues       []League           `json:"leagues" bson:"leagues"`
}

type League struct {
	LeagueKey      string             `bson:"league_key" json:"id"`
	Name           string             `bson:"name" json:"name"`
	LeagueID       int                `bson:"league_id" json:"league_id"`
	LeagueGroup    primitive.ObjectID `bson:"league_group_id" json:"league_group"`
	PreviousLeague *string            `bson:"previous_league_id" json:"previous_league_id"`
	Settings       *LeagueSettings    `bson:"settings" json:"settings" ` // year in which settings are for
	Teams          []Team             `bson:"teams" json:"teams"`        // season is the key
	Game           Game               `json:"game" bson:"game"`          // season is the key
	DraftOrder     []string           `json:"-" bson:"draft_order,omitempty"`
	TeamDraftOrder []Team             `json:"draft_order,omitempty" bson:"-"`
	DraftStarted   bool               `json:"draft_started" bson:"draft_started,omitempty"`
	DraftedCheck   []string           `json:"drafted_check" bson:"draft_check"`
}

func (l *League) GetParentID() *int {

	ids := strings.Split(l.Settings.Renew, "_")
	if len(ids) > 1 {
		parentLeagueID, _ := strconv.Atoi(ids[1])
		return &parentLeagueID
	}
	return nil

}

type LeagueSettings struct {
	LeagueKey                  string           `json:"league_key" json:"league_key"`
	LeagueID                   int              `json:"league_id" json:"league_id"`
	Name                       string           `json:"name"`
	URL                        string           `json:"url"`
	LogoURL                    string           `json:"logo_url"`
	Password                   string           `json:"password"`
	DraftStatus                string           `json:"draft_status"`
	NumTeams                   int              `json:"num_teams"`
	EditKey                    string           `json:"edit_key"`
	WeeklyDeadline             string           `json:"weekly_deadline"`
	LeagueUpdateTimestamp      string           `json:"league_update_timestamp"`
	LeagueType                 string           `json:"league_type"`
	Renew                      string           `json:"renew"`
	Renewed                    string           `json:"renewed"`
	IrisGroupChatID            string           `json:"iris_group_chat_id"`
	ShortInvitationURL         string           `json:"short_invitation_url"`
	AllowAddToDlExtraPos       string           `json:"allow_add_to_dl_extra_pos"`
	IsProLeague                string           `json:"is_pro_league"`
	IsCashLeague               string           `json:"is_cash_league"`
	CurrentWeek                string           `json:"current_week"`
	StartWeek                  string           `json:"start_week"`
	StartDate                  string           `json:"start_date"`
	EndWeek                    string           `json:"end_week"`
	EndDate                    string           `json:"end_date"`
	GameCode                   string           `json:"game_code"`
	Season                     string           `json:"season"`
	ID                         int              `json:"id"`
	DraftType                  string           `json:"draft_type"`
	IsAuctionDraft             bool             `json:"is_auction_draft"`
	ScoringType                string           `json:"scoring_type"`
	PersistentURL              string           `json:"persistent_url"`
	UsesPlayoff                string           `json:"uses_playoff"`
	HasPlayoffConsolationGames bool             `json:"has_playoff_consolation_games"`
	PlayoffStartWeek           string           `json:"playoff_start_week"`
	UsesPlayoffReseeding       bool             `json:"uses_playoff_reseeding"`
	UsesLockEliminatedTeams    bool             `json:"uses_lock_eliminated_teams"`
	NumPlayoffTeams            int              `json:"num_playoff_teams"`
	NumPlayoffConsolationTeams int              `json:"num_playoff_consolation_teams"`
	UsesRosterImport           bool             `json:"uses_roster_import"`
	RosterImportDeadline       string           `json:"roster_import_deadline"`
	WaiverType                 string           `json:"waiver_type"`
	WaiverRule                 string           `json:"waiver_rule"`
	UsesFaab                   bool             `json:"uses_faab"`
	DraftTime                  string           `json:"draft_time"`
	PostDraftPlayers           string           `json:"post_draft_players"`
	MaxTeams                   string           `json:"max_teams"`
	WaiverTime                 string           `json:"waiver_time"`
	TradeEndDate               string           `json:"trade_end_date"`
	TradeRatifyType            string           `json:"trade_ratify_type"`
	TradeRejectTime            string           `json:"trade_reject_time"`
	PlayerPool                 string           `json:"player_pool"`
	CantCutList                string           `json:"cant_cut_list"`
	IsPubliclyViewable         bool             `json:"is_publicly_viewable"`
	RosterPositions            []RosterPosition `json:"roster_positions"`
	StatCategories             []StatCategory   `json:"stat_categories"`
	StatModifiers              []StatModifier   `json:"stat_modifiers"`
	MaxAdds                    int              `json:"max_adds"`
	SeasonType                 string           `json:"season_type"`
	MinInningsPitched          string           `json:"min_innings_pitched"`
	UsesFractalPoints          bool             `json:"uses_fractal_points"`
	UsesNegativePoints         bool             `json:"uses_negative_points"`
}

type RosterPosition struct {
	Position     string `json:"position"`
	PositionType string `json:"position_type,omitempty"`
	Count        int    `json:"count"`
}

type PositionType struct {
	PositionType      string `json:"position_type"`
	IsOnlyDisplayStat bool   `json:"is_only_display_stat"`
}
type StatCategory struct {
	StatID            int            `json:"stat_id"`
	Enabled           bool           `json:"enabled"`
	Name              string         `json:"name"`
	DisplayName       string         `json:"display_name"`
	SortOrder         int            `json:"sort_order"`
	PositionType      string         `json:"position_type"`
	StatPositionTypes []PositionType `json:"stat_position_types"`
	IsOnlyDisplayStat string         `json:"is_only_display_stat,omitempty"`
}

type StatModifier struct {
	StatID  int     `json:"stat_id"`
	Value   float32 `json:"value"`
	Bonuses *Bonus  `json:"bonuses,omitempty"`
}

type Bonus struct {
	Target float32
	Points float32
}

type LeagueStandings struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   string `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs,omitempty"`
	Managers         []User `json:"managers"`
}

type Game struct {
	GameID             int    `json:"game_id" bson:"game_id"`
	GameKey            string `json:"game_key" bson:"game_key"`
	Name               string `json:"name"`
	Code               string `json:"code"`
	Type               string `json:"type"`
	URL                string `json:"url"`
	Season             int    `json:"season"`
	IsRegistrationOver bool   `json:"is_registration_over"  bson:"is_registration_over"`
	IsGameOver         bool   `json:"is_game_over" bson:"is_game_over"`
	IsOffseason        bool   `json:"is_offseason" bson:"is_offseason"`
}

type TeamLogo struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

type RosterAdds struct {
	CoverageType  string `json:"coverage_type"`
	CoverageValue int    `json:"coverage_value"`
	Value         int    `json:"value"`
}

type Team struct {
	TeamKey               string     `json:"team_key" bson:"team_key"`
	TeamID                int        `json:"team_id"  bson:"team_id"`
	Name                  string     `json:"name" `
	IsOwnedByCurrentLogin bool       `json:"is_owned_by_current_login" bson:"is_owned_by_current_login"`
	URL                   string     `json:"url"`
	TeamLogo              TeamLogo   `json:"team_logo" bson:"team_logo"`
	WaiverPriority        int        `json:"waiver_priority" bson:"waiver_priority"`
	NumberOfMoves         int        `json:"number_of_moves" bson:"number_of_moves"`
	NumberOfTrades        int        `json:"number_of_trades" bson:"number_of_trades"`
	RosterAdds            RosterAdds `json:"roster_adds" bson:"roster_adds"`
	LeagueScoringType     string     `json:"league_scoring_type" bson:"league_scroting_Type"`
	HasDraftGrade         bool       `json:"has_draft_grade" bson:"has_draft_grade"`
	DraftGrade            string     `json:"draft_grade" bson:"draft_grade"`
	Standing              Standings  `json:"standing" `
	Manager               []User     `json:"managers"`
}

type Outcome struct {
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`
	Ties       int `json:"ties"`
	Percentage int `json:"percentage"`
}

type Standings struct {
	Rank          int     `json:"rank"`
	PlayoffSeed   int     `json:"playoff_seed" bson:"playoff_seed"`
	OutcomeTotals Outcome `json:"outcome_totals" bson:"outcome_totals"`
	GamesBack     string  `json:"games_back" bson:"games_back"`
	PointsFor     float32 `json:"points_for" bson:"points_for"`
	PointsAgainst float32 `json:"points_against" bson:"points_against"`
}
