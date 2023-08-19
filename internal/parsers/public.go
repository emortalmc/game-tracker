package parsers

import (
	"game-tracker/internal/repository/model"
	pbmodel "github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"google.golang.org/protobuf/proto"
)

var DualParsers = map[proto.Message]func(data proto.Message, g *model.Game) error{
	// Common
	&pbmodel.CommonGameTeamData{}: parseGameTeamData,
}

var LiveParsers = map[proto.Message]func(data proto.Message, g *model.LiveGame) error{
	// TowerDefence
	&pbmodel.TowerDefenceStartData{}:  handleTowerDefenceStartData,
	&pbmodel.TowerDefenceUpdateData{}: handleTowerDefenceUpdateData,
}

var HistoricParsers = map[proto.Message]func(data proto.Message, g *model.HistoricGame) error{
	// Common
	&pbmodel.CommonGameFinishWinnerData{}: parseGameFinishWinnerData,

	// TowerDefence
	&pbmodel.TowerDefenceFinishData{}: handleTowerDefenceFinishData,
}
