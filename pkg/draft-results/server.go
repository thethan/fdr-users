package draft_results

import (
	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/thethan/fdr_proto"
	"strconv"
	"time"
)

type Server struct {
	logger log.Logger
}

func NewServer(logger log.Logger) Server {
	return Server{logger: logger}
}

func (s *Server) Draft(req *pb.DraftRequest, stream pb.Draft_DraftServer) error {
	season := req.GetSeason()
	// Going to make a fake season
	season.ID = "1313"
	season.DraftType = pb.DraftType_DraftType_Snake
	season.League = pb.League_League_NFL

	getLastDraftResult := 200
	c := make(chan *pb.DraftPlayerResponse, getLastDraftResult)
	// get previous draft-results results
	// get from kubemq
	go s.getAllDraftResults(season, c)

	for result := range c {
		_ = stream.Send(result)
	}

	return nil
}

func (s *Server) getAllDraftResults(season *pb.Season, results chan *pb.DraftPlayerResponse) {
	// loop through 200 times
	originalTime := time.Now()
	for i := 0; i < cap(results); i++ {

		player := pb.Player{
			Id:                   int32(i),
			Name:                 "SomeName"+strconv.Itoa(i),
			Image:                "",
			Positions:            nil,
			Relations:            nil,
			SeasonStats:          nil,
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		}


		timeNow := originalTime.Add(time.Minute * 1)
		result := pb.DraftPlayerResponse{
			DraftResultID:        int32(i),
			Season:               season,
			Player:               &player,
			User:                 nil,
			Order:                int32(i +1),
			DraftTime:            &timestamp.Timestamp{Seconds: timeNow.Unix()},
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     nil,
			XXX_sizecache:        0,
		}

		results <- &result
	}

	close(results)
}