package kubemq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kubemq-io/kubemq-go"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
	"strconv"
)

const draftPrefix = "draft"

type Repository struct {
	client *kubemq.Client
	logger log.Logger
}

func NewDraftRepository(logger log.Logger, client *kubemq.Client) Repository {
	return Repository{
		client: client,
		logger: logger,
	}
}

func buildDraftChannel(leagueKey string) string {
	return fmt.Sprintf("%s-%s", draftPrefix, leagueKey)
}

type BroadcastDraftResult struct {
	Message     string               `json:"message"`
	DraftOpen   bool                 `json:"draft_open"`
	Team        entities.Team        `json:"team"`
	League      entities.League      `json:"league"`
	User        entities.User        `json:"user"`
	DraftResult entities.DraftResult `json:"draft_result"`
}

func (r *Repository) BroadCastDraftResult(ctx context.Context, league entities.League, user entities.User, team entities.Team, draftResult entities.DraftResult, pick, round int) error {
	span, ctx := apm.StartSpan(ctx, "BroadCastDraftResult", "kubemq")
	defer span.End()

	broadcast := BroadcastDraftResult{
		Team:        team,
		League:      league,
		User:        user,
		DraftResult: draftResult,
	}

	event := r.client.ES()
	event.AddTag("player_key", draftResult.PlayerKey)
	event.AddTag("pick", strconv.Itoa(draftResult.Pick))
	event.AddTag("round", strconv.Itoa(draftResult.Round))
	event.AddTag("team_key", draftResult.TeamKey)
	event.AddTag("player_id", strconv.Itoa(draftResult.PlayerID))

	channelName := buildDraftChannel(league.LeagueKey)
	channelName = "draft-399.l.19481"

	data, err := json.Marshal(&broadcast)
	if err != nil {
		level.Error(r.logger).Log("message", "could not set broadcast data", "error", err, "channel_name", channelName)
		return err
	}

	level.Debug(r.logger).Log("message", "sending draft result to kubeqm", "pick", pick, "round", round, "league_key", league.LeagueKey, "channel_name", channelName)

	result, err :=
		event.SetId(fmt.Sprintf("%s-%d", channelName,  draftResult.Pick)).
			SetChannel(channelName).
		SetMetadata(fmt.Sprintf("%s-%d-%d", league.LeagueKey, draftResult.Round, draftResult.Pick)).
		SetBody(data).
		Send(ctx)

	if err != nil {
		level.Debug(r.logger).Log("message", "error sending draft result to kubeqm", "pick", pick, "round", round, "league_key", league.LeagueKey, "channel_name", channelName)
	}
	level.Debug(r.logger).Log("kubemq_result", result)
	return err
}
