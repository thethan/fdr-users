package yahoo

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thethan/fdr-users/pkg/users"
	"testing"
)

type mockGetUserInformation struct {
	mock.Mock
}

func getUser() users.User {
	return users.User{
		AccessToken: "yobokzCY4x_Btuckuyb67JaTPDi7M.zEvtt_ZmIoKSUPvcc.ayGoBdlLdczPjo3AlD1RbWBf8yfUQ5lXlxXXjOSgRBn9CKIm0O2hKqggqIW9.B4LpczY774ui05fCcQ77SQ57fR2IgRp1XJFqd1CHqagJ1jDUj_yJm.TKiXzdLUYDtaedAR86IOXRbdsKZwTWp08GXo8ijSJgavoVVBURr5lVCy42QoVuc8Kk9rzsRbSJYZ93WtthvA.Bmj1lj.TQ_AYbhlGv.qzlRrJ0ZhlK8Xhd_NFmdDvW5mPE2qTTP0ZytIX0VGLKrnK5AYnhx6el1g3xVtOaH92uKv2CTE1AlDyBz4zln4NFaLczGjJD2.yge_SPv6l5ULukwYZ1W8RK9ijpZv0VE2PUeKXPYQ2fNZRRTd5_exOGlPvy9E.ggg5sC2scCwTKDtZJrUPcPMcoBNex39LrQmJ9M.mdx9b.x8stZtKfa5h8HsNKvkAhNqgOgUSXPGeCQVkWLRt9.K3rmwPxy90xTjNVLaSmwgJWX30w1kgw4X0TA1Ba3y3NG8knxCsF4hGGVTfR.NKXLxb9KrwP7HIrklwqGWSPPhmRPd5TM_flwgioK.BwPn2AHuXwqJrAtqzAQiz3oBdpJSCIlSbFAmvVyH05YAj8S7FpbvoTynRM7AHyuAUGvptkaxT3hc3aKv417C3gSD9EmPa9FfTYt6RMucw_03tazB0G3G_A_G7hjf0zHrPZEPTRq6CQXTyfy1csX0d8oU_WJN.NlgLrEk4KYe59U8TEWhIs4PUpgCg2EPZ2vhp4ycy63ZlWKBscSizCcvZGYfJcpbwz_T4iKZjn3909e4sooxGCqg7RhvFqFMcrKIygdvcTI_rXfw8Tye7bRLNTa.1xJhPdJlxTZKuhOmY2CiXS_LlAtt_HILnTyahrpEJw574YUXM0Lw6dyFEwLd44WnWm9e21TBjKjR6Ddo8.AYWxHXcvSXFaZJyFsMJjFHvT0kSbs4S71isRuP4jsJZEvfHQ.Xmvl1q1HkWd_G13OhnrYKuZ.kR4zeMfpry3H7qFFK7vhTgtp8Cb07UkgrPAL5CiGqQxCYzj9Ts_eTk0egQW9IxOr0MfBy4iLkDFb_bdpF.M.D86y4zFPav.pc14jiS_OoiG.uLq8X443ojPAJSNddCQ1L_FYY7ZHnMjWhgV_ovIA--",
	}
}


func (m mockGetUserInformation) GetCredentialInformation(ctx context.Context, session string) (users.User, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(users.User), args.Error(1)
}

func TestService_GetUserResourcesGames(t *testing.T) {
	mockUserInfo :=  mockGetUserInformation{}
	logger := log.NewNopLogger()
	ctx := context.Background()

	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)


	svc := NewService(logger, mockUserInfo)
	games, err := svc.GetUserResourcesGames(ctx, getUser())
	assert.Nil(t, err)
	mockUserInfo.AssertExpectations(t)
	assert.True(t, len(games.Games) > 0, "Length of games is not greater than 0")
}

func TestService_GetLeaguesUserInfo(t *testing.T) {
	mockUserInfo :=  mockGetUserInformation{}
	logger := log.NewNopLogger()
	ctx := context.Background()

	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)


	svc := NewService(logger, mockUserInfo)
	res, err := svc.GetLeagueResourcesStandings(ctx, "273.l.652120")
	assert.Nil(t, err)
	mockUserInfo.AssertExpectations(t)
	fmt.Printf("%v", res)
	//assert.True(t, len(res.League) > 0, "Length of games is not greater than 0")
}

func TestService_GetLeaguesSettings(t *testing.T) {
	mockUserInfo :=  mockGetUserInformation{}
	logger := log.NewNopLogger()
	ctx := context.Background()

	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)


	svc := NewService(logger, mockUserInfo)
	res, err := svc.GetLeagueResourcesSettings(ctx, "273.l.652120")
	assert.Nil(t, err)
	mockUserInfo.AssertExpectations(t)
	fmt.Printf("%v", res)
	//assert.True(t, len(res.League) > 0, "Length of games is not greater than 0")
}
//
//func TestService_GetUserResourcesGameLeagues(t *testing.T) {
//	mockUserInfo :=  mockGetUserInformation{}
//	logger := log.NewNopLogger()
//	ctx := context.Background()
//
//	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)
//
//
//	svc := NewService(logger, mockUserInfo)
//	leagues, err := svc.GetUserResourcesGameLeagues(ctx, "390")
//	assert.Nil(t, err)
//	mockUserInfo.AssertExpectations(t)
//	assert.True(t, len(leagues.GameLeague.Leagues) > 0, "Length of games is not greater than 0")
//}
//
//func TestService_GetUserResourcesGameLeaguesSettings(t *testing.T) {
//	mockUserInfo :=  mockGetUserInformation{}
//	logger := log.NewNopLogger()
//	ctx := context.Background()
//
//	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)
//
//
//	svc := NewService(logger, mockUserInfo)
//	leagues, err := svc.GetUserResourcesGameLeagues(ctx, "390")
//	assert.Nil(t, err)
//	mockUserInfo.AssertExpectations(t)
//
//	for _, league := range leagues.GameLeague.Leagues {
//		_, err := svc.GetLeagueResourcesSettings(ctx, league.LeagueKey)
//		assert.Nil(t, err)
//		// getLeague Resource
//		_, err = svc.GetLeagueResourcesTeams(ctx, league.LeagueKey)
//		assert.Nil(t, err)
//
//
//	}
//
//}
//
//
//func TestService_GetUserResourcesGameTeams(t *testing.T) {
//	mockUserInfo :=  mockGetUserInformation{}
//	logger := log.NewNopLogger()
//	ctx := context.Background()
//
//	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)
//
//
//	svc := NewService(logger, mockUserInfo)
//	leagues, err := svc.GetUserResourcesGameTeams(ctx, "390")
//	assert.Nil(t, err)
//	mockUserInfo.AssertExpectations(t)
//
//	for _, league := range leagues.GameLeague.Leagues {
//		_, err := svc.GetLeagueResourcesSettings(ctx, league.LeagueKey)
//		assert.Nil(t, err)
//		// getLeague Resource
//		_, err = svc.GetLeagueResourcesTeams(ctx, league.LeagueKey)
//		assert.Nil(t, err)
//
//
//	}
//
//}

func TestService_ImportForAUser(t *testing.T) {
	// Get User Games
}


func TestService_GetGameResourcesPositionTypes(t *testing.T) {
	//mockUserInfo :=  mockGetUserInformation{}
	//logger := log.NewNopLogger()
	//ctx := context.Background()
	//
	//mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)


	//svc := NewService(logger, mockUserInfo)
	//leagues, err := svc.GetGameResourcesPositionTypes(ctx, "390")
	//assert.Nil(t, err)
	//mockUserInfo.AssertExpectations(t)



}
//func TestService_GetUserResourcesGameLeaguesSettings(t *testing.T) {
//	mockUserInfo := mockGetUserInformation{}
//	logger := log.NewNopLogger()
//	ctx := context.Background()
//
//	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)
//
//	svc := NewService(logger, mockUserInfo)
//	leagues, err := svc.GetGameResourcesStatCategories(ctx, "390")
//	assert.Nil(t, err)
//	mockUserInfo.AssertExpectations(t)
//
//	for _, league := range leagues.GameLeague.Leagues {
//		_, err := svc.GetLeagueResourcesSettings(ctx, league.LeagueKey)
//		assert.Nil(t, err)
//	}
//
//}
