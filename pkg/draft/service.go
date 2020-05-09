package draft

import (
	"context"
	"github.com/go-kit/kit/log"
	pb "github.com/thethan/fdr_proto"
)

func NewService(logger log.Logger, repository CreateRepository, ) Service {
	return Service{logger: logger, createRepository: repository}
}


type Draft struct {
	ID            string
	Year          string
	League        string
	DraftType     string
	Users         []User
	Commissioners []User
	Roster
}

type Roster struct {
	Position  string
	Starting  int32
	MaxNumber int32
}

type User struct {
	ID    string
	Name  string
	Email string
}

type CreateRepository interface {
	CreateDraft(context.Context, pb.Season) (pb.Season, error)
}

type UpdateRepository interface {
	UpdateDraft(context.Context, pb.Season) (pb.Season, error)
}

type ListUserDraftRepository interface {
	ListUserDrafts(ctx context.Context, user pb.User) ([]pb.Season, error)
}

type Service struct {
	logger                  log.Logger
	createRepository        CreateRepository
	updateRepository        UpdateRepository
	listUserDraftRepository ListUserDraftRepository
}

func (service *Service) CreateDraft(ctx context.Context, draft pb.Season) (pb.Season, error) {
	return service.createRepository.CreateDraft(ctx, draft)
}

func (service *Service) List(ctx context.Context, user pb.User) ([]pb.Season, error) {
	return service.listUserDraftRepository.ListUserDrafts(ctx, user)
}

func (service *Service) UpdateDraft(ctx context.Context, draft pb.Season) (pb.Season, error) {
	return service.updateRepository.UpdateDraft(ctx, draft)
}
