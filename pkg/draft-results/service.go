package draft_results

import (
	"errors"
	"math"
)

var ErrIncorrectUser = errors.New("user is not permitted")

type DraftInformationRepo interface {
	GetDraftInformation(DraftID int) (Draft, error)
}

type ListDraftResultsRepo interface {
	GetDraftResults(DraftID int) ([]DraftResult, error)
}

type Draft struct {
	Users     []User
	DraftType string // snake or traditional
	Rounds    int
	Commissioners []User
}

type User struct {
	ID    string
	Order int
}

type DraftResult struct {
	Order   int // this is the order of the draft-results
	DraftID int
	UserID  string
}

type Service struct {
	listDraftResults     ListDraftResultsRepo
	DraftInformationRepo DraftInformationRepo
}

func (s Service) SaveDraftResult() {
	// validate position
	// validate user
	// validate player
}

func (s Service) StreamDraftResults() {

}

func (s Service) calculateWhoseTurnItIs(result *DraftResult) (err error) {

	// get all the draft-results results
	results, err := s.listDraftResults.GetDraftResults(result.DraftID)
	if len(results) != result.Order {
		return err
	}
	lenOfResults := len(results)
	// get the users and draft-results order
	draft, err := s.DraftInformationRepo.GetDraftInformation(result.DraftID)
	if err != nil {
		return err
	}
	currentRoundFloatValue := math.Ceil(float64(lenOfResults / draft.Rounds))
	// get the draft-results type
	users := draft.Users
	// if the draft-results type is snake and we are in an odd number round,
	// we flip the slice of users
	if draft.DraftType == "snake" && math.Mod(currentRoundFloatValue, float64(2)) != 0 {
		users = reverseSlice(users)
	}

	// get current position in the round
	draftOrderInRound := result.Order - (int(currentRoundFloatValue - float64(1))* len(users))
	// check if the user matches
	if users[draftOrderInRound-1].ID != result.UserID || isDraftUserInListOfCommissioners(result.UserID, draft.Commissioners) {
		return ErrIncorrectUser
	}

	return nil

}

func isDraftUserInListOfCommissioners(userID string, commissioners []User) bool {
	for _, commissioner := range commissioners {
		if userID == commissioner.ID {
			return true
		}
	}
	return false
}

func reverseSlice(s []User) []User {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
