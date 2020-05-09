package draft

import (
	pb "github.com/thethan/fdr_proto"

	"strings"
)

func TransformDraftToSeasonPB(draft Draft) pb.Season {

	return pb.Season{
		ID:            draft.ID,
		Year:          draft.Year,
		League:        pb.League_League_NFL,
		DraftType:     transformDraftTypeToPBSeason(draft),
		Users:         transformUsersToPBUsers(draft.Users),
		Commissioners: transformUsersToPBUsers(draft.Commissioners),
	}

}

func transformDraftTypeToPBSeason(draft Draft) pb.DraftType {
	if strings.Contains(pb.DraftType_DraftType_Snake.String(), draft.DraftType) {
		return pb.DraftType_DraftType_Snake
	}

	if strings.Contains(pb.DraftType_DraftType_Traditional.String(), draft.DraftType) {
		return pb.DraftType_DraftType_Traditional
	}

	return pb.DraftType_DraftType_Snake

}

func transformUsersToPBUsers(users []User) []*pb.User {
	pbUsers := make([]*pb.User, len(users))
	for idx := range users {
		user := transformUserToPBUser(users[idx])
		pbUsers[idx] = &user
	}
	return pbUsers
}

func transformUserToPBUser(user User) pb.User {
	// string to

	return pb.User{
		Id:    user.ID,
		Name:  user.Name,
		Image: "",
		Email: user.Email,
	}
}
