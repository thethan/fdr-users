package yahoo

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/users"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type UserInformation interface {
	GetCredentialInformation(ctx context.Context, session string) (users.User, error)
}

type Service struct {
	logger   log.Logger
	userRepo UserInformation
	client   http.Client
	session  string
}

type ServiceOptions func()

func NewService(logger log.Logger, information UserInformation, ) *Service {
	svc := Service{logger: logger, userRepo: information, client: http.Client{
		Timeout: 5 * time.Second,
	}}
	return &svc
}

func (s *Service) WithSession(session string) *Service {
	s.session = session
	return s
}

// Get adheres to goth fantasy
func (s *Service) Get(url string) (response *http.Response, err error) {
	return s.get(context.Background(), url)
}

func (s *Service) get(ctx context.Context, url string) (response *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// get user information
	user, err := s.userRepo.GetCredentialInformation(ctx, s.session)
	if err != nil {
		return nil, err
	}
	// add authentication credentials
	req.Header.Set("Authorization", "Bearer "+user.AccessToken)

	return s.client.Do(req)
}

//

type GameResourceMetaResponse struct {
	XMLName     xml.Name  `xml:"fantasy_content"`
	Text        string    `xml:",chardata"`
	Lang        string    `xml:"lang,attr"`
	URI         string    `xml:"uri,attr"`
	Time        string    `xml:"time,attr"`
	Copyright   string    `xml:"copyright,attr"`
	RefreshRate string    `xml:"refresh_rate,attr"`
	Yahoo       string    `xml:"yahoo,attr"`
	Xmlns       string    `xml:"xmlns,attr"`
	Game        YahooGame `xml:"game"`
}

// GameResourcesMeta
func (s *Service) GetGameResourcesMeta(gameKey string) (*Game, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/metadata", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	v := GameResourceMetaResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// transform response to games
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}

	game := transformYahooResponseGameToGame(v.Game)

	return &game, nil

}

type GameResourceLeagues struct {
	GameKey string   `json:"game_key"`
	GameID  string   `json:"game_id"`
	Name    string   `json:"name"`
	Code    string   `json:"code"`
	Type    string   `json:"type"`
	URL     string   `json:"url"`
	Season  string   `json:"season"`
	Leagues []League `json:"leagues"`
}
type League struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              string `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    string `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
}
type GameResourceLeaguesResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Game        struct {
		Text               string `xml:",chardata"`
		GameKey            string `xml:"game_key"`
		GameID             string `xml:"game_id"`
		Name               string `xml:"name"`
		Code               string `xml:"code"`
		Type               string `xml:"type"`
		URL                string `xml:"url"`
		Season             string `xml:"season"`
		IsRegistrationOver string `xml:"is_registration_over"`
		IsGameOver         string `xml:"is_game_over"`
		IsOffseason        string `xml:"is_offseason"`
		Leagues            struct {
			Text   string        `xml:",chardata"`
			Count  string        `xml:"count,attr"`
			League []YahooLeague `xml:"league"`
		} `xml:"leagues"`
	} `xml:"game"`
}

type YahooLeague struct {
	Text                  string `xml:",chardata"`
	LeagueKey             string `xml:"league_key"`
	LeagueID              string `xml:"league_id"`
	Name                  string `xml:"name"`
	URL                   string `xml:"url"`
	LogoURL               string `xml:"logo_url"`
	DraftStatus           string `xml:"draft_status"`
	NumTeams              string `xml:"num_teams"`
	EditKey               string `xml:"edit_key"`
	WeeklyDeadline        string `xml:"weekly_deadline"`
	LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
	ScoringType           string `xml:"scoring_type"`
	LeagueType            string `xml:"league_type"`
	Renew                 string `xml:"renew"`
	Renewed               string `xml:"renewed"`
	IrisGroupChatID       string `xml:"iris_group_chat_id"`
	AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
	IsProLeague           string `xml:"is_pro_league"`
	IsCashLeague          string `xml:"is_cash_league"`
	CurrentWeek           string `xml:"current_week"`
	StartWeek             string `xml:"start_week"`
	StartDate             string `xml:"start_date"`
	EndWeek               string `xml:"end_week"`
	EndDate               string `xml:"end_date"`
	IsFinished            int    `xml:"is_finished"`
	GameCode              string `xml:"game_code"`
	Season                string `xml:"season"`
	Password              string `xml:"password"`
	ShortInvitationURL    string `xml:"short_invitation_url"`
}

// GetGameResourcesLeagues
func (s *Service) GetGameResourcesLeagues(gameKey string, leagueKeys []string) (*GameResourceLeagues, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/leagues;league_keys=%s", gameKey, strings.Join(leagueKeys, ","))
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	v := GameResourceLeaguesResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// transform response to games
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}

	gLeagues := GameResourceLeagues{
		GameKey: v.Game.GameKey,
		GameID:  v.Game.GameID,
		Name:    v.Game.Name,
		Code:    v.Game.Code,
		Type:    v.Game.Type,
		URL:     v.Game.URL,
		Season:  v.Game.Season,
	}

	leagues := make([]League, len(v.Game.Leagues.League))

	for idx, yahooLeague := range v.Game.Leagues.League {
		leagues[idx] = League{
			LeagueKey:   yahooLeague.LeagueKey,
			LeagueID:    yahooLeague.LeagueID,
			Name:        yahooLeague.Name,
			URL:         yahooLeague.URL,
			DraftStatus: yahooLeague.DraftStatus,
			//NumTeams:              yahooLeague.NumTeams,
			EditKey:               yahooLeague.EditKey,
			WeeklyDeadline:        yahooLeague.WeeklyDeadline,
			LeagueUpdateTimestamp: yahooLeague.LeagueUpdateTimestamp,
			ScoringType:           yahooLeague.ScoringType,
			LeagueType:            yahooLeague.LeagueType,
			Renew:                 yahooLeague.Renew,
			Renewed:               yahooLeague.Renewed,
			ShortInvitationURL:    yahooLeague.ShortInvitationURL,
			IsProLeague:           yahooLeague.IsProLeague,
			CurrentWeek:           yahooLeague.CurrentWeek,
			StartWeek:             yahooLeague.StartWeek,
			StartDate:             yahooLeague.StartDate,
			EndWeek:               yahooLeague.EndWeek,
			EndDate:               yahooLeague.EndDate,
			IsFinished:            yahooLeague.IsFinished,
		}
	}
	gLeagues.Leagues = leagues

	return &gLeagues, nil
}

type GameResourcePlayer struct {
	GameKey string `json:"game_key"`
	GameID  string `json:"game_id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Season  string `json:"season"`
	Players []struct {
		PlayerKey string `json:"player_key"`
		PlayerID  string `json:"player_id"`
		Name      struct {
			Full       string `json:"full"`
			First      string `json:"first"`
			Last       string `json:"last"`
			ASCIIFirst string `json:"ascii_first"`
			ASCIILast  string `json:"ascii_last"`
		} `json:"name"`
		EditorialPlayerKey    string   `json:"editorial_player_key"`
		EditorialTeamKey      string   `json:"editorial_team_key"`
		EditorialTeamFullName string   `json:"editorial_team_full_name"`
		EditorialTeamAbbr     string   `json:"editorial_team_abbr"`
		UniformNumber         string   `json:"uniform_number"`
		DisplayPosition       string   `json:"display_position"`
		Headshot              string   `json:"headshot"`
		IsUndroppable         string   `json:"is_undroppable"`
		PositionType          string   `json:"position_type"`
		EligiblePositions     []string `json:"eligible_positions"`
	} `json:"players"`
}
type GameResourcePlayerResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Game        struct {
		Text               string `xml:",chardata"`
		GameKey            string `xml:"game_key"`
		GameID             string `xml:"game_id"`
		Name               string `xml:"name"`
		Code               string `xml:"code"`
		Type               string `xml:"type"`
		URL                string `xml:"url"`
		Season             string `xml:"season"`
		IsRegistrationOver string `xml:"is_registration_over"`
		IsGameOver         string `xml:"is_game_over"`
		IsOffseason        string `xml:"is_offseason"`
		Players            struct {
			Text   string `xml:",chardata"`
			Count  string `xml:"count,attr"`
			Player []struct {
				Text      string `xml:",chardata"`
				PlayerKey string `xml:"player_key"`
				PlayerID  string `xml:"player_id"`
				Name      struct {
					Text       string `xml:",chardata"`
					Full       string `xml:"full"`
					First      string `xml:"first"`
					Last       string `xml:"last"`
					AsciiFirst string `xml:"ascii_first"`
					AsciiLast  string `xml:"ascii_last"`
				} `xml:"name"`
				EditorialPlayerKey    string `xml:"editorial_player_key"`
				EditorialTeamKey      string `xml:"editorial_team_key"`
				EditorialTeamFullName string `xml:"editorial_team_full_name"`
				EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
				ByeWeeks              struct {
					Text string `xml:",chardata"`
					Week string `xml:"week"`
				} `xml:"bye_weeks"`
				UniformNumber   string `xml:"uniform_number"`
				DisplayPosition string `xml:"display_position"`
				Headshot        struct {
					Text string `xml:",chardata"`
					URL  string `xml:"url"`
					Size string `xml:"size"`
				} `xml:"headshot"`
				ImageURL          string `xml:"image_url"`
				IsUndroppable     string `xml:"is_undroppable"`
				PositionType      string `xml:"position_type"`
				EligiblePositions struct {
					Text     string `xml:",chardata"`
					Position string `xml:"position"`
				} `xml:"eligible_positions"`
				HasPlayerNotes           string `xml:"has_player_notes"`
				PlayerNotesLastTimestamp string `xml:"player_notes_last_timestamp"`
				Status                   string `xml:"status"`
				StatusFull               string `xml:"status_full"`
			} `xml:"player"`
		} `xml:"players"`
	} `xml:"game"`
}

// GetGameResourcesLeagues
func (s *Service) GetGameResourcesPlayers(gameKey string) (*GameResourcePlayer, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/players", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")

}

type GameResourcesGameWeeks struct {
	GameKey string `json:"game_key"`
	GameID  string `json:"game_id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	Season  string `json:"season"`
	Weeks   []struct {
		Week  string `json:"week"`
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"weeks"`
}
type GameResourcesGameWeeksResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Game        struct {
		Text               string `xml:",chardata"`
		GameKey            string `xml:"game_key"`
		GameID             string `xml:"game_id"`
		Name               string `xml:"name"`
		Code               string `xml:"code"`
		Type               string `xml:"type"`
		URL                string `xml:"url"`
		Season             string `xml:"season"`
		IsRegistrationOver string `xml:"is_registration_over"`
		IsGameOver         string `xml:"is_game_over"`
		IsOffseason        string `xml:"is_offseason"`
		GameWeeks          struct {
			Text     string `xml:",chardata"`
			Count    string `xml:"count,attr"`
			GameWeek []struct {
				Text        string `xml:",chardata"`
				Week        string `xml:"week"`
				DisplayName string `xml:"display_name"`
				Start       string `xml:"start"`
				End         string `xml:"end"`
			} `xml:"game_week"`
		} `xml:"game_weeks"`
	} `xml:"game"`
}

// GetGameResourcesGameWeeks
func (s *Service) GetGameResourcesGameWeeks(gameKey string) (*GameResourcesGameWeeks, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/game_weeks", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")

}

type GameResourcesStatCategories struct {
	GameKey        string `json:"game_key"`
	GameID         string `json:"game_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Type           string `json:"type"`
	URL            string `json:"url"`
	Season         string `json:"season"`
	StatCategories []struct {
		StatID          int      `json:"stat_id"`
		Name            string   `json:"name"`
		DisplayName     string   `json:"display_name"`
		SortOrder       string   `json:"sort_order"`
		PositionTypes   []string `json:"position_types,omitempty"`
		IsCompositeStat int      `json:"is_composite_stat,omitempty"`
		BaseStats       []string `json:"base_stats,omitempty"`
	} `json:"stat_categories"`
}
type GameResourcesStatCategoriesResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Game        struct {
		Text               string `xml:",chardata"`
		GameKey            string `xml:"game_key"`
		GameID             string `xml:"game_id"`
		Name               string `xml:"name"`
		Code               string `xml:"code"`
		Type               string `xml:"type"`
		URL                string `xml:"url"`
		Season             string `xml:"season"`
		IsRegistrationOver string `xml:"is_registration_over"`
		IsGameOver         string `xml:"is_game_over"`
		IsOffseason        string `xml:"is_offseason"`
		StatCategories     struct {
			Text  string `xml:",chardata"`
			Stats struct {
				Text string `xml:",chardata"`
				Stat []struct {
					Text          string `xml:",chardata"`
					StatID        string `xml:"stat_id"`
					Name          string `xml:"name"`
					DisplayName   string `xml:"display_name"`
					SortOrder     string `xml:"sort_order"`
					PositionTypes struct {
						Text         string   `xml:",chardata"`
						PositionType []string `xml:"position_type"`
					} `xml:"position_types"`
				} `xml:"stat"`
			} `xml:"stats"`
		} `xml:"stat_categories"`
	} `xml:"game"`
}

// GetGameResourcesGameWeeks
func (s *Service) GetGameResourcesStatCategories(gameKey string) (*GameResourcesStatCategories, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/stat_categories", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")

}

type GameResourcesPositionTypes struct {
	GameKey       string `json:"game_key"`
	GameID        string `json:"game_id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	Season        string `json:"season"`
	PositionTypes []struct {
		Type        string `json:"type"`
		DisplayName string `json:"display_name"`
	} `json:"position_types"`
}

// GetGameResourcesGameWeeks
func (s *Service) GetGameResourcesPositionTypes(gameKey string) (*GameResourcesPositionTypes, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/position_types", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	v := UserResourcesResponse{}
	// transform response to games
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}

	return nil, errors.New("not implemented")

}

type GameResourcesRosterPositions struct {
	GameKey         string `json:"game_key"`
	GameID          string `json:"game_id"`
	Name            string `json:"name"`
	Code            string `json:"code"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	Season          string `json:"season"`
	RosterPositions []struct {
		Position       string `json:"position"`
		Abbreviation   string `json:"abbreviation"`
		DisplayName    string `json:"display_name"`
		PositionType   string `json:"position_type,omitempty"`
		IsBench        int    `json:"is_bench,omitempty"`
		IsDisabledList int    `json:"is_disabled_list,omitempty"`
	} `json:"roster_positions"`
}
type GetGameResourcesRosterPositionsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Game        struct {
		Text               string `xml:",chardata"`
		GameKey            string `xml:"game_key"`
		GameID             string `xml:"game_id"`
		Name               string `xml:"name"`
		Code               string `xml:"code"`
		Type               string `xml:"type"`
		URL                string `xml:"url"`
		Season             string `xml:"season"`
		IsRegistrationOver string `xml:"is_registration_over"`
		IsGameOver         string `xml:"is_game_over"`
		IsOffseason        string `xml:"is_offseason"`
		PositionTypes      struct {
			Text         string `xml:",chardata"`
			PositionType []struct {
				Text        string `xml:",chardata"`
				Type        string `xml:"type"`
				DisplayName string `xml:"display_name"`
			} `xml:"position_type"`
		} `xml:"position_types"`
	} `xml:"game"`
}

func (s *Service) GetGameResourcesRosterPositions(gameKey string) (*GameResourcesRosterPositions, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%s/roster_positions", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")

}

/// User
func transformYahooResponseGameToGame(yahooGame YahooGame) Game {
	return Game{
		GameKey: yahooGame.GameKey,
		GameID:  yahooGame.GameID,
		Name:    yahooGame.Name,
		Code:    yahooGame.Code,
		Type:    yahooGame.Type,
		URL:     yahooGame.URL,
		Season:  yahooGame.Season,
	}
}

type UserResourcesGames struct {
	GUID  string `json:"guid"`
	Games []Game
}
type Game struct {
	GameKey string      `json:"game_key"`
	GameID  string      `json:"game_id"`
	Name    string      `json:"name"`
	Code    string      `json:"code"`
	Type    string      `json:"type"`
	URL     interface{} `json:"url"`
	Season  string      `json:"season"`
}

type YahooGame struct {
	Text               string `xml:",chardata"`
	GameKey            string `xml:"game_key"`
	GameID             string `xml:"game_id"`
	Name               string `xml:"name"`
	Code               string `xml:"code"`
	Type               string `xml:"type"`
	URL                string `xml:"url"`
	Season             string `xml:"season"`
	IsRegistrationOver string `xml:"is_registration_over"`
	IsGameOver         string `xml:"is_game_over"`
	IsOffseason        string `xml:"is_offseason"`
	EditorialSeason    string `xml:"editorial_season"`
	PicksStatus        string `xml:"picks_status"`
	ContestGroupID     string `xml:"contest_group_id"`
	ScenarioGenerator  string `xml:"scenario_generator"`
	CurrentWeek        string `xml:"current_week"`
	IsContestRegActive string `xml:"is_contest_reg_active"`
	IsContestOver      string `xml:"is_contest_over"`
}
type UserResourcesResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Users       struct {
		Text  string `xml:",chardata"`
		Count string `xml:"count,attr"`
		User  struct {
			Text  string `xml:",chardata"`
			Guid  string `xml:"guid"`
			Games struct {
				Text  string      `xml:",chardata"`
				Count string      `xml:"count,attr"`
				Game  []YahooGame `xml:"game"`
			} `xml:"games"`
		} `xml:"user"`
	} `xml:"users"`
}

//GetUserResourcesRosterPositions
func (s *Service) GetUserResourcesGames(ctx context.Context, user users.User) (*UserResourcesGames, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games")
	res, err := s.get(ctx, url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	v := UserResourcesResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// transform response to games
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	games := make([]Game, len(v.Users.User.Games.Game))

	for idx, yahooGame := range v.Users.User.Games.Game {
		game := transformYahooResponseGameToGame(yahooGame)
		games[idx] = game
	}

	return &UserResourcesGames{
		GUID:  v.Users.User.Guid,
		Games: games,
	}, nil

}

type GameLeagues struct {
	GameKey string   `json:"game_key"`
	GameID  string   `json:"game_id"`
	Name    string   `json:"name"`
	Code    string   `json:"code"`
	Type    string   `json:"type"`
	URL     string   `json:"url"`
	Season  string   `json:"season"`
	Leagues []League `json:"leagues"`
}

type UserResourcesGameLeagues struct {
	GUID       string      `json:"guid"`
	GameLeague GameLeagues `json:"leagues"`
}
type UserResourcesGameLeaguesResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Users       struct {
		Text  string `xml:",chardata"`
		Count string `xml:"count,attr"`
		User  struct {
			Text  string `xml:",chardata"`
			Guid  string `xml:"guid"`
			Games struct {
				Text  string `xml:",chardata"`
				Count string `xml:"count,attr"`
				Game  struct {
					Text               string `xml:",chardata"`
					GameKey            string `xml:"game_key"`
					GameID             string `xml:"game_id"`
					Name               string `xml:"name"`
					Code               string `xml:"code"`
					Type               string `xml:"type"`
					URL                string `xml:"url"`
					Season             string `xml:"season"`
					IsRegistrationOver string `xml:"is_registration_over"`
					IsGameOver         string `xml:"is_game_over"`
					IsOffseason        string `xml:"is_offseason"`
					Leagues            struct {
						Text   string        `xml:",chardata"`
						Count  string        `xml:"count,attr"`
						League []YahooLeague `xml:"league"`
					} `xml:"leagues"`
				} `xml:"game"`
			} `xml:"games"`
		} `xml:"user"`
	} `xml:"users"`
}

//GetUserResourcesRosterPositions
func (s *Service) GetUserResourcesGameLeagues(ctx context.Context, gameKey string) (*UserResourcesGameLeagues, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games;game_keys=%s/leagues", gameKey)
	res, err := s.get(ctx, url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	v := UserResourcesGameLeaguesResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	leagues := make([]League, len(v.Users.User.Games.Game.Leagues.League))
	gameLeague := GameLeagues{
		GameKey: v.Users.User.Games.Game.GameKey,
		GameID:  v.Users.User.Games.Game.GameID,
		Name:    v.Users.User.Games.Game.Name,
		Code:    v.Users.User.Games.Game.Code,
		Type:    v.Users.User.Games.Game.Type,
		URL:     v.Users.User.Games.Game.URL,
		Season:  v.Users.User.Games.Game.Season,
	}
	for idx, yahooLeague := range v.Users.User.Games.Game.Leagues.League {
		league := League{
			LeagueKey:             yahooLeague.LeagueKey,
			LeagueID:              yahooLeague.LeagueID,
			Name:                  yahooLeague.Name,
			URL:                   yahooLeague.URL,
			DraftStatus:           yahooLeague.DraftStatus,
			NumTeams:              yahooLeague.NumTeams,
			EditKey:               yahooLeague.EditKey,
			WeeklyDeadline:        yahooLeague.WeeklyDeadline,
			LeagueUpdateTimestamp: yahooLeague.LeagueUpdateTimestamp,
			ScoringType:           yahooLeague.ScoringType,
			LeagueType:            yahooLeague.LeagueType,
			Renew:                 yahooLeague.Renew,
			Renewed:               yahooLeague.Renewed,
			ShortInvitationURL:    yahooLeague.ShortInvitationURL,
			IsProLeague:           yahooLeague.IsProLeague,
			CurrentWeek:           yahooLeague.CurrentWeek,
			StartWeek:             yahooLeague.StartWeek,
			StartDate:             yahooLeague.StartDate,
			EndWeek:               yahooLeague.EndWeek,
			EndDate:               yahooLeague.EndDate,
			IsFinished:            yahooLeague.IsFinished,
		}
		leagues[idx] = league
	}
	gameLeague.Leagues = leagues
	r := UserResourcesGameLeagues{
		GUID:       v.Users.User.Guid,
		GameLeague: gameLeague,
	}

	return &r, nil

}

type UserResourcesGameTeams struct {
	GUID  string `json:"guid"`
	Teams []struct {
		GameKey string `json:"game_key"`
		GameID  string `json:"game_id"`
		Name    string `json:"name"`
		Code    string `json:"code"`
		Type    string `json:"type"`
		URL     string `json:"url"`
		Season  string `json:"season"`
		Teams   []struct {
			TeamKey               string `json:"team_key"`
			TeamID                string `json:"team_id"`
			Name                  string `json:"name"`
			IsOwnedByCurrentLogin int    `json:"is_owned_by_current_login"`
			URL                   string `json:"url"`
			TeamLogo              string `json:"team_logo"`
			WaiverPriority        int    `json:"waiver_priority"`
			NumberOfMoves         string `json:"number_of_moves"`
			NumberOfTrades        int    `json:"number_of_trades"`
			Managers              []struct {
				ManagerID      string `json:"manager_id"`
				Nickname       string `json:"nickname"`
				GUID           string `json:"guid"`
				IsCurrentLogin string `json:"is_current_login"`
				Email          string `json:"email"`
				ImageURL       string `json:"image_url"`
			} `json:"managers"`
			ClinchedPlayoffs int `json:"clinched_playoffs,omitempty"`
		} `json:"teams"`
	} `json:"teams"`
}
type UserResourcesGameTeamsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Users       struct {
		Text  string `xml:",chardata"`
		Count string `xml:"count,attr"`
		User  struct {
			Text  string `xml:",chardata"`
			Guid  string `xml:"guid"`
			Games struct {
				Text  string `xml:",chardata"`
				Count string `xml:"count,attr"`
				Game  struct {
					Text               string `xml:",chardata"`
					GameKey            string `xml:"game_key"`
					GameID             string `xml:"game_id"`
					Name               string `xml:"name"`
					Code               string `xml:"code"`
					Type               string `xml:"type"`
					URL                string `xml:"url"`
					Season             string `xml:"season"`
					IsRegistrationOver string `xml:"is_registration_over"`
					IsGameOver         string `xml:"is_game_over"`
					IsOffseason        string `xml:"is_offseason"`
					Teams              struct {
						Text  string `xml:",chardata"`
						Count string `xml:"count,attr"`
						Team  struct {
							Text                  string `xml:",chardata"`
							TeamKey               string `xml:"team_key"`
							TeamID                string `xml:"team_id"`
							Name                  string `xml:"name"`
							IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
							URL                   string `xml:"url"`
							TeamLogos             struct {
								Text     string `xml:",chardata"`
								TeamLogo struct {
									Text string `xml:",chardata"`
									Size string `xml:"size"`
									URL  string `xml:"url"`
								} `xml:"team_logo"`
							} `xml:"team_logos"`
							WaiverPriority string `xml:"waiver_priority"`
							NumberOfMoves  string `xml:"number_of_moves"`
							NumberOfTrades string `xml:"number_of_trades"`
							RosterAdds     struct {
								Text          string `xml:",chardata"`
								CoverageType  string `xml:"coverage_type"`
								CoverageValue string `xml:"coverage_value"`
								Value         string `xml:"value"`
							} `xml:"roster_adds"`
							LeagueScoringType string `xml:"league_scoring_type"`
							HasDraftGrade     string `xml:"has_draft_grade"`
							Managers          struct {
								Text    string `xml:",chardata"`
								Manager struct {
									Text           string `xml:",chardata"`
									ManagerID      string `xml:"manager_id"`
									Nickname       string `xml:"nickname"`
									Guid           string `xml:"guid"`
									IsCommissioner string `xml:"is_commissioner"`
									IsCurrentLogin string `xml:"is_current_login"`
									Email          string `xml:"email"`
									ImageURL       string `xml:"image_url"`
								} `xml:"manager"`
							} `xml:"managers"`
						} `xml:"team"`
					} `xml:"teams"`
				} `xml:"game"`
			} `xml:"games"`
		} `xml:"user"`
	} `xml:"users"`
}

//GetUserResourcesRosterPositions
func (s *Service) GetUserResourcesGameTeams(gameKey string) (*UserResourcesGameTeams, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games;game_keys=%s/teams", gameKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")

}

/*
	League Resources
*/

type LeagueResourcesMeta struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              int    `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    string `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
}
type LeagueResourcesMetaResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesMeta(leagueKey string) (*LeagueResourcesMeta, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/metadata", leagueKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type PositionType struct {
	PositionType      string `json:"position_type"`
	IsOnlyDisplayStat string `json:"is_only_display_stat"`
}
type StatCategory struct {
	StatID            int            `json:"stat_id"`
	Enabled           string         `json:"enabled"`
	Name              string         `json:"name"`
	DisplayName       string         `json:"display_name"`
	SortOrder         string         `json:"sort_order"`
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
type LeagueResourcesSettings struct {
	DraftType                  string `json:"draft_type"`
	IsAuctionDraft             string `json:"is_auction_draft"`
	ScoringType                string `json:"scoring_type"`
	PersistentURL              string `json:"persistent_url"`
	UsesPlayoff                string `json:"uses_playoff"`
	HasPlayoffConsolationGames bool   `json:"has_playoff_consolation_games"`
	PlayoffStartWeek           string `json:"playoff_start_week"`
	UsesPlayoffReseeding       int    `json:"uses_playoff_reseeding"`
	UsesLockEliminatedTeams    int    `json:"uses_lock_eliminated_teams"`
	NumPlayoffTeams            string `json:"num_playoff_teams"`
	NumPlayoffConsolationTeams int    `json:"num_playoff_consolation_teams"`
	UsesRosterImport           string `json:"uses_roster_import"`
	RosterImportDeadline       string `json:"roster_import_deadline"`
	WaiverType                 string `json:"waiver_type"`
	WaiverRule                 string `json:"waiver_rule"`
	UsesFaab                   string `json:"uses_faab"`
	DraftTime                  string `json:"draft_time"`
	PostDraftPlayers           string `json:"post_draft_players"`
	MaxTeams                   string `json:"max_teams"`
	WaiverTime                 string `json:"waiver_time"`
	TradeEndDate               string `json:"trade_end_date"`
	TradeRatifyType            string `json:"trade_ratify_type"`
	TradeRejectTime            string `json:"trade_reject_time"`
	PlayerPool                 string `json:"player_pool"`
	CantCutList                string `json:"cant_cut_list"`
	IsPubliclyViewable         string `json:"is_publicly_viewable"`
	RosterPositions            []struct {
		Position     string `json:"position"`
		PositionType string `json:"position_type,omitempty"`
		Count        int    `json:"count"`
	} `json:"roster_positions"`
	StatCategories     []StatCategory `json:"stat_categories"`
	StatModifiers      []StatModifier
	MaxAdds            string `json:"max_adds"`
	SeasonType         string `json:"season_type"`
	MinInningsPitched  string `json:"min_innings_pitched"`
	UsesFractalPoints  bool   `json:"uses_fractal_points"`
	UsesNegativePoints bool   `json:"uses_negative_points"`
	League             struct {
		LeagueKey             string `json:"league_key"`
		LeagueID              string `json:"league_id"`
		Name                  string `json:"name"`
		URL                   string `json:"url"`
		DraftStatus           string `json:"draft_status"`
		NumTeams              int    `json:"num_teams"`
		EditKey               string `json:"edit_key"`
		WeeklyDeadline        string `json:"weekly_deadline"`
		LeagueUpdateTimestamp string `json:"league_update_timestamp"`
		ScoringType           string `json:"scoring_type"`
		LeagueType            string `json:"league_type"`
		Renew                 string `json:"renew"`
		Renewed               string `json:"renewed"`
		ShortInvitationURL    string `json:"short_invitation_url"`
		IsProLeague           string `json:"is_pro_league"`
		CurrentWeek           string `json:"current_week"`
		StartWeek             string `json:"start_week"`
		StartDate             string `json:"start_date"`
		EndWeek               string `json:"end_week"`
		EndDate               string `json:"end_date"`
		IsFinished            int    `json:"is_finished"`
	} `json:"league"`
}
type LeagueResourcesSettingsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Settings              struct {
			Text                       string `xml:",chardata"`
			DraftType                  string `xml:"draft_type"`
			IsAuctionDraft             string `xml:"is_auction_draft"`
			ScoringType                string `xml:"scoring_type"`
			PersistentURL              string `xml:"persistent_url"`
			UsesPlayoff                string `xml:"uses_playoff"`
			HasPlayoffConsolationGames string `xml:"has_playoff_consolation_games"`
			PlayoffStartWeek           string `xml:"playoff_start_week"`
			UsesPlayoffReseeding       int    `xml:"uses_playoff_reseeding"`
			UsesLockEliminatedTeams    int    `xml:"uses_lock_eliminated_teams"`
			NumPlayoffTeams            string `xml:"num_playoff_teams"`
			NumPlayoffConsolationTeams int    `xml:"num_playoff_consolation_teams"`
			HasMultiweekChampionship   string `xml:"has_multiweek_championship"`
			UsesRosterImport           string `xml:"uses_roster_import"`
			RosterImportDeadline       string `xml:"roster_import_deadline"`
			WaiverType                 string `xml:"waiver_type"`
			WaiverRule                 string `xml:"waiver_rule"`
			UsesFaab                   string `xml:"uses_faab"`
			DraftPickTime              string `xml:"draft_pick_time"`
			PostDraftPlayers           string `xml:"post_draft_players"`
			MaxTeams                   string `xml:"max_teams"`
			WaiverTime                 string `xml:"waiver_time"`
			TradeEndDate               string `xml:"trade_end_date"`
			TradeRatifyType            string `xml:"trade_ratify_type"`
			TradeRejectTime            string `xml:"trade_reject_time"`
			PlayerPool                 string `xml:"player_pool"`
			CantCutList                string `xml:"cant_cut_list"`
			IsPubliclyViewable         string `xml:"is_publicly_viewable"`
			CanTradeDraftPicks         string `xml:"can_trade_draft_picks"`
			SendbirdChannelURL         string `xml:"sendbird_channel_url"`
			RosterPositions            struct {
				Text           string `xml:",chardata"`
				RosterPosition []struct {
					Text         string `xml:",chardata"`
					Position     string `xml:"position"`
					PositionType string `xml:"position_type"`
					Count        string `xml:"count"`
				} `xml:"roster_position"`
			} `xml:"roster_positions"`
			StatCategories struct {
				Text  string `xml:",chardata"`
				Stats struct {
					Text string `xml:",chardata"`
					Stat []struct {
						Text              string `xml:",chardata"`
						StatID            int    `xml:"stat_id"`
						Enabled           string `xml:"enabled"`
						Name              string `xml:"name"`
						DisplayName       string `xml:"display_name"`
						SortOrder         string `xml:"sort_order"`
						PositionType      string `xml:"position_type"`
						StatPositionTypes struct {
							Text             string `xml:",chardata"`
							StatPositionType struct {
								Text              string `xml:",chardata"`
								PositionType      string `xml:"position_type"`
								IsOnlyDisplayStat string `xml:"is_only_display_stat"`
							} `xml:"stat_position_type"`
						} `xml:"stat_position_types"`
						IsOnlyDisplayStat     string `xml:"is_only_display_stat"`
						IsExcludedFromDisplay string `xml:"is_excluded_from_display"`
					} `xml:"stat"`
				} `xml:"stats"`
			} `xml:"stat_categories"`
			StatModifiers struct {
				Text  string `xml:",chardata"`
				Stats struct {
					Text string `xml:",chardata"`
					Stat []struct {
						Text   string  `xml:",chardata"`
						StatID int     `xml:"stat_id"`
						Value  float32 `xml:"value"`
						Bonus  struct {
							Target float32 `xml:"target"`
							Points float32 `xml:"points"`
						} `xml:"bonuses>bonus"`
					} `xml:"stat"`
				} `xml:"stats"`
			} `xml:"stat_modifiers"`
			MaxTrades            string `xml:"max_trades"`
			PickemEnabled        string `xml:"pickem_enabled"`
			UsesFractionalPoints string `xml:"uses_fractional_points"`
			UsesNegativePoints   string `xml:"uses_negative_points"`
		} `xml:"settings"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesSettings(ctx context.Context, leagueKey string) (*LeagueResourcesSettings, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/settings", leagueKey)
	res, err := s.get(ctx, url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	v := LeagueResourcesSettingsResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}

	var UsesNegativePoints bool
	if v.League.Settings.UsesNegativePoints == "1" {
		UsesNegativePoints = true
	}

	var UsesFractionalPoints bool
	if v.League.Settings.UsesFractionalPoints == "1" {
		UsesFractionalPoints = true
	}
	r := LeagueResourcesSettings{
		DraftType:      v.League.Settings.DraftType,
		IsAuctionDraft: v.League.Settings.IsAuctionDraft,
		ScoringType:    v.League.Settings.ScoringType,
		PersistentURL:  v.League.Settings.PersistentURL,
		UsesPlayoff:    v.League.Settings.UsesPlayoff,
		//HasPlayoffConsolationGames: v.League.Settings.UsesPlay,
		PlayoffStartWeek:           v.League.Settings.PlayoffStartWeek,
		UsesPlayoffReseeding:       v.League.Settings.UsesPlayoffReseeding,
		UsesLockEliminatedTeams:    v.League.Settings.UsesLockEliminatedTeams,
		NumPlayoffTeams:            v.League.Settings.NumPlayoffTeams,
		NumPlayoffConsolationTeams: v.League.Settings.NumPlayoffConsolationTeams,
		UsesRosterImport:           v.League.Settings.UsesRosterImport,
		RosterImportDeadline:       v.League.Settings.RosterImportDeadline,
		WaiverType:                 v.League.Settings.WaiverType,
		WaiverRule:                 v.League.Settings.WaiverRule,
		UsesFaab:                   v.League.Settings.UsesFaab,
		DraftTime:                  v.League.Settings.DraftPickTime,
		PostDraftPlayers:           v.League.Settings.PostDraftPlayers,
		MaxTeams:                   v.League.Settings.MaxTeams,
		WaiverTime:                 v.League.Settings.WaiverTime,
		TradeEndDate:               v.League.Settings.TradeEndDate,
		TradeRatifyType:            v.League.Settings.TradeRatifyType,
		TradeRejectTime:            v.League.Settings.TradeRejectTime,
		PlayerPool:                 v.League.Settings.PlayerPool,
		CantCutList:                v.League.Settings.CantCutList,
		IsPubliclyViewable:         v.League.Settings.IsPubliclyViewable,
		UsesNegativePoints:         UsesNegativePoints,
		UsesFractalPoints:          UsesFractionalPoints,
	}

	yahooStatCategories := v.League.Settings.StatCategories.Stats.Stat
	statCategoties := make([]StatCategory, len(v.League.Settings.StatCategories.Stats.Stat))
	for idx, val := range yahooStatCategories {
		statPositions := make([]PositionType, 1)

		posType := PositionType{
			PositionType:      val.StatPositionTypes.StatPositionType.PositionType,
			IsOnlyDisplayStat: val.StatPositionTypes.StatPositionType.IsOnlyDisplayStat,
		}
		statPositions[0] = posType

		statcategory := StatCategory{
			StatID:            val.StatID,
			Enabled:           val.Enabled,
			Name:              val.Name,
			DisplayName:       val.DisplayName,
			SortOrder:         val.SortOrder,
			PositionType:      val.PositionType,
			StatPositionTypes: statPositions,
			IsOnlyDisplayStat: val.IsOnlyDisplayStat,
		}
		statCategoties[idx] = statcategory
	}
	r.StatCategories = statCategoties

	yahooStatModifiers := v.League.Settings.StatModifiers.Stats.Stat
	statModifiers := make([]StatModifier, len(v.League.Settings.StatCategories.Stats.Stat))
	for idx, val := range yahooStatModifiers {

		var bonus *Bonus
		if val.Bonus.Target != 0 {
			bonus = &Bonus{}
			bonus.Target = val.Bonus.Target
			bonus.Points = val.Bonus.Points

		}

		statModifier := StatModifier{
			StatID:  val.StatID,
			Value:   val.Value,
			Bonuses: bonus,
		}
		statModifiers[idx] = statModifier
	}

	r.StatModifiers = statModifiers
	return &r, nil
}

type LeagueResourcesStandings struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              int    `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    bool   `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
	Standings             []struct {
		TeamKey          string `json:"team_key"`
		TeamID           string `json:"team_id"`
		Name             string `json:"name"`
		URL              string `json:"url"`
		TeamLogo         string `json:"team_logo"`
		WaiverPriority   int    `json:"waiver_priority"`
		NumberOfMoves    string `json:"number_of_moves"`
		NumberOfTrades   string `json:"number_of_trades"`
		ClinchedPlayoffs int    `json:"clinched_playoffs,omitempty"`
		Managers         []struct {
			ManagerID      string `json:"manager_id"`
			Nickname       string `json:"nickname"`
			GUID           string `json:"guid"`
			IsCommissioner string `json:"is_commissioner"`
		} `json:"managers"`
		Standings struct {
			Rank          int `json:"rank"`
			OutcomeTotals struct {
				Wins       string `json:"wins"`
				Losses     string `json:"losses"`
				Ties       string `json:"ties"`
				Percentage string `json:"percentage"`
			} `json:"outcome_totals"`
			GamesBack string `json:"games_back"`
		} `json:"standings"`
	} `json:"standings"`
}
type LeagueResourcesStandingsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Standings             struct {
			Text  string `xml:",chardata"`
			Teams struct {
				Text  string `xml:",chardata"`
				Count string `xml:"count,attr"`
				Team  []struct {
					Text                  string `xml:",chardata"`
					TeamKey               string `xml:"team_key"`
					TeamID                string `xml:"team_id"`
					Name                  string `xml:"name"`
					IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
					URL                   string `xml:"url"`
					TeamLogos             struct {
						Text     string `xml:",chardata"`
						TeamLogo struct {
							Text string `xml:",chardata"`
							Size string `xml:"size"`
							URL  string `xml:"url"`
						} `xml:"team_logo"`
					} `xml:"team_logos"`
					WaiverPriority string `xml:"waiver_priority"`
					NumberOfMoves  string `xml:"number_of_moves"`
					NumberOfTrades string `xml:"number_of_trades"`
					RosterAdds     struct {
						Text          string `xml:",chardata"`
						CoverageType  string `xml:"coverage_type"`
						CoverageValue string `xml:"coverage_value"`
						Value         string `xml:"value"`
					} `xml:"roster_adds"`
					LeagueScoringType string `xml:"league_scoring_type"`
					HasDraftGrade     string `xml:"has_draft_grade"`
					Managers          struct {
						Text    string `xml:",chardata"`
						Manager struct {
							Text           string `xml:",chardata"`
							ManagerID      string `xml:"manager_id"`
							Nickname       string `xml:"nickname"`
							Guid           string `xml:"guid"`
							IsCommissioner string `xml:"is_commissioner"`
							IsCurrentLogin string `xml:"is_current_login"`
							Email          string `xml:"email"`
							ImageURL       string `xml:"image_url"`
						} `xml:"manager"`
					} `xml:"managers"`
					TeamPoints struct {
						Text         string `xml:",chardata"`
						CoverageType string `xml:"coverage_type"`
						Season       string `xml:"season"`
						Total        string `xml:"total"`
					} `xml:"team_points"`
					TeamStandings struct {
						Text          string `xml:",chardata"`
						Rank          string `xml:"rank"`
						OutcomeTotals struct {
							Text       string `xml:",chardata"`
							Wins       string `xml:"wins"`
							Losses     string `xml:"losses"`
							Ties       string `xml:"ties"`
							Percentage string `xml:"percentage"`
						} `xml:"outcome_totals"`
						PointsFor     string `xml:"points_for"`
						PointsAgainst string `xml:"points_against"`
					} `xml:"team_standings"`
				} `xml:"team"`
			} `xml:"teams"`
		} `xml:"standings"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesStandings(leagueKey string) (*LeagueResourcesStandings, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/standings", leagueKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type LeagueResourcesScoreboard struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              int    `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    string `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
	Scoreboard            struct {
		Matchups []struct {
			Week          string `json:"week"`
			WeekStart     string `json:"week_start"`
			WeekEnd       string `json:"week_end"`
			Status        string `json:"status"`
			IsPlayoffs    string `json:"is_playoffs"`
			IsConsolation string `json:"is_consolation"`
			IsTied        int    `json:"is_tied"`
			WinnerTeamKey string `json:"winner_team_key"`
			Teams         []struct {
				TeamKey          string `json:"team_key"`
				TeamID           string `json:"team_id"`
				Name             string `json:"name"`
				URL              string `json:"url"`
				TeamLogo         string `json:"team_logo"`
				WaiverPriority   int    `json:"waiver_priority"`
				NumberOfMoves    string `json:"number_of_moves"`
				NumberOfTrades   int    `json:"number_of_trades"`
				ClinchedPlayoffs int    `json:"clinched_playoffs"`
				Managers         []struct {
					ManagerID      string `json:"manager_id"`
					Nickname       string `json:"nickname"`
					GUID           string `json:"guid"`
					IsCommissioner string `json:"is_commissioner"`
				} `json:"managers"`
				Points struct {
					CoverageType string `json:"coverage_type"`
					Week         string `json:"week"`
					Total        string `json:"total"`
				} `json:"points"`
				Stats []struct {
					StatID string `json:"stat_id"`
					Value  string `json:"value"`
				} `json:"stats"`
			} `json:"teams"`
		} `json:"matchups"`
		Week string `json:"week"`
	} `json:"scoreboard"`
}
type LeagueResourcesScoreboardResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Scoreboard            struct {
			Text     string `xml:",chardata"`
			Week     string `xml:"week"`
			Matchups struct {
				Text    string `xml:",chardata"`
				Count   string `xml:"count,attr"`
				Matchup []struct {
					Text                    string `xml:",chardata"`
					Week                    string `xml:"week"`
					WeekStart               string `xml:"week_start"`
					WeekEnd                 string `xml:"week_end"`
					Status                  string `xml:"status"`
					IsPlayoffs              string `xml:"is_playoffs"`
					IsConsolation           string `xml:"is_consolation"`
					IsMatchupRecapAvailable string `xml:"is_matchup_recap_available"`
					Teams                   struct {
						Text  string `xml:",chardata"`
						Count string `xml:"count,attr"`
						Team  []struct {
							Text                  string `xml:",chardata"`
							TeamKey               string `xml:"team_key"`
							TeamID                string `xml:"team_id"`
							Name                  string `xml:"name"`
							IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
							URL                   string `xml:"url"`
							TeamLogos             struct {
								Text     string `xml:",chardata"`
								TeamLogo struct {
									Text string `xml:",chardata"`
									Size string `xml:"size"`
									URL  string `xml:"url"`
								} `xml:"team_logo"`
							} `xml:"team_logos"`
							WaiverPriority string `xml:"waiver_priority"`
							NumberOfMoves  string `xml:"number_of_moves"`
							NumberOfTrades string `xml:"number_of_trades"`
							RosterAdds     struct {
								Text          string `xml:",chardata"`
								CoverageType  string `xml:"coverage_type"`
								CoverageValue string `xml:"coverage_value"`
								Value         string `xml:"value"`
							} `xml:"roster_adds"`
							LeagueScoringType string `xml:"league_scoring_type"`
							HasDraftGrade     string `xml:"has_draft_grade"`
							Managers          struct {
								Text    string `xml:",chardata"`
								Manager struct {
									Text           string `xml:",chardata"`
									ManagerID      string `xml:"manager_id"`
									Nickname       string `xml:"nickname"`
									Guid           string `xml:"guid"`
									IsCommissioner string `xml:"is_commissioner"`
									IsCurrentLogin string `xml:"is_current_login"`
									Email          string `xml:"email"`
									ImageURL       string `xml:"image_url"`
								} `xml:"manager"`
							} `xml:"managers"`
							WinProbability string `xml:"win_probability"`
							TeamPoints     struct {
								Text         string `xml:",chardata"`
								CoverageType string `xml:"coverage_type"`
								Week         string `xml:"week"`
								Total        string `xml:"total"`
							} `xml:"team_points"`
							TeamProjectedPoints struct {
								Text         string `xml:",chardata"`
								CoverageType string `xml:"coverage_type"`
								Week         string `xml:"week"`
								Total        string `xml:"total"`
							} `xml:"team_projected_points"`
						} `xml:"team"`
					} `xml:"teams"`
				} `xml:"matchup"`
			} `xml:"matchups"`
		} `xml:"scoreboard"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesScoreboard(leagueKey string) (*LeagueResourcesScoreboard, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/scoreboard", leagueKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type LeagueResourcesTeams struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              int    `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    string `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
	Teams                 []struct {
		TeamKey          string `json:"team_key"`
		TeamID           string `json:"team_id"`
		Name             string `json:"name"`
		URL              string `json:"url"`
		TeamLogo         string `json:"team_logo"`
		WaiverPriority   int    `json:"waiver_priority"`
		NumberOfMoves    string `json:"number_of_moves"`
		NumberOfTrades   int    `json:"number_of_trades"`
		ClinchedPlayoffs int    `json:"clinched_playoffs,omitempty"`
		Managers         []struct {
			ManagerID      string `json:"manager_id"`
			Nickname       string `json:"nickname"`
			GUID           string `json:"guid"`
			IsCommissioner string `json:"is_commissioner"`
		} `json:"managers"`
	} `json:"teams"`
}
type LeagueResourcesTeamsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Teams                 struct {
			Text  string `xml:",chardata"`
			Count string `xml:"count,attr"`
			Team  []struct {
				Text                  string `xml:",chardata"`
				TeamKey               string `xml:"team_key"`
				TeamID                string `xml:"team_id"`
				Name                  string `xml:"name"`
				IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
				URL                   string `xml:"url"`
				TeamLogos             struct {
					Text     string `xml:",chardata"`
					TeamLogo struct {
						Text string `xml:",chardata"`
						Size string `xml:"size"`
						URL  string `xml:"url"`
					} `xml:"team_logo"`
				} `xml:"team_logos"`
				WaiverPriority string `xml:"waiver_priority"`
				NumberOfMoves  string `xml:"number_of_moves"`
				NumberOfTrades string `xml:"number_of_trades"`
				RosterAdds     struct {
					Text          string `xml:",chardata"`
					CoverageType  string `xml:"coverage_type"`
					CoverageValue string `xml:"coverage_value"`
					Value         string `xml:"value"`
				} `xml:"roster_adds"`
				LeagueScoringType string `xml:"league_scoring_type"`
				HasDraftGrade     string `xml:"has_draft_grade"`
				Managers          struct {
					Text    string `xml:",chardata"`
					Manager struct {
						Text           string `xml:",chardata"`
						ManagerID      string `xml:"manager_id"`
						Nickname       string `xml:"nickname"`
						Guid           string `xml:"guid"`
						IsCommissioner string `xml:"is_commissioner"`
						IsCurrentLogin string `xml:"is_current_login"`
						Email          string `xml:"email"`
						ImageURL       string `xml:"image_url"`
					} `xml:"manager"`
				} `xml:"managers"`
			} `xml:"team"`
		} `xml:"teams"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesTeams(leagueKey string) (*LeagueResourcesTeams, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/teams", leagueKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type LeagueResourcesDraftResults struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              int    `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    string `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
	DraftResults          []struct {
		Pick      int    `json:"pick"`
		Round     int    `json:"round"`
		Cost      string `json:"cost"`
		TeamKey   string `json:"team_key"`
		PlayerKey string `json:"player_key"`
	} `json:"draft_results"`
}
type LeagueResourcesDraftResultsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		IsFinished            string `xml:"is_finished"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		DraftResults          struct {
			Text        string `xml:",chardata"`
			Count       string `xml:"count,attr"`
			DraftResult []struct {
				Text      string `xml:",chardata"`
				Pick      string `xml:"pick"`
				Round     string `xml:"round"`
				TeamKey   string `xml:"team_key"`
				PlayerKey string `xml:"player_key"`
			} `xml:"draft_result"`
		} `xml:"draft_results"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesDraftResults(leagueKey string) (*LeagueResourcesDraftResults, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/draftresults", leagueKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type LeagueResourcesTransaction struct {
	LeagueKey             string `json:"league_key"`
	LeagueID              string `json:"league_id"`
	Name                  string `json:"name"`
	URL                   string `json:"url"`
	DraftStatus           string `json:"draft_status"`
	NumTeams              int    `json:"num_teams"`
	EditKey               string `json:"edit_key"`
	WeeklyDeadline        string `json:"weekly_deadline"`
	LeagueUpdateTimestamp string `json:"league_update_timestamp"`
	ScoringType           string `json:"scoring_type"`
	LeagueType            string `json:"league_type"`
	Renew                 string `json:"renew"`
	Renewed               string `json:"renewed"`
	ShortInvitationURL    string `json:"short_invitation_url"`
	IsProLeague           string `json:"is_pro_league"`
	CurrentWeek           string `json:"current_week"`
	StartWeek             string `json:"start_week"`
	StartDate             string `json:"start_date"`
	EndWeek               string `json:"end_week"`
	EndDate               string `json:"end_date"`
	IsFinished            int    `json:"is_finished"`
	Transactions          []struct {
		TransactionKey string        `json:"transaction_key"`
		TransactionID  string        `json:"transaction_id"`
		Type           string        `json:"type"`
		Status         string        `json:"status"`
		Timestamp      string        `json:"timestamp"`
		Players        []interface{} `json:"players"`
	} `json:"transactions"`
}
type LeagueResourcesTransactionResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		IsFinished            string `xml:"is_finished"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Transactions          struct {
			Text        string `xml:",chardata"`
			Count       string `xml:"count,attr"`
			Transaction []struct {
				Text           string `xml:",chardata"`
				TransactionKey string `xml:"transaction_key"`
				TransactionID  string `xml:"transaction_id"`
				Type           string `xml:"type"`
				Status         string `xml:"status"`
				Timestamp      string `xml:"timestamp"`
				Players        struct {
					Text   string `xml:",chardata"`
					Count  string `xml:"count,attr"`
					Player []struct {
						Text      string `xml:",chardata"`
						PlayerKey string `xml:"player_key"`
						PlayerID  string `xml:"player_id"`
						Name      struct {
							Text       string `xml:",chardata"`
							Full       string `xml:"full"`
							First      string `xml:"first"`
							Last       string `xml:"last"`
							AsciiFirst string `xml:"ascii_first"`
							AsciiLast  string `xml:"ascii_last"`
						} `xml:"name"`
						EditorialTeamAbbr string `xml:"editorial_team_abbr"`
						DisplayPosition   string `xml:"display_position"`
						PositionType      string `xml:"position_type"`
						TransactionData   struct {
							Text                string `xml:",chardata"`
							Type                string `xml:"type"`
							SourceType          string `xml:"source_type"`
							DestinationType     string `xml:"destination_type"`
							DestinationTeamKey  string `xml:"destination_team_key"`
							DestinationTeamName string `xml:"destination_team_name"`
							SourceTeamKey       string `xml:"source_team_key"`
							SourceTeamName      string `xml:"source_team_name"`
						} `xml:"transaction_data"`
					} `xml:"player"`
				} `xml:"players"`
			} `xml:"transaction"`
		} `xml:"transactions"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesTransaction(leagueKey string) (*LeagueResourcesTransaction, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/transactions", leagueKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type GetLeagueResourcesPlayersResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		IsFinished            string `xml:"is_finished"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Players               struct {
			Text   string `xml:",chardata"`
			Count  string `xml:"count,attr"`
			Player []struct {
				Text      string `xml:",chardata"`
				PlayerKey string `xml:"player_key"`
				PlayerID  string `xml:"player_id"`
				Name      struct {
					Text       string `xml:",chardata"`
					Full       string `xml:"full"`
					First      string `xml:"first"`
					Last       string `xml:"last"`
					AsciiFirst string `xml:"ascii_first"`
					AsciiLast  string `xml:"ascii_last"`
				} `xml:"name"`
				Status                string `xml:"status"`
				StatusFull            string `xml:"status_full"`
				EditorialPlayerKey    string `xml:"editorial_player_key"`
				EditorialTeamKey      string `xml:"editorial_team_key"`
				EditorialTeamFullName string `xml:"editorial_team_full_name"`
				EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
				ByeWeeks              struct {
					Text string `xml:",chardata"`
					Week string `xml:"week"`
				} `xml:"bye_weeks"`
				UniformNumber   string `xml:"uniform_number"`
				DisplayPosition string `xml:"display_position"`
				Headshot        struct {
					Text string `xml:",chardata"`
					URL  string `xml:"url"`
					Size string `xml:"size"`
				} `xml:"headshot"`
				ImageURL          string `xml:"image_url"`
				IsUndroppable     string `xml:"is_undroppable"`
				PositionType      string `xml:"position_type"`
				PrimaryPosition   string `xml:"primary_position"`
				EligiblePositions struct {
					Text     string   `xml:",chardata"`
					Position []string `xml:"position"`
				} `xml:"eligible_positions"`
				HasPlayerNotes           string `xml:"has_player_notes"`
				PlayerNotesLastTimestamp string `xml:"player_notes_last_timestamp"`
			} `xml:"player"`
		} `xml:"players"`
	} `xml:"league"`
}
type GetLeagueResourcesPlayersStatsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		IsFinished            string `xml:"is_finished"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Players               struct {
			Text   string `xml:",chardata"`
			Count  string `xml:"count,attr"`
			Player struct {
				Text      string `xml:",chardata"`
				PlayerKey string `xml:"player_key"`
				PlayerID  string `xml:"player_id"`
				Name      struct {
					Text       string `xml:",chardata"`
					Full       string `xml:"full"`
					First      string `xml:"first"`
					Last       string `xml:"last"`
					AsciiFirst string `xml:"ascii_first"`
					AsciiLast  string `xml:"ascii_last"`
				} `xml:"name"`
				Status                string `xml:"status"`
				StatusFull            string `xml:"status_full"`
				EditorialPlayerKey    string `xml:"editorial_player_key"`
				EditorialTeamKey      string `xml:"editorial_team_key"`
				EditorialTeamFullName string `xml:"editorial_team_full_name"`
				EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
				ByeWeeks              struct {
					Text string `xml:",chardata"`
					Week string `xml:"week"`
				} `xml:"bye_weeks"`
				UniformNumber   string `xml:"uniform_number"`
				DisplayPosition string `xml:"display_position"`
				Headshot        struct {
					Text string `xml:",chardata"`
					URL  string `xml:"url"`
					Size string `xml:"size"`
				} `xml:"headshot"`
				ImageURL          string `xml:"image_url"`
				IsUndroppable     string `xml:"is_undroppable"`
				PositionType      string `xml:"position_type"`
				PrimaryPosition   string `xml:"primary_position"`
				EligiblePositions struct {
					Text     string `xml:",chardata"`
					Position string `xml:"position"`
				} `xml:"eligible_positions"`
				PlayerStats struct {
					Text         string `xml:",chardata"`
					CoverageType string `xml:"coverage_type"`
					Week         string `xml:"week"`
					Stats        struct {
						Text string `xml:",chardata"`
						Stat []struct {
							Text   string `xml:",chardata"`
							StatID string `xml:"stat_id"`
							Value  string `xml:"value"`
						} `xml:"stat"`
					} `xml:"stats"`
				} `xml:"player_stats"`
				PlayerPoints struct {
					Text         string `xml:",chardata"`
					CoverageType string `xml:"coverage_type"`
					Week         string `xml:"week"`
					Total        string `xml:"total"`
				} `xml:"player_points"`
			} `xml:"player"`
		} `xml:"players"`
	} `xml:"league"`
}

func (s *Service) GetLeagueResourcesPlayers(leagueKey string, playerKeys []string, week int) (*GetLeagueResourcesPlayersResponse, error) {

	playerKeyStrings := strings.Join(playerKeys, ",")
	var weekString string
	if week > 0 {
		weekString = fmt.Sprintf(";type=week;week=%d", week)
	}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/players;player_keys=%s/stats%s", leagueKey, playerKeyStrings, weekString)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

/*

	Players

*/

type PlayerResourcesMeta struct {
	PlayerKey string `json:"player_key"`
	PlayerID  string `json:"player_id"`
	Name      struct {
		Full       string `json:"full"`
		First      string `json:"first"`
		Last       string `json:"last"`
		ASCIIFirst string `json:"ascii_first"`
		ASCIILast  string `json:"ascii_last"`
	} `json:"name"`
	EditorialPlayerKey    string   `json:"editorial_player_key"`
	EditorialTeamKey      string   `json:"editorial_team_key"`
	EditorialTeamFullName string   `json:"editorial_team_full_name"`
	EditorialTeamAbbr     string   `json:"editorial_team_abbr"`
	UniformNumber         string   `json:"uniform_number"`
	DisplayPosition       string   `json:"display_position"`
	Headshot              string   `json:"headshot"`
	IsUndroppable         string   `json:"is_undroppable"`
	PositionType          string   `json:"position_type"`
	EligiblePositions     []string `json:"eligible_positions"`
}
type PlayerResourcesMetaResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Player      struct {
		Text      string `xml:",chardata"`
		PlayerKey string `xml:"player_key"`
		PlayerID  string `xml:"player_id"`
		Name      struct {
			Text       string `xml:",chardata"`
			Full       string `xml:"full"`
			First      string `xml:"first"`
			Last       string `xml:"last"`
			AsciiFirst string `xml:"ascii_first"`
			AsciiLast  string `xml:"ascii_last"`
		} `xml:"name"`
		Status                string `xml:"status"`
		StatusFull            string `xml:"status_full"`
		EditorialPlayerKey    string `xml:"editorial_player_key"`
		EditorialTeamKey      string `xml:"editorial_team_key"`
		EditorialTeamFullName string `xml:"editorial_team_full_name"`
		EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
		ByeWeeks              struct {
			Text string `xml:",chardata"`
			Week string `xml:"week"`
		} `xml:"bye_weeks"`
		UniformNumber   string `xml:"uniform_number"`
		DisplayPosition string `xml:"display_position"`
		Headshot        struct {
			Text string `xml:",chardata"`
			URL  string `xml:"url"`
			Size string `xml:"size"`
		} `xml:"headshot"`
		ImageURL          string `xml:"image_url"`
		IsUndroppable     string `xml:"is_undroppable"`
		PositionType      string `xml:"position_type"`
		EligiblePositions struct {
			Text     string `xml:",chardata"`
			Position string `xml:"position"`
		} `xml:"eligible_positions"`
	} `xml:"player"`
}

func (s *Service) GetPlayerResourcesMeta(playerKey string) (*PlayerResourcesMeta, error) {
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/player/%s/metadata", playerKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type PlayerResourcesStats struct {
	PlayerKey string `json:"player_key"`
	PlayerID  string `json:"player_id"`
	Name      struct {
		Full       string `json:"full"`
		First      string `json:"first"`
		Last       string `json:"last"`
		ASCIIFirst string `json:"ascii_first"`
		ASCIILast  string `json:"ascii_last"`
	} `json:"name"`
	EditorialPlayerKey    string   `json:"editorial_player_key"`
	EditorialTeamKey      string   `json:"editorial_team_key"`
	EditorialTeamFullName string   `json:"editorial_team_full_name"`
	EditorialTeamAbbr     string   `json:"editorial_team_abbr"`
	UniformNumber         string   `json:"uniform_number"`
	DisplayPosition       string   `json:"display_position"`
	Headshot              string   `json:"headshot"`
	IsUndroppable         string   `json:"is_undroppable"`
	PositionType          string   `json:"position_type"`
	EligiblePositions     []string `json:"eligible_positions"`
	Stats                 struct {
		CoverageType  string `json:"coverage_type"`
		CoverageValue string `json:"coverage_value"`
		Stats         []struct {
			StatID string `json:"stat_id"`
			Value  string `json:"value"`
		} `json:"stats"`
	} `json:"stats,omitempty"`
	Ownership struct {
		Value int `json:"value,omitempty"`
	} `json:"ownership,omitempty"`
}
type PlayerResourcesStatsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Player      struct {
		Text      string `xml:",chardata"`
		PlayerKey string `xml:"player_key"`
		PlayerID  string `xml:"player_id"`
		Name      struct {
			Text       string `xml:",chardata"`
			Full       string `xml:"full"`
			First      string `xml:"first"`
			Last       string `xml:"last"`
			AsciiFirst string `xml:"ascii_first"`
			AsciiLast  string `xml:"ascii_last"`
		} `xml:"name"`
		Status                string `xml:"status"`
		StatusFull            string `xml:"status_full"`
		EditorialPlayerKey    string `xml:"editorial_player_key"`
		EditorialTeamKey      string `xml:"editorial_team_key"`
		EditorialTeamFullName string `xml:"editorial_team_full_name"`
		EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
		ByeWeeks              struct {
			Text string `xml:",chardata"`
			Week string `xml:"week"`
		} `xml:"bye_weeks"`
		UniformNumber   string `xml:"uniform_number"`
		DisplayPosition string `xml:"display_position"`
		Headshot        struct {
			Text string `xml:",chardata"`
			URL  string `xml:"url"`
			Size string `xml:"size"`
		} `xml:"headshot"`
		ImageURL          string `xml:"image_url"`
		IsUndroppable     string `xml:"is_undroppable"`
		PositionType      string `xml:"position_type"`
		EligiblePositions struct {
			Text     string `xml:",chardata"`
			Position string `xml:"position"`
		} `xml:"eligible_positions"`
		PlayerStats struct {
			Text         string `xml:",chardata"`
			CoverageType string `xml:"coverage_type"`
			Season       string `xml:"season"`
			Stats        struct {
				Text string `xml:",chardata"`
				Stat []struct {
					Text   string `xml:",chardata"`
					StatID string `xml:"stat_id"`
					Value  string `xml:"value"`
				} `xml:"stat"`
			} `xml:"stats"`
		} `xml:"player_stats"`
	} `xml:"player"`
}

func (s *Service) GetPlayerResourcesStats(playerKey string, week int) (*PlayerResourcesStats, error) {
	var weekString string
	if week > 0 {
		weekString = fmt.Sprintf(";type=week;week=%d", week)
	}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/player/%s/stats%s", playerKey, weekString)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type PlayerResourcesPercentOwned struct {
	PlayerKey string `json:"player_key"`
	PlayerID  string `json:"player_id"`
	Name      struct {
		Full       string `json:"full"`
		First      string `json:"first"`
		Last       string `json:"last"`
		ASCIIFirst string `json:"ascii_first"`
		ASCIILast  string `json:"ascii_last"`
	} `json:"name"`
	EditorialPlayerKey    string   `json:"editorial_player_key"`
	EditorialTeamKey      string   `json:"editorial_team_key"`
	EditorialTeamFullName string   `json:"editorial_team_full_name"`
	EditorialTeamAbbr     string   `json:"editorial_team_abbr"`
	UniformNumber         string   `json:"uniform_number"`
	DisplayPosition       string   `json:"display_position"`
	Headshot              string   `json:"headshot"`
	IsUndroppable         string   `json:"is_undroppable"`
	PositionType          string   `json:"position_type"`
	EligiblePositions     []string `json:"eligible_positions"`
	Ownership             struct {
		Value int `json:"value"`
	} `json:"ownership"`
}
type PlayerResourcesPercentOwnedResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Player      struct {
		Text      string `xml:",chardata"`
		PlayerKey string `xml:"player_key"`
		PlayerID  string `xml:"player_id"`
		Name      struct {
			Text       string `xml:",chardata"`
			Full       string `xml:"full"`
			First      string `xml:"first"`
			Last       string `xml:"last"`
			AsciiFirst string `xml:"ascii_first"`
			AsciiLast  string `xml:"ascii_last"`
		} `xml:"name"`
		Status                string `xml:"status"`
		StatusFull            string `xml:"status_full"`
		EditorialPlayerKey    string `xml:"editorial_player_key"`
		EditorialTeamKey      string `xml:"editorial_team_key"`
		EditorialTeamFullName string `xml:"editorial_team_full_name"`
		EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
		ByeWeeks              struct {
			Text string `xml:",chardata"`
			Week string `xml:"week"`
		} `xml:"bye_weeks"`
		UniformNumber   string `xml:"uniform_number"`
		DisplayPosition string `xml:"display_position"`
		Headshot        struct {
			Text string `xml:",chardata"`
			URL  string `xml:"url"`
			Size string `xml:"size"`
		} `xml:"headshot"`
		ImageURL          string `xml:"image_url"`
		IsUndroppable     string `xml:"is_undroppable"`
		PositionType      string `xml:"position_type"`
		EligiblePositions struct {
			Text     string `xml:",chardata"`
			Position string `xml:"position"`
		} `xml:"eligible_positions"`
		PercentOwned struct {
			Text         string `xml:",chardata"`
			CoverageType string `xml:"coverage_type"`
			Week         string `xml:"week"`
			Value        string `xml:"value"`
			Delta        string `xml:"delta"`
		} `xml:"percent_owned"`
	} `xml:"player"`
}

func (s *Service) GetPlayerResourcesPercentOwned(playerKey string) (*PlayerResourcesPercentOwned, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/player/%s/percent_owned", playerKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type PlayerResourcesOwnership struct {
	PlayerKey string `json:"player_key"`
	PlayerID  string `json:"player_id"`
	Name      struct {
		Full       string `json:"full"`
		First      string `json:"first"`
		Last       string `json:"last"`
		ASCIIFirst string `json:"ascii_first"`
		ASCIILast  string `json:"ascii_last"`
	} `json:"name"`
	EditorialPlayerKey    string   `json:"editorial_player_key"`
	EditorialTeamKey      string   `json:"editorial_team_key"`
	EditorialTeamFullName string   `json:"editorial_team_full_name"`
	EditorialTeamAbbr     string   `json:"editorial_team_abbr"`
	UniformNumber         string   `json:"uniform_number"`
	DisplayPosition       string   `json:"display_position"`
	Headshot              string   `json:"headshot"`
	IsUndroppable         string   `json:"is_undroppable"`
	PositionType          string   `json:"position_type"`
	EligiblePositions     []string `json:"eligible_positions"`
	Ownership             struct {
		Value int `json:"value"`
	} `json:"ownership"`
}
type PlayerResourcesOwnershipResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	League      struct {
		Text                  string `xml:",chardata"`
		LeagueKey             string `xml:"league_key"`
		LeagueID              string `xml:"league_id"`
		Name                  string `xml:"name"`
		URL                   string `xml:"url"`
		LogoURL               string `xml:"logo_url"`
		Password              string `xml:"password"`
		DraftStatus           string `xml:"draft_status"`
		NumTeams              string `xml:"num_teams"`
		EditKey               string `xml:"edit_key"`
		WeeklyDeadline        string `xml:"weekly_deadline"`
		LeagueUpdateTimestamp string `xml:"league_update_timestamp"`
		ScoringType           string `xml:"scoring_type"`
		LeagueType            string `xml:"league_type"`
		Renew                 string `xml:"renew"`
		Renewed               string `xml:"renewed"`
		IrisGroupChatID       string `xml:"iris_group_chat_id"`
		ShortInvitationURL    string `xml:"short_invitation_url"`
		AllowAddToDlExtraPos  string `xml:"allow_add_to_dl_extra_pos"`
		IsProLeague           string `xml:"is_pro_league"`
		IsCashLeague          string `xml:"is_cash_league"`
		CurrentWeek           string `xml:"current_week"`
		StartWeek             string `xml:"start_week"`
		StartDate             string `xml:"start_date"`
		EndWeek               string `xml:"end_week"`
		EndDate               string `xml:"end_date"`
		IsFinished            string `xml:"is_finished"`
		GameCode              string `xml:"game_code"`
		Season                string `xml:"season"`
		Players               struct {
			Text   string `xml:",chardata"`
			Count  string `xml:"count,attr"`
			Player struct {
				Text      string `xml:",chardata"`
				PlayerKey string `xml:"player_key"`
				PlayerID  string `xml:"player_id"`
				Name      struct {
					Text       string `xml:",chardata"`
					Full       string `xml:"full"`
					First      string `xml:"first"`
					Last       string `xml:"last"`
					AsciiFirst string `xml:"ascii_first"`
					AsciiLast  string `xml:"ascii_last"`
				} `xml:"name"`
				Status                string `xml:"status"`
				StatusFull            string `xml:"status_full"`
				EditorialPlayerKey    string `xml:"editorial_player_key"`
				EditorialTeamKey      string `xml:"editorial_team_key"`
				EditorialTeamFullName string `xml:"editorial_team_full_name"`
				EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
				ByeWeeks              struct {
					Text string `xml:",chardata"`
					Week string `xml:"week"`
				} `xml:"bye_weeks"`
				UniformNumber   string `xml:"uniform_number"`
				DisplayPosition string `xml:"display_position"`
				Headshot        struct {
					Text string `xml:",chardata"`
					URL  string `xml:"url"`
					Size string `xml:"size"`
				} `xml:"headshot"`
				ImageURL          string `xml:"image_url"`
				IsUndroppable     string `xml:"is_undroppable"`
				PositionType      string `xml:"position_type"`
				PrimaryPosition   string `xml:"primary_position"`
				EligiblePositions struct {
					Text     string `xml:",chardata"`
					Position string `xml:"position"`
				} `xml:"eligible_positions"`
				Ownership struct {
					Text          string `xml:",chardata"`
					OwnershipType string `xml:"ownership_type"`
					WaiverDate    string `xml:"waiver_date"`
				} `xml:"ownership"`
			} `xml:"player"`
		} `xml:"players"`
	} `xml:"league"`
}

func (s *Service) GetPlayerResourcesOwnership(leagueKey, playerKey string) (*PlayerResourcesOwnership, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/league/%s/players;player_keys=%s/ownership", leagueKey, playerKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type PlayerResourcesDraftAnalysis struct {
	PlayerKey string `json:"player_key"`
	PlayerID  string `json:"player_id"`
	Name      struct {
		Full       string `json:"full"`
		First      string `json:"first"`
		Last       string `json:"last"`
		ASCIIFirst string `json:"ascii_first"`
		ASCIILast  string `json:"ascii_last"`
	} `json:"name"`
	EditorialPlayerKey    string   `json:"editorial_player_key"`
	EditorialTeamKey      string   `json:"editorial_team_key"`
	EditorialTeamFullName string   `json:"editorial_team_full_name"`
	EditorialTeamAbbr     string   `json:"editorial_team_abbr"`
	UniformNumber         string   `json:"uniform_number"`
	DisplayPosition       string   `json:"display_position"`
	Headshot              string   `json:"headshot"`
	IsUndroppable         string   `json:"is_undroppable"`
	PositionType          string   `json:"position_type"`
	EligiblePositions     []string `json:"eligible_positions"`
	DraftAnalysis         struct {
		AveragePick    string `json:"average_pick"`
		AverageRound   string `json:"average_round"`
		AverageCost    string `json:"average_cost"`
		PercentDrafted string `json:"percent_drafted"`
	} `json:"draft_analysis"`
}
type PlayerResourcesDraftAnalysisResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Player      struct {
		Text      string `xml:",chardata"`
		PlayerKey string `xml:"player_key"`
		PlayerID  string `xml:"player_id"`
		Name      struct {
			Text       string `xml:",chardata"`
			Full       string `xml:"full"`
			First      string `xml:"first"`
			Last       string `xml:"last"`
			AsciiFirst string `xml:"ascii_first"`
			AsciiLast  string `xml:"ascii_last"`
		} `xml:"name"`
		Status                string `xml:"status"`
		StatusFull            string `xml:"status_full"`
		EditorialPlayerKey    string `xml:"editorial_player_key"`
		EditorialTeamKey      string `xml:"editorial_team_key"`
		EditorialTeamFullName string `xml:"editorial_team_full_name"`
		EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
		ByeWeeks              struct {
			Text string `xml:",chardata"`
			Week string `xml:"week"`
		} `xml:"bye_weeks"`
		UniformNumber   string `xml:"uniform_number"`
		DisplayPosition string `xml:"display_position"`
		Headshot        struct {
			Text string `xml:",chardata"`
			URL  string `xml:"url"`
			Size string `xml:"size"`
		} `xml:"headshot"`
		ImageURL          string `xml:"image_url"`
		IsUndroppable     string `xml:"is_undroppable"`
		PositionType      string `xml:"position_type"`
		EligiblePositions struct {
			Text     string `xml:",chardata"`
			Position string `xml:"position"`
		} `xml:"eligible_positions"`
		DraftAnalysis struct {
			Text           string `xml:",chardata"`
			AveragePick    string `xml:"average_pick"`
			AverageRound   string `xml:"average_round"`
			AverageCost    string `xml:"average_cost"`
			PercentDrafted string `xml:"percent_drafted"`
		} `xml:"draft_analysis"`
	} `xml:"player"`
}

func (s *Service) GetPlayerResourcesDraftAnalysis(playerKey string) (*PlayerResourcesDraftAnalysis, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/player/%s/draft_analysis", playerKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

/*
	Roster
*/
type RosterResourcesPlayers struct {
	TeamKey          string        `json:"team_key"`
	TeamID           string        `json:"team_id"`
	Name             string        `json:"name"`
	URL              string        `json:"url"`
	TeamLogo         string        `json:"team_logo"`
	WaiverPriority   int           `json:"waiver_priority"`
	NumberOfMoves    string        `json:"number_of_moves"`
	NumberOfTrades   int           `json:"number_of_trades"`
	ClinchedPlayoffs int           `json:"clinched_playoffs"`
	Managers         []interface{} `json:"managers"`
	Roster           []struct {
		PlayerKey string `json:"player_key"`
		PlayerID  string `json:"player_id"`
		Name      struct {
			Full       string `json:"full"`
			First      string `json:"first"`
			Last       string `json:"last"`
			ASCIIFirst string `json:"ascii_first"`
			ASCIILast  string `json:"ascii_last"`
		} `json:"name"`
		EditorialPlayerKey    string `json:"editorial_player_key"`
		EditorialTeamKey      string `json:"editorial_team_key"`
		EditorialTeamFullName string `json:"editorial_team_full_name"`
		EditorialTeamAbbr     string `json:"editorial_team_abbr"`
		UniformNumber         string `json:"uniform_number"`
		DisplayPosition       string `json:"display_position"`
		Headshot              struct {
			URL  string `json:"url"`
			Size string `json:"size"`
		} `json:"headshot"`
		IsUndroppable        string   `json:"is_undroppable"`
		PositionType         string   `json:"position_type"`
		EligiblePositions    []string `json:"eligible_positions"`
		HasPlayerNotes       int      `json:"has_player_notes,omitempty"`
		Status               string   `json:"status,omitempty"`
		OnDisabledList       string   `json:"on_disabled_list,omitempty"`
		HasRecentPlayerNotes int      `json:"has_recent_player_notes,omitempty"`
	} `json:"roster"`
}
type RosterResourcesPlayersResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
		Roster struct {
			Text         string `xml:",chardata"`
			CoverageType string `xml:"coverage_type"`
			Week         string `xml:"week"`
			IsEditable   string `xml:"is_editable"`
			Players      struct {
				Text   string `xml:",chardata"`
				Count  string `xml:"count,attr"`
				Player []struct {
					Text      string `xml:",chardata"`
					PlayerKey string `xml:"player_key"`
					PlayerID  string `xml:"player_id"`
					Name      struct {
						Text       string `xml:",chardata"`
						Full       string `xml:"full"`
						First      string `xml:"first"`
						Last       string `xml:"last"`
						AsciiFirst string `xml:"ascii_first"`
						AsciiLast  string `xml:"ascii_last"`
					} `xml:"name"`
					EditorialPlayerKey    string `xml:"editorial_player_key"`
					EditorialTeamKey      string `xml:"editorial_team_key"`
					EditorialTeamFullName string `xml:"editorial_team_full_name"`
					EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
					ByeWeeks              struct {
						Text string `xml:",chardata"`
						Week string `xml:"week"`
					} `xml:"bye_weeks"`
					UniformNumber   string `xml:"uniform_number"`
					DisplayPosition string `xml:"display_position"`
					Headshot        struct {
						Text string `xml:",chardata"`
						URL  string `xml:"url"`
						Size string `xml:"size"`
					} `xml:"headshot"`
					ImageURL          string `xml:"image_url"`
					IsUndroppable     string `xml:"is_undroppable"`
					PositionType      string `xml:"position_type"`
					PrimaryPosition   string `xml:"primary_position"`
					EligiblePositions struct {
						Text     string   `xml:",chardata"`
						Position []string `xml:"position"`
					} `xml:"eligible_positions"`
					SelectedPosition struct {
						Text         string `xml:",chardata"`
						CoverageType string `xml:"coverage_type"`
						Week         string `xml:"week"`
						Position     string `xml:"position"`
						IsFlex       string `xml:"is_flex"`
					} `xml:"selected_position"`
					IsEditable               string `xml:"is_editable"`
					HasPlayerNotes           string `xml:"has_player_notes"`
					PlayerNotesLastTimestamp string `xml:"player_notes_last_timestamp"`
					Status                   string `xml:"status"`
					StatusFull               string `xml:"status_full"`
				} `xml:"player"`
			} `xml:"players"`
		} `xml:"roster"`
	} `xml:"team"`
}

func (s *Service) GetRosterResourcesPlayers(teamKey string, dateString string) (*RosterResourcesPlayers, error) {
	var dateFormat string
	if len(dateString) > 2 {
		dateFormat = fmt.Sprintf(";date=%s", dateString)
	} else {
		dateFormat = fmt.Sprintf(";week=%s", dateString)
	}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/roster%s", teamKey, dateFormat)
	res, err := s.Get(url)
	// string is date, of format y-m-d

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

/*

	Team Resource

*/

type TeamResourcesMeta struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   int    `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs"`
	Managers         []struct {
		ManagerID      string `json:"manager_id"`
		Nickname       string `json:"nickname"`
		GUID           string `json:"guid"`
		IsCommissioner string `json:"is_commissioner"`
	} `json:"managers"`
}
type TeamResourcesMetaResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
	} `xml:"team"`
}

func (s *Service) GetTeamResourcesMeta(teamKey string) (*TeamResourcesMeta, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/metadata", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type TeamResourcesStats struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   int    `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs"`
	Managers         []struct {
		ManagerID      string `json:"manager_id"`
		Nickname       string `json:"nickname"`
		GUID           string `json:"guid"`
		IsCommissioner string `json:"is_commissioner"`
	} `json:"managers"`
}
type TeamResourcesStatsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
		TeamPoints struct {
			Text         string `xml:",chardata"`
			CoverageType string `xml:"coverage_type"`
			Season       string `xml:"season"`
			Total        string `xml:"total"`
		} `xml:"team_points"`
	} `xml:"team"`
}

func (s *Service) GetTeamResourcesStats(teamKey string, dateString string) (*TeamResourcesStats, error) {

	var dateFormat string
	if len(dateString) > 2 {
		dateFormat = fmt.Sprintf(";date=%s", dateString)
	} else {
		dateFormat = fmt.Sprintf(";week=%s", dateString)
	}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/stats%s", teamKey, dateFormat)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type TeamResourcesStandings struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   int    `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs"`
	Managers         []struct {
		ManagerID      string `json:"manager_id"`
		Nickname       string `json:"nickname"`
		GUID           string `json:"guid"`
		IsCommissioner string `json:"is_commissioner"`
	} `json:"managers"`
	Standings struct {
		Rank          int `json:"rank"`
		OutcomeTotals struct {
			Wins       string `json:"wins"`
			Losses     string `json:"losses"`
			Ties       string `json:"ties"`
			Percentage string `json:"percentage"`
		} `json:"outcome_totals"`
		GamesBack string `json:"games_back"`
	} `json:"standings"`
}
type TeamResourcesStandingsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
		TeamStandings struct {
			Text          string `xml:",chardata"`
			Rank          string `xml:"rank"`
			PlayoffSeed   string `xml:"playoff_seed"`
			OutcomeTotals struct {
				Text       string `xml:",chardata"`
				Wins       string `xml:"wins"`
				Losses     string `xml:"losses"`
				Ties       string `xml:"ties"`
				Percentage string `xml:"percentage"`
			} `xml:"outcome_totals"`
			Streak struct {
				Text  string `xml:",chardata"`
				Type  string `xml:"type"`
				Value string `xml:"value"`
			} `xml:"streak"`
			PointsFor     string `xml:"points_for"`
			PointsAgainst string `xml:"points_against"`
		} `xml:"team_standings"`
	} `xml:"team"`
}

func (s *Service) GetTeamResourcesStandings(teamKey string) (*TeamResourcesStandings, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/standings", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type TeamResourcesRoster struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   int    `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs"`
	Managers         []struct {
		ManagerID      string `json:"manager_id"`
		Nickname       string `json:"nickname"`
		GUID           string `json:"guid"`
		IsCommissioner string `json:"is_commissioner"`
	} `json:"managers"`
	Roster []struct {
		PlayerKey string `json:"player_key"`
		PlayerID  string `json:"player_id"`
		Name      struct {
			Full       string `json:"full"`
			First      string `json:"first"`
			Last       string `json:"last"`
			ASCIIFirst string `json:"ascii_first"`
			ASCIILast  string `json:"ascii_last"`
		} `json:"name"`
		EditorialPlayerKey    string   `json:"editorial_player_key,omitempty"`
		EditorialTeamKey      string   `json:"editorial_team_key,omitempty"`
		EditorialTeamFullName string   `json:"editorial_team_full_name,omitempty"`
		EditorialTeamAbbr     string   `json:"editorial_team_abbr,omitempty"`
		UniformNumber         string   `json:"uniform_number,omitempty"`
		DisplayPosition       string   `json:"display_position,omitempty"`
		Headshot              string   `json:"headshot"`
		IsUndroppable         string   `json:"is_undroppable"`
		PositionType          string   `json:"position_type"`
		EligiblePositions     []string `json:"eligible_positions"`
	} `json:"roster"`
}
type TeamResourcesRosterResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
		Roster struct {
			Text         string `xml:",chardata"`
			CoverageType string `xml:"coverage_type"`
			Week         string `xml:"week"`
			IsEditable   string `xml:"is_editable"`
			Players      struct {
				Text   string `xml:",chardata"`
				Count  string `xml:"count,attr"`
				Player []struct {
					Text      string `xml:",chardata"`
					PlayerKey string `xml:"player_key"`
					PlayerID  string `xml:"player_id"`
					Name      struct {
						Text       string `xml:",chardata"`
						Full       string `xml:"full"`
						First      string `xml:"first"`
						Last       string `xml:"last"`
						AsciiFirst string `xml:"ascii_first"`
						AsciiLast  string `xml:"ascii_last"`
					} `xml:"name"`
					EditorialPlayerKey    string `xml:"editorial_player_key"`
					EditorialTeamKey      string `xml:"editorial_team_key"`
					EditorialTeamFullName string `xml:"editorial_team_full_name"`
					EditorialTeamAbbr     string `xml:"editorial_team_abbr"`
					ByeWeeks              struct {
						Text string `xml:",chardata"`
						Week string `xml:"week"`
					} `xml:"bye_weeks"`
					UniformNumber   string `xml:"uniform_number"`
					DisplayPosition string `xml:"display_position"`
					Headshot        struct {
						Text string `xml:",chardata"`
						URL  string `xml:"url"`
						Size string `xml:"size"`
					} `xml:"headshot"`
					ImageURL          string `xml:"image_url"`
					IsUndroppable     string `xml:"is_undroppable"`
					PositionType      string `xml:"position_type"`
					PrimaryPosition   string `xml:"primary_position"`
					EligiblePositions struct {
						Text     string   `xml:",chardata"`
						Position []string `xml:"position"`
					} `xml:"eligible_positions"`
					SelectedPosition struct {
						Text         string `xml:",chardata"`
						CoverageType string `xml:"coverage_type"`
						Week         string `xml:"week"`
						Position     string `xml:"position"`
						IsFlex       string `xml:"is_flex"`
					} `xml:"selected_position"`
					IsEditable               string `xml:"is_editable"`
					HasPlayerNotes           string `xml:"has_player_notes"`
					PlayerNotesLastTimestamp string `xml:"player_notes_last_timestamp"`
					Status                   string `xml:"status"`
					StatusFull               string `xml:"status_full"`
				} `xml:"player"`
			} `xml:"players"`
		} `xml:"roster"`
	} `xml:"team"`
}

func (s *Service) GetTeamResourcesRoster(teamKey string) (*TeamResourcesRoster, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/roster", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type TeamResourcesDraftResults struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   int    `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs"`
	Managers         []struct {
		ManagerID      string `json:"manager_id"`
		Nickname       string `json:"nickname"`
		GUID           string `json:"guid"`
		IsCommissioner string `json:"is_commissioner"`
	} `json:"managers"`
	DraftResults []struct {
		Pick      int    `json:"pick"`
		Round     int    `json:"round"`
		Cost      string `json:"cost"`
		TeamKey   string `json:"team_key"`
		PlayerKey string `json:"player_key"`
	} `json:"draft_results"`
}
type TeamResourcesDraftResultsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
		DraftResults struct {
			Text        string `xml:",chardata"`
			Count       string `xml:"count,attr"`
			DraftResult []struct {
				Text      string `xml:",chardata"`
				Pick      string `xml:"pick"`
				Round     string `xml:"round"`
				TeamKey   string `xml:"team_key"`
				PlayerKey string `xml:"player_key"`
			} `xml:"draft_result"`
		} `xml:"draft_results"`
	} `xml:"team"`
}

func (s *Service) GetTeamResourcesDraftResults(teamKey string) (*TeamResourcesDraftResults, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/draftresults", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type TeamResourcesMatchups struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   int    `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs"`
	Managers         []struct {
		ManagerID      string `json:"manager_id"`
		Nickname       string `json:"nickname"`
		GUID           string `json:"guid"`
		IsCommissioner string `json:"is_commissioner"`
	} `json:"managers"`
	Matchups []struct {
		Week          string `json:"week"`
		WeekStart     string `json:"week_start"`
		WeekEnd       string `json:"week_end"`
		Status        string `json:"status"`
		IsPlayoffs    string `json:"is_playoffs"`
		IsConsolation string `json:"is_consolation"`
		IsTied        int    `json:"is_tied"`
		WinnerTeamKey string `json:"winner_team_key,omitempty"`
		Teams         []struct {
			TeamKey          string `json:"team_key"`
			TeamID           string `json:"team_id"`
			Name             string `json:"name"`
			URL              string `json:"url"`
			TeamLogo         string `json:"team_logo"`
			WaiverPriority   int    `json:"waiver_priority"`
			NumberOfMoves    string `json:"number_of_moves"`
			NumberOfTrades   int    `json:"number_of_trades"`
			ClinchedPlayoffs int    `json:"clinched_playoffs"`
			Managers         []struct {
				ManagerID      string `json:"manager_id"`
				Nickname       string `json:"nickname"`
				GUID           string `json:"guid"`
				IsCommissioner string `json:"is_commissioner"`
			} `json:"managers"`
			Points struct {
				CoverageType string `json:"coverage_type"`
				Week         string `json:"week"`
				Total        string `json:"total"`
			} `json:"points"`
			Stats []struct {
				StatID string `json:"stat_id"`
				Value  string `json:"value"`
			} `json:"stats"`
		} `json:"teams"`
	} `json:"matchups"`
}
type TeamResourcesMatchupsResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Team        struct {
		Text                  string `xml:",chardata"`
		TeamKey               string `xml:"team_key"`
		TeamID                string `xml:"team_id"`
		Name                  string `xml:"name"`
		IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
		URL                   string `xml:"url"`
		TeamLogos             struct {
			Text     string `xml:",chardata"`
			TeamLogo struct {
				Text string `xml:",chardata"`
				Size string `xml:"size"`
				URL  string `xml:"url"`
			} `xml:"team_logo"`
		} `xml:"team_logos"`
		WaiverPriority string `xml:"waiver_priority"`
		NumberOfMoves  string `xml:"number_of_moves"`
		NumberOfTrades string `xml:"number_of_trades"`
		RosterAdds     struct {
			Text          string `xml:",chardata"`
			CoverageType  string `xml:"coverage_type"`
			CoverageValue string `xml:"coverage_value"`
			Value         string `xml:"value"`
		} `xml:"roster_adds"`
		ClinchedPlayoffs  string `xml:"clinched_playoffs"`
		LeagueScoringType string `xml:"league_scoring_type"`
		HasDraftGrade     string `xml:"has_draft_grade"`
		DraftGrade        string `xml:"draft_grade"`
		DraftRecapURL     string `xml:"draft_recap_url"`
		Managers          struct {
			Text    string `xml:",chardata"`
			Manager struct {
				Text           string `xml:",chardata"`
				ManagerID      string `xml:"manager_id"`
				Nickname       string `xml:"nickname"`
				Guid           string `xml:"guid"`
				IsCommissioner string `xml:"is_commissioner"`
				IsCurrentLogin string `xml:"is_current_login"`
				Email          string `xml:"email"`
				ImageURL       string `xml:"image_url"`
			} `xml:"manager"`
		} `xml:"managers"`
		Matchups struct {
			Text    string `xml:",chardata"`
			Count   string `xml:"count,attr"`
			Matchup []struct {
				Text                    string `xml:",chardata"`
				Week                    string `xml:"week"`
				WeekStart               string `xml:"week_start"`
				WeekEnd                 string `xml:"week_end"`
				Status                  string `xml:"status"`
				IsPlayoffs              string `xml:"is_playoffs"`
				IsConsolation           string `xml:"is_consolation"`
				IsMatchupRecapAvailable string `xml:"is_matchup_recap_available"`
				MatchupRecapURL         string `xml:"matchup_recap_url"`
				MatchupRecapTitle       string `xml:"matchup_recap_title"`
				MatchupGrades           struct {
					Text         string `xml:",chardata"`
					MatchupGrade []struct {
						Text    string `xml:",chardata"`
						TeamKey string `xml:"team_key"`
						Grade   string `xml:"grade"`
					} `xml:"matchup_grade"`
				} `xml:"matchup_grades"`
				IsTied        string `xml:"is_tied"`
				WinnerTeamKey string `xml:"winner_team_key"`
				Teams         struct {
					Text  string `xml:",chardata"`
					Count string `xml:"count,attr"`
					Team  []struct {
						Text                  string `xml:",chardata"`
						TeamKey               string `xml:"team_key"`
						TeamID                string `xml:"team_id"`
						Name                  string `xml:"name"`
						IsOwnedByCurrentLogin string `xml:"is_owned_by_current_login"`
						URL                   string `xml:"url"`
						TeamLogos             struct {
							Text     string `xml:",chardata"`
							TeamLogo struct {
								Text string `xml:",chardata"`
								Size string `xml:"size"`
								URL  string `xml:"url"`
							} `xml:"team_logo"`
						} `xml:"team_logos"`
						WaiverPriority string `xml:"waiver_priority"`
						NumberOfMoves  string `xml:"number_of_moves"`
						NumberOfTrades string `xml:"number_of_trades"`
						RosterAdds     struct {
							Text          string `xml:",chardata"`
							CoverageType  string `xml:"coverage_type"`
							CoverageValue string `xml:"coverage_value"`
							Value         string `xml:"value"`
						} `xml:"roster_adds"`
						ClinchedPlayoffs  string `xml:"clinched_playoffs"`
						LeagueScoringType string `xml:"league_scoring_type"`
						HasDraftGrade     string `xml:"has_draft_grade"`
						DraftGrade        string `xml:"draft_grade"`
						DraftRecapURL     string `xml:"draft_recap_url"`
						Managers          struct {
							Text    string `xml:",chardata"`
							Manager struct {
								Text           string `xml:",chardata"`
								ManagerID      string `xml:"manager_id"`
								Nickname       string `xml:"nickname"`
								Guid           string `xml:"guid"`
								IsCommissioner string `xml:"is_commissioner"`
								IsCurrentLogin string `xml:"is_current_login"`
								Email          string `xml:"email"`
								ImageURL       string `xml:"image_url"`
							} `xml:"manager"`
						} `xml:"managers"`
						WinProbability string `xml:"win_probability"`
						TeamPoints     struct {
							Text         string `xml:",chardata"`
							CoverageType string `xml:"coverage_type"`
							Week         string `xml:"week"`
							Total        string `xml:"total"`
						} `xml:"team_points"`
						TeamProjectedPoints struct {
							Text         string `xml:",chardata"`
							CoverageType string `xml:"coverage_type"`
							Week         string `xml:"week"`
							Total        string `xml:"total"`
						} `xml:"team_projected_points"`
					} `xml:"team"`
				} `xml:"teams"`
			} `xml:"matchup"`
		} `xml:"matchups"`
	} `xml:"team"`
}

func (s *Service) GetTeamResourcesMatchups(teamKey string) (*TeamResourcesMatchups, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/matchups", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

/*
	transactions
*/

type TransactionsResourcesMeta struct {
	TransactionKey string `json:"transaction_key"`
	TransactionID  string `json:"transaction_id"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	Timestamp      string `json:"timestamp"`
	Players        []struct {
		PlayerKey string `json:"player_key"`
		PlayerID  string `json:"player_id"`
		Name      struct {
			Full       string `json:"full"`
			First      string `json:"first"`
			Last       string `json:"last"`
			ASCIIFirst string `json:"ascii_first"`
			ASCIILast  string `json:"ascii_last"`
		} `json:"name"`
		EditorialTeamAbbr string `json:"editorial_team_abbr"`
		DisplayPosition   string `json:"display_position"`
		PositionType      string `json:"position_type"`
		TransactionData   struct {
			Type                string `json:"type"`
			SourceType          string `json:"source_type"`
			DestinationType     string `json:"destination_type"`
			DestinationTeamKey  string `json:"destination_team_key"`
			DestinationTeamName string `json:"destination_team_name"`
			SourceTeamKey       string `json:"source_team_key"`
			SourceTeamName      string `json:"source_team_name"`
		} `json:"transaction_data,omitempty"`
	} `json:"players"`
}

func (s *Service) GetTransactionsResourcesMeta(teamKey string) (*TransactionsResourcesMeta, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/transaction/%s/meta", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}

type TransactionsResourcesPlayers struct {
	TransactionKey string `json:"transaction_key"`
	TransactionID  string `json:"transaction_id"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	Timestamp      string `json:"timestamp"`
	Players        []struct {
		PlayerKey string `json:"player_key"`
		PlayerID  string `json:"player_id"`
		Name      struct {
			Full       string `json:"full"`
			First      string `json:"first"`
			Last       string `json:"last"`
			ASCIIFirst string `json:"ascii_first"`
			ASCIILast  string `json:"ascii_last"`
		} `json:"name"`
		EditorialTeamAbbr string `json:"editorial_team_abbr"`
		DisplayPosition   string `json:"display_position"`
		PositionType      string `json:"position_type"`
		TransactionData   struct {
			Type                string `json:"type"`
			SourceType          string `json:"source_type"`
			DestinationType     string `json:"destination_type"`
			DestinationTeamKey  string `json:"destination_team_key"`
			DestinationTeamName string `json:"destination_team_name"`
			SourceTeamKey       string `json:"source_team_key"`
			SourceTeamName      string `json:"source_team_name"`
		} `json:"transaction_data,omitempty"`
	} `json:"players"`
}
type TransactionsResourcesPlayersResponse struct {
	XMLName     xml.Name `xml:"fantasy_content"`
	Text        string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
	URI         string   `xml:"uri,attr"`
	Time        string   `xml:"time,attr"`
	Copyright   string   `xml:"copyright,attr"`
	RefreshRate string   `xml:"refresh_rate,attr"`
	Yahoo       string   `xml:"yahoo,attr"`
	Xmlns       string   `xml:"xmlns,attr"`
	Transaction struct {
		Text           string `xml:",chardata"`
		TransactionKey string `xml:"transaction_key"`
		TransactionID  string `xml:"transaction_id"`
		Type           string `xml:"type"`
		Status         string `xml:"status"`
		Timestamp      string `xml:"timestamp"`
		Players        struct {
			Text   string `xml:",chardata"`
			Count  string `xml:"count,attr"`
			Player []struct {
				Text      string `xml:",chardata"`
				PlayerKey string `xml:"player_key"`
				PlayerID  string `xml:"player_id"`
				Name      struct {
					Text       string `xml:",chardata"`
					Full       string `xml:"full"`
					First      string `xml:"first"`
					Last       string `xml:"last"`
					AsciiFirst string `xml:"ascii_first"`
					AsciiLast  string `xml:"ascii_last"`
				} `xml:"name"`
				EditorialTeamAbbr string `xml:"editorial_team_abbr"`
				DisplayPosition   string `xml:"display_position"`
				PositionType      string `xml:"position_type"`
				TransactionData   struct {
					Text                string `xml:",chardata"`
					Type                string `xml:"type"`
					SourceType          string `xml:"source_type"`
					DestinationType     string `xml:"destination_type"`
					DestinationTeamKey  string `xml:"destination_team_key"`
					DestinationTeamName string `xml:"destination_team_name"`
					SourceTeamKey       string `xml:"source_team_key"`
					SourceTeamName      string `xml:"source_team_name"`
				} `xml:"transaction_data"`
			} `xml:"player"`
		} `xml:"players"`
	} `xml:"transaction"`
}

func (s *Service) GetTransactionsResourcesPlayers(teamKey string) (*TransactionsResourcesPlayers, error) {

	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/transaction/%s/players", teamKey)
	res, err := s.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return nil, errors.New("not implemented")
}
