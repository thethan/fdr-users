package draft

import (
	"github.com/stretchr/testify/assert"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"testing"
)

func Test_getRound(t *testing.T) {
	league := entities.League{
		DraftOrder: []string{
			"a", "b", "c", "d",
			"e", "f", "g", "h",
			"i", "j",
		},
		Settings: &entities.LeagueSettings{
			LeagueKey:                  "",
			LeagueID:                   0,
			Name:                       "",
			URL:                        "",
			LogoURL:                    "",
			Password:                   "",
			DraftStatus:                "",
			NumTeams:                   0,
			EditKey:                    "",
			WeeklyDeadline:             "",
			LeagueUpdateTimestamp:      "",
			LeagueType:                 "",
			Renew:                      "",
			Renewed:                    "",
			IrisGroupChatID:            "",
			ShortInvitationURL:         "",
			AllowAddToDlExtraPos:       "",
			IsProLeague:                "",
			IsCashLeague:               "",
			CurrentWeek:                "",
			StartWeek:                  "",
			StartDate:                  "",
			EndWeek:                    "",
			EndDate:                    "",
			GameCode:                   "",
			Season:                     "",
			ID:                         0,
			DraftType:                  "",
			IsAuctionDraft:             false,
			ScoringType:                "",
			PersistentURL:              "",
			UsesPlayoff:                "",
			HasPlayoffConsolationGames: false,
			PlayoffStartWeek:           "",
			UsesPlayoffReseeding:       false,
			UsesLockEliminatedTeams:    false,
			NumPlayoffTeams:            0,
			NumPlayoffConsolationTeams: 0,
			UsesRosterImport:           false,
			RosterImportDeadline:       "",
			WaiverType:                 "",
			WaiverRule:                 "",
			UsesFaab:                   false,
			DraftTime:                  "",
			PostDraftPlayers:           "",
			MaxTeams:                   "",
			WaiverTime:                 "",
			TradeEndDate:               "",
			TradeRatifyType:            "",
			TradeRejectTime:            "",
			PlayerPool:                 "",
			CantCutList:                "",
			IsPubliclyViewable:         false,
			RosterPositions:            []entities.RosterPosition{
				{Count: 1}, {Count: 3}, {Count: 2}, {Count: 1}, {Count: 1}, {Count: 1}, {Count: 6},
			},
			StatCategories:             nil,
			StatModifiers:              nil,
			MaxAdds:                    0,
			SeasonType:                 "",
			MinInningsPitched:          "",
			UsesFractalPoints:          false,
			UsesNegativePoints:         false,
		},
	}
	round := getRound(1, league)
	assert.Equal(t, 1, round)

	round = getRound(10, league)
	assert.Equal(t, 1, round)
	
	round = getRound(11, league)
	assert.Equal(t, 2, round)
}