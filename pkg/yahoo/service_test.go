package yahoo

import (
	"context"
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
		AccessToken: "KEaG9iCY4x_cvW_2ExyUYvwFFesYkOkSZWnvnaUt7GYBKJIjxTnklB2sGwzAC6.GznYOEXHwLsH8ArFoVTFCEeqF7tgitaiHGgHTExMU8AOgwnSegLcxX9BQGcmglEyQoOC7yXaxqIxy6tVPFcbZE.SPaS5LTtKpnv8asWuBxMDBrja8bmQCMzVmoTRMWmHJxmCiaIW8HWnVphqRp4C5WSZWre9HSoSx_4JB0s1DNMLvPC_9b2g.dm7ORi5r9GBZxcgHqP50H.kC4d8PV5x.9pzzrFZrgkMJEukE2xXJUd8ZIwph4rDe3LZcKh8O6LbEJTkprhEc4UI8wdO7TvFHYsSTw8mueXlFSI3MbePzY6HRKKMVYfQmsqBXn4q2IPK8dDZRaACCr0mlBp4tfo87Jhd8fE9PGRf57Zh9xrn3pWclG2ABBEXNMLt3FAlU.2cWT8UOGepNV.uzuoiCukZZhqOg0qWfpGl67vio.nCB6qsn42C7q4alk9MovUcj1wu1MOv5.3d.D..PvYnaaSzZElaPLfMcCPZCJIXuZXT4AnyUOr1NWuUoDOUuTt19V.qvR6xPMsGJdvsooM7.j0RYvn61JWo01KGH8m57s0mc1t3IMJ9ChGJrSKkelxDb8J9WzL6oMeNNSlpdBzxLBuA.RpJlOZXECD2hbr1wrQGHnZI8IsiG9oESl1pYEe7yI9A1VXouM17LF8pmHx0ZkLwajRx6SXnQWKpMAcEBW_CzMoFJZSy2TbCjwazsYAIwt4Tyleuv9sfCzzlP_GfSs8PjxmAqT36B_AuD2nf1.1HWYqUHmGTcFiZZiwWmIuZ7jUlVxirRHcjxj4fPL2hYDwYtCC8dy6Yr8YSWLtbbt9.BR8.NTWjP0XHxmgV5T2gq0Ohj8hm6ct0_DJXlS4tKZluBGy2q_6Vp2q1wFtnku2cW0bUXhlIjm5.Nej6PZdnpqR5Yv8.ifUJm9FgBQAAbHwuNzA8JGLhmA4zkA68v13Hy7x8xj7cREuQfJuCFEWeaJ3VDBuqftFHe7.s1bSyFUmJM1kX0pq6o8hbEPzSk2RFtUR.1zBblsC5EuCa6JvKaiPtrQKuxNqUmKiXoKvQlBvQwboFBc7u0P6r9B8umog8.OdW1FQHfa6N_0cTbmuwhsu4dCiNSnTkOVjciIYSa7jLsBo_6oh2TbzQ8c30p.aIS.Q--",
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

func TestService_GetUserResourcesGameLeagues(t *testing.T) {
	mockUserInfo :=  mockGetUserInformation{}
	logger := log.NewNopLogger()
	ctx := context.Background()

	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)


	svc := NewService(logger, mockUserInfo)
	leagues, err := svc.GetUserResourcesGameLeagues(ctx, "390")
	assert.Nil(t, err)
	mockUserInfo.AssertExpectations(t)
	assert.True(t, len(leagues.GameLeague.Leagues) > 0, "Length of games is not greater than 0")
}

func TestService_GetUserResourcesGameLeaguesSettings(t *testing.T) {
	mockUserInfo :=  mockGetUserInformation{}
	logger := log.NewNopLogger()
	ctx := context.Background()

	mockUserInfo.On("GetCredentialInformation", ctx, mock.AnythingOfType("string")).Return(getUser(), nil)


	svc := NewService(logger, mockUserInfo)
	leagues, err := svc.GetUserResourcesGameLeagues(ctx, "390")
	assert.Nil(t, err)
	mockUserInfo.AssertExpectations(t)

	for _, league := range leagues.GameLeague.Leagues {
		_, err := svc.GetLeagueResourcesSettings(ctx, league.LeagueKey)
		assert.Nil(t, err)
	}

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
