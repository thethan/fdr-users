package league

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/draft/repositories"
	"github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/test_helpers"
	entities2 "github.com/thethan/fdr-users/pkg/users/entities"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"os"
	"reflect"
	"testing"
)

type mockGetUserInformation struct {
	mock.Mock
}

func getUser() entities2.User {
	return entities2.User{
		AccessToken: "BJk9GkGY4x.TiR2hB27H2sF06b3XsuhJliHlS62udUht3bdpe6YNFiYDrgXQSSGr8i8ztsOvdwWgHftXNf.dvqJ1mFYcaGIB8AG_v3FtAS0hahE7oY.YSJ2TC8R1e7kL3mShi_rC1_Hp5yP76H.e2dqNIoju7ls0vxlkS9Uzkds28t8j25zuOvQq4buiWwYN1dQMaoX5wQUJv4A9H2cPX6rC8LEZH3rvDQwrEzYnmeCob_mY37GBm5VMkXuaunp.wq9CiuFiyti8SgRPspNMwbe7WJ9IQ1UC7L6btP1Gx6ifytkg1BjxeptH1n0IqUFWbqVQ_eX0JFjOyGcFTtemikV3F2trtzRfW5uKQ0LmCAtgwVSwrckEw12v4iRYglMgiwJ2EXBOapp8Vezmf1EjGy4vAQze1SQdZQoDf33Oz4CiU1igk_x6AwC3lQQ7knbslG4HnNgS8hObZG9y8bx7BNaz7uXxH__n5pCp9YGQ63fVdnS0lg6G9ULYdJvEvYPnIFoDtgS9u.28qYRH2z15MoSeB0JsFubnjKvZzStNj5xg0pYwHqeRqH2UQEMivJDzl_vDqA5ndRyHhioLUcZWT7Hhajj3JchB4fb1eX.JKOSfgsupnClNPjJL2Af8f45D9U9WfTNDhtb.X19QKl1ssCInDmkdiezEsMday5OXmncfgk9eUUctTSn.y0yZAJUB0RYe.A_Yd7IMDAtqq6i2hMj23XkB17NWNAK4nKQjVe.UJl4JISBvcJlz6YrPJP6NC55wGcc_ctzNpOVQNTtdkicUnRv7Q8uz.EnVBHK77tJA8cd1AJ4PT_mCIvOlK7k7sK8VBuawsr3GEsiMQQNhvcSZYrOz.OiW7IYY0d8hneU3AKIA8KaakeNikAgbAfL9gDGusxr_xS8Ry5_KOpvtsCWAD4KoZ4l4jlMqO30iQyJhVT6Wylv4fbV3ZeDfcrKAg0XbzxXQU2.vdLKc0LTZZ0xwLPKW3xb28xISjcYDL0H09NXfeMsZ5jprgN7FbJwlkNN1uWkYPcfuIyh2oYixsIp9S0NUhzQFc9RBg8hsgq2ouLbBCu4_kSJBof8mO136uO_LKWelKfT47jUPa_SEyUoTSsD1Urx48yWygg5V",
	}
}

func (m *mockGetUserInformation) GetCredentialInformation(ctx context.Context, session string) (entities2.User, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(entities2.User), args.Error(1)
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

func TestImporter_ImportGamePlayers(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	ctx := context.Background()
	tracer, err := apm.NewTracer(t.Name(), "0")
	if tracer == nil || err != nil {
		t.Fail()
		t.FailNow()
	}
	tx := tracer.StartTransaction(t.Name(), "0")
	span, ctx := apm.StartSpanOptions(ctx, "TestImporter_ImportGamePlayers", "test", apm.SpanOptions{
		Parent: tx.TraceContext(),
	})
	defer span.End()
	defer tx.End()

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

	err = svc.ImportGamePlayers(ctx, 390)
	assert.Nil(t, err)
}
