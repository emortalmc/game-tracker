package kafka

import (
	"game-tracker/internal/repository/model"
	pbmodel "github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"google.golang.org/protobuf/proto"
)

func parseGameStartTeamData(m proto.Message, g *model.LiveGame) error {
	cast := m.(*pbmodel.CommonGameStartTeamData)

	teams := make([]*model.Team, len(cast.Teams))
	for i, t := range cast.Teams {
		parsed, err := model.TeamFromProto(t)
		if err != nil {
			return err
		}

		teams[i] = parsed
	}

	g.TeamData = &teams

	return nil
}
