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
		AccessToken: "WnFFO7WY4x_0EL6r9tPgpqQazK7oQ.LLIpHq6DD33VQ_EdGBfoMWMEAC5h26FjRMtJOfShnocLVAyeJTuKewYJH6Ugf0rvx6V.xxwa9Gwx.F2h1yiYRAa9wMtKp_q2b_wc9kqY4260aZ.C3ZZyk9q83eaC8QBPOa37K3MdEnsb0NDxbSkcJGypEU_x729sUuwLmWcSJ_QYzPaaobW3MhM9TSiEWwP5oOAqPOiXHhYbbeBdSAfLpcCGk9clFCGlWpcBOOea7KS7xf8c7HFPUEy2ZbsA4e5pttJvIPai4NDW6akB3XYaISbr0aYQlrg7P3_9kA1TQKNYS7VtfMHLQN8VBtG9kEu2tnxN2QN2zt.zsOYxjLBm_jr2fcA7yy8pybnpHqTpuakEAdf.P5k5Vy_AFle71uQRh43PKoG1_cw2ff.SnhScDSypiPG.u8idl6VsRNg1JM6nTSXb.xrEmBNKbX.CHtGGI_8LCV2_db6GT_IxPBYgMc7mVVeOjTvhbsjdFyYT6i4lZP2BrpYv_teV6uOpDguCpb6rhIf616cZOSjWfL7xnnfRvyKII27mP8YypOLuZw.0msVQ8wrv6kjZSKf3mECE5RSKZQfsXTIgFo_PKm1FhBL03iMt_DY4dJoy2ZbFwmhC5Vo8r7FyAtRtsr4ARtTMXxPGvki1uotmxWlRlBlUpPM2Pie6hGrxnpz0kXx8LHAqibOtcO2J8p2tkAq0Zg4apdpdUkVGRYcfCbsnH5nr6LMxU0FPKVGlYBovj1ClrAmxm6gW0zKLsdIcU8kpdbddR.PWFWGRHHmq.nZUzQbCa0Yb4.m6v2xe8mgp9vMqapm.CcmYcwyVX8dHkX6X_feen4H3eWRNxFQeuyf7a4FYRvNcz3lxgJIEQJOrshIwVVTZ5zXw8GaQhU6oCS6fVm0fmfKXaiG0XBSukb3DvAmGPxjgEU.eVFY3_xFpEUhXi0wk8_RLf3sXbnffhogGGtCmK.8CNwGkcYlbUt3REYeU.gE_8eCe7ydZiUgewPAGWLsntaZXN2s8Tq1Ur80PcVPOVjfVC_S863uoPvHeiJWDJCSYGNSasTFAqGLJArtTzgudRNTy0QIkQzLy08piWWyedwwB8EhipO",
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

	err = svc.ImportGamePlayers(ctx, 273)
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

	mongoClient, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo)

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

	mongoClient, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		t.FailNow()
	}

	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	svc := NewImportService(logger, yahooService, &mongoRepo)

	 err = svc.ImportDraftResultsForUser(ctx, "MFG5HMFDHC634Q7W2FPKJBVTKY")
	assert.Nil(t, err)
}



// T3 Cancer