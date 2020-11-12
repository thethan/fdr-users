package league

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thethan/fdr-users/internal/test_helpers"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/draft/repositories"
	"github.com/thethan/fdr-users/pkg/mongo"
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
		AccessToken: "3oxRd8Sd7gCQ1_5hI6It6aon70cfgK_yg5PjVt4mnCIhFSeMZ6u9SUNlweofBLten53DznzPyIe._wPbKSKZapFlXGvgmLwJewZJqP_WwG3TcUzZfIPEY5wOcEMQBwsmExsyoj_Eb5uzY69Ptp6tgVfORfyhFY1Kl4HsHljUbzPhMW_2m97uMcubv5heQqjtOwv5tHSqXadUMF4A8ENS.txbjPmZIu2KWwQkR7rwYvE8WxRucaXX6K0KG.Hu8gDgkUFJfM67CzK25PaH0DI7dG1X.2DtDrcSDF85xVy8vObWKHKejXEUdHllnxEpUQbJ3wN3Cb2oYX8b77f50y00jki.J5.F58Df3wNLFZQvVmWF_2_eCueBbPgXigLsnDGw1ltlDgS5vtmxvqe_c8HGPCjOAvkaeaBk2qx10Zwcpt9yiBdv1t1.3XWl3F3nL9qsoIk1u1bk.ldCkZxnh1YItKIxwdy1y_aE5SBxidr7NAVuhyeRv0ZRI0vv7Fi9tMoflQfKIe_fWwvJOQ.s.2qjp3nSUSGPwrHgYbAt4QrcYqXvclJ5IJt5EsCQmpS6X5r3L5O33ujzRRs9gR07m2_VuYajr6wFB4OeMpSGXL0edxJoncOJP5vmS46nOVvHCW3NsHFN5z5gZ_QYBo5ifihWy2qTJnzL6Cd2IfhwWRxiSZGFWAnDRT1yz95QDazsMSM.fF1WHAaxYoq7xwldSiTQtLeTswejOj8AsscB7vnNozblYg0.jq0kUKUVxHSueXl21zjmy6Euew005YSukQTaNaiFLcdyGUd01HP56Joed4KVO7CvA54QtEuTH2dedWkNwkTzVyl1yYFRX2HWFcSZoD.4YVMfnivAxl92EHg.qiJ3w4_5xcLmCGm8lbR5IJ0o1JQKi2BhmN7_d5M4d58tTbXEFlRs3IpMwdbWZzSetwza7geVeLrIT_F6CD2nd.dg189s76TYh6Fyokx.d1PyGGGrLQkIXH3VAxdq1PQEiJnV7dlxCN3CYHBol5gviOseYQtym_YdJszvWVtix2BY74L8DeMpLLmeIlB8fvGTLIdfBeVAlFDMBG1hhDYV5GJr4yU.bLZQ8yHAgHDLzpJ4Y_h6Lrl1tC4YjCBVjuieug--",
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

	mongoClient, err := mongo.NewMongoDBClient(context.TODO(),os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo, nil)

	leagueGroups, err := svc.ImportLeagueFromUser(ctx)
	assert.Nil(t, err)
	assert.True(t, len(leagueGroups) > 0, "league groups is not greater than 0")
	if t.Failed() {
		t.FailNow()
	}
	for idx := range leagueGroups {
		assert.True(t, len(leagueGroups[idx].Leagues) > 0, "league groups do not have league")
		// add league group to the first one
		//nLeague, err := mongoRepo.SaveLeagueLeagueGroup(ctx, leagueGroup)
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

	mongoClient, err := mongo.NewMongoDBClient(context.TODO(),os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"),os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo, nil)

	err = svc.ImportGamePlayers(ctx, 399)
	assert.Nil(t, err)
}



func TestImporter_ImportDraftResult(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	ctx := context.Background()
	tracer, err := apm.NewTracer(t.Name(), "0")
	if tracer == nil || err != nil {
		t.Fail()
		t.FailNow()
	}
	tx := tracer.StartTransaction(t.Name(), "0")
	span, ctx := apm.StartSpanOptions(ctx, "TestImporter_ImportDraftResults", "test", apm.SpanOptions{
		Parent: tx.TraceContext(),
	})
	defer span.End()
	defer tx.End()

	mockUserInfo := mockGetUserInformation{}

	mockUserInfo.On("GetCredentialInformation", mock.AnythingOfType(reflect.TypeOf(ctx).String()), mock.AnythingOfType("string")).Return(getUser(), nil)

	yahooService := yahoo.NewService(logger, &mockUserInfo)

	mongoClient, err := mongo.NewMongoDBClient(context.TODO(),os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo, nil)

	_, err = svc.ImportDraftResults(ctx, "380.l.53275")
	assert.Nil(t, err)
}



func TestImporter_ImportDraftResultForUser(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	ctx := context.Background()
	tracer, err := apm.NewTracer(t.Name(), "0")
	if tracer == nil || err != nil {
		t.Fail()
		t.FailNow()
	}
	tx := tracer.StartTransaction(t.Name(), "0")
	span, ctx := apm.StartSpanOptions(ctx, "TestImporter_ImportDraftResults", "test", apm.SpanOptions{
		Parent: tx.TraceContext(),
	})
	defer span.End()
	defer tx.End()

	mockUserInfo := mockGetUserInformation{}

	mockUserInfo.On("GetCredentialInformation", mock.AnythingOfType(reflect.TypeOf(ctx).String()), mock.AnythingOfType("string")).Return(getUser(), nil)

	yahooService := yahoo.NewService(logger, &mockUserInfo)

	mongoClient, err := mongo.NewMongoDBClient(context.TODO(),os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo, nil)

	 err = svc.ImportDraftResultsForUser(ctx, "MFG5HMFDHC634Q7W2FPKJBVTKY")
	assert.Nil(t, err)
}



func TestImporter_ImportPlayersFromGamesForUser(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	ctx := context.Background()
	tracer, err := apm.NewTracer(t.Name(), "0")
	if tracer == nil || err != nil {
		t.Fail()
		t.FailNow()
	}
	tx := tracer.StartTransaction(t.Name(), "0")
	span, ctx := apm.StartSpanOptions(ctx, "TestImporter_ImportPlayersFromGamesForUser", "test", apm.SpanOptions{
		Parent: tx.TraceContext(),
	})
	defer span.End()
	defer tx.End()

	mockUserInfo := mockGetUserInformation{}

	mockUserInfo.On("GetCredentialInformation", mock.AnythingOfType(reflect.TypeOf(ctx).String()), mock.AnythingOfType("string")).Return(getUser(), nil)

	yahooService := yahoo.NewService(logger, &mockUserInfo)

	mongoClient, err := mongo.NewMongoDBClient(context.TODO(),os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo, nil)

	err = svc.ImportGamePlayersUserHasAccessTo(ctx, "MFG5HMFDHC634Q7W2FPKJBVTKY")
	assert.Nil(t, err)
}

