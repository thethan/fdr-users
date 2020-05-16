package league

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/draft/repositories"
	"github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/users"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"os"
	"reflect"
	"testing"
)

type mockGetUserInformation struct {
	mock.Mock
}

func getUser() users.User {
	return users.User{
		AccessToken: "tip22zOY4x_EBhkj4XO2ag9PJBcXOsNQUhGk9lxTA8KVRSDMSg6Lbpd6FRbEPLP4EkStuiWWwi9rGmO.wJwCbVHIYkJW.mktZqmgxK4liSu7ZQOPak.GwPwFzVOSkbesJesfDTwlyQ5RC.LbA6V4EjiTvlygqx4NjaeT3dKd1LwlrGSfVGZcYxiURejD3zzyf8ntAH8nHGigItZ.jt2KY4p4B3lwoOqgfypAgztqpVSnOl09OYNtaAc2tWKIDYFr5uBdjJM9BWEY2VOGo6dI511Q1bTxTIMWj.bPBRBuIwnfU.PhYt07Wz6SgGcRbBRcHyVbwqd_UjufbfUBZJY1Zx7rf.JvEKVte7hePO.swbt_E4lKzHicbsCW2.O0lsM6wUhn1oP8EzhBBBRV_1D_PYsCMtA.Pab5T8Q6JEkWfCeyW6QHFx8GZdCXOkrInSigmix8wmwkBCXbck8dsX8u6pQgooTWM25QizzdeFVqgtwTc64g6bEnH6Vu0Kn8wQMoHOIZAqymFIKICvUFUG1MOS0l9mckIrZM2DBKs4BvIFogZDymWsQLtaszcrmv0QjKbjjfzWAmvFV9AE3vzctVmI77EAIYOjae04JoDAkEbZ4rDJmzfkn6qj47_0G7cyi2Gc5iByK3fO71ut5urm9nkwdOu5Ya3zEfN92xC5Og.S633Gb0I6CAPd3MUvv5V6b704lOUuaeMKfm1cM1liCKWHR78WBjJ9sFD2cj_IPd7uFHJSXogDB5P_RRS5e4.HgwTPVI3Ogto5w6NsV1WmCHqmgxH5Nb9t6FA.8lOpBhto8CGqiWryMCwv_R2cZHbCVLtXyUXIkAaXa5l.3OhGDNC33hnZDq86GnTbQUJptlLGztjc_Mt0pVvh7ZWzqD74CaeHnyJLNgo1tpFxfRzAIYJs4cloX_bXA1tetvrv_VRyIBxT_A6qVcDbWXEI3ePqQEh6ngZjEg..tR2Xt1zi472iktAV3FMq8N1lk_HnECGIn0ImknFEFRo1VlV057jXrPbZ9WZ92uLSXu9Y82wotnI2vhL2Bf.6Uy20p9kNXjOGWtXW42vPlJXhf8K3dHenrevElzz9x5M_gwPHtfDELzCEFwPEBVNPG1wK2WWeLY",
	}
}

func (m *mockGetUserInformation) GetCredentialInformation(ctx context.Context, session string) (users.User, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(users.User), args.Error(1)
}

type mockSaveLeagues struct {
	mock.Mock
}

func (m *mockSaveLeagues) SaveLeague(ctx context.Context, group *entities.LeagueGroup) (*entities.LeagueGroup, error) {
	args := m.Called(ctx, group)
	return args.Get(0).(*entities.LeagueGroup), args.Error(1)
}

func TestImporter_ImportLeagueFromUser(t *testing.T) {
	logger := log.NewNopLogger()
	ctx := context.Background()

	mockUserInfo := mockGetUserInformation{}

	mockUserInfo.On("GetCredentialInformation", mock.AnythingOfType(reflect.TypeOf(ctx).String()), mock.AnythingOfType("string")).Return(getUser(), nil)

	yahooService := yahoo.NewService(logger, &mockUserInfo)

	mongoClient, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo)

	leagueGroups, err := svc.ImportLeagueFromUser(ctx)
	assert.Nil(t, err)
	assert.True(t, len(leagueGroups) > 0, "league groups is not greater than 0")
	if t.Failed() {
		t.FailNow()
	}
	for idx := range leagueGroups {
		assert.True(t, len(leagueGroups[idx].Leagues) > 0, "league groups do not have league")
		// add league group to the first one
		//nLeague, err := mongoRepo.SaveLeague(ctx, leagueGroup)
		assert.Nil(t, err)
	}

}
