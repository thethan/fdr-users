package importer

import (
	"context"
	"github.com/go-kit/kit/log"
	pkgEntities "github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/players/entities"
)

type repo interface {
	//ReceiveMessages(ctx context.Context, messageChannel chan<- entities.ImportPlayerStat) error
	SendPlayerStatMessage(ctx context.Context, stats entities.ImportPlayerStat) error
}

type savePlayerStats interface {
	SavePlayerStats(ctx context.Context, playerKey string, season, week string, stats []pkgEntities.PlayerStat) error
}

type ImportSendService struct {
	logger log.Logger
	repo   repo
}

func NewImportSendService(logger log.Logger, repo repo) ImportSendService {
	return ImportSendService{logger: logger, repo: repo}
}

func (s *ImportSendService) SavePlayerStats(ctx context.Context, playerKey string, season, week string, stats []pkgEntities.PlayerStat) error {
	importPlayerStats := entities.ImportPlayerStat{
		PlayerKey:   playerKey,
		Week:        week,
		Season:      season,
		PlayerStats: stats,
	}

	return s.repo.SendPlayerStatMessage(ctx, importPlayerStats)
}
