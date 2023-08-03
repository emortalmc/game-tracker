package kafka

import (
	"game-tracker/internal/repository/model"
	pbmodel "github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"google.golang.org/protobuf/proto"
	"log"
)

func handleTowerDefenceStartData(m proto.Message, g *model.LiveGame) error {
	cast := m.(*pbmodel.TowerDefenceStartData)
	g.SetGameData(model.LiveTowerDefenceDataFromStart(cast))

	return nil
}

func handleTowerDefenceUpdateData(m proto.Message, g *model.LiveGame) error {
	cast := m.(*pbmodel.TowerDefenceUpdateData)

	(g.GameData).(*model.LiveTowerDefenceData).Update(cast)

	log.Printf("Changed: %+v", g.GameData)

	return nil
}
