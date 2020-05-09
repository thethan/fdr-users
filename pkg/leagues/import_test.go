package leagues

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thethan/fdr-users/pkg/users"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"reflect"
	"testing"
)


type mockGetUserInformation struct {
	mock.Mock
}

func getUser() users.User {
	return users.User{
		AccessToken: "ebySnieY4x9_LPfBLJqHKQ.JjGin37CHtKWyzLZf9VRAXBTPNGDKbQ1W8t9aodukgutdgjrvozAXtvIwlF_hc2e7attlM9GrmzneuGAxME.fNILjGkOeaXxJEMK_xmW3TGsxCs4.p50E9qd7y1UIikznR1KW35Tk05f7y5rCfPAoQgQRlneEXWcwXtWJxddlRGUFCLF1fYQVqi2.O8Op9lbxFx5bQIzITpsW5PkVUerZJeIzpaG2fO8aUmvbCbAuuAdC.D12FaIqOwojODNl.h2ejeUudAOXFV1CTWPDfMp4fQl25GsymDxYIuE9748zQ6dK8cHXH_7tfeVVBBRstQNDeanXvO7BN5vFeisuQ4BhiPwYbdqXAFT40YGNZaa0kGFomlFlMaqhCahrQzfbqjwObEYT05ck_3Yfn4anw.qaxXCBNpsHHv1aylFHAosDC.2cRbJpyQRuXMzL8S0y.i44oAmqZdFSJz0wIs58eQa31Jrg8yUfxuzIIlwhK_BBGUlihUkDnMciXZCiSHUMqFlUn2BlWaZDTl5U5GiYjQIY1YR8JdIEPpnz4jROQrpuWP6El2HxAykqxu1CyO92jHK16OxzvixhhAHk2OwduQtxJsiKLJbosfIyrDwB_EuXjb98GTHl1sA0yq27W3JeuxAJQrPl464nfU9eHbUzvYWBWY3Jb7ZDElupJCJtV5DwQKEF8SRdt9G5d3WuhTT.EVmcFgUbzCpxF5sEKlgNWz39tX0x448NRoDdAOnsTpH0qwSA22dY6FzuI_J2kg07s1slCpLTGrs_HFgE9GO0g160SHnMypQQdbCafxUsAx1xDfmQdg1o3VfTgah_X8tLayNRv0QnBTOBgCN7prOqfx99I4XlDPcxbS_0jUjVESGYi3winl.vhLZ7DULaLwPU8Z_F2v1TKcjLT6.yZJJvLmrdsjpquu7dv7Yq5a5FTWzh6F4oYz0dOrjx51qnYThEozm_xVTF1HTZYOJiTgW9KvdBKKQk2zdGUFy0aXWVhVEhi0aleboW4MiyiZ8yPFFQVOUvTseZQuh1u8Btx5c3TqaHqEVm0GG9dubPnQBbplK9O1mfKVyd7LCb3yRKWdEBmVZcFORId4f8.vrtZ4Fcd_z6JwQNsyputvipBZge91yR5BvbcrcHckYyInX..sO6_VfuOqEIMBVCWH9tePE-",
	}
}


func (m *mockGetUserInformation) GetCredentialInformation(ctx context.Context, session string) (users.User, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(users.User), args.Error(1)
}

func startNewSvc(t *testing.T) {
	t.Helper()

}

func TestImporter_ImportLeagueFromUser(t *testing.T) {
	logger := log.NewNopLogger()
	ctx := context.Background()
	mockUserInfo := mockGetUserInformation{}

	mockUserInfo.On("GetCredentialInformation", mock.AnythingOfType(reflect.TypeOf(ctx).String()), mock.AnythingOfType("string")).Return(getUser(), nil)

	yahooService := yahoo.NewService(logger, &mockUserInfo)

	svc := NewImportService(logger, yahooService)

	t.Run("GetUserLeagues", func(t *testing.T) {

		leagueGroups, err := svc.ImportLeagueFromUser(ctx, )
		assert.Nil(t, err)
		assert.True(t, len(leagueGroups) > 0, "league groups is not greater than 0")
		for idx := range leagueGroups {
			assert.True(t, len(leagueGroups[idx].Leagues) > 0, "league groups do not have leagues" )
		}


	})
}
