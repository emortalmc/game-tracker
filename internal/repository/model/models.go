package model

import (
	"fmt"
	"game-tracker/internal/utils"
	"github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type GameStage uint8

const (
	StagePreGame    GameStage = 0
	StageInProgress GameStage = 1

	LiveTowerDefenceDataId int32 = 1
)

var dataType = map[int32]interface{}{
	LiveTowerDefenceDataId: &LiveTowerDefenceData{},
}

func getDataType(example interface{}) int32 {
	for k, v := range dataType {
		if utils.IsSameType(example, v) {
			return k
		}
	}

	return 0 // 0 is ignored by omitempty so it won't be put in the db
}

type LiveGame struct {
	// GameData provided at game create (allocation messages)

	Id         primitive.ObjectID `bson:"_id"`
	GameModeId string             `bson:"gameModeId"`
	Stage      GameStage          `bson:"stage"`

	LastUpdated time.Time      `bson:"lastUpdated"`
	ServerId    string         `bson:"serverId"`
	Players     []*BasicPlayer `bson:"players"`

	// GameData provided at game start (message)

	StartTime time.Time `bson:"startTime"`

	// The below data is all optional and varies by game mode

	// GameData is data specific to the game mode. It is only present if the game sends it
	GameData     interface{} `bson:"gameData,omitempty"`
	GameDataType int32       `bson:"gameDataType,omitempty"`

	// TeamData only present if the game sends it
	TeamData *[]*Team `bson:"teams,omitempty"`
}

func (g *LiveGame) SetGameData(data interface{}) {
	g.GameData = data
	g.GameDataType = getDataType(data)
}

// ParseGameData converts and replaces the game data from bson.D to the correct type
func (g *LiveGame) ParseGameData() error {
	if g.GameData == nil {
		return nil
	}

	if g.GameDataType == 0 {
		return fmt.Errorf("game data type not set but game data is present")
	}

	example, ok := dataType[g.GameDataType]
	if !ok {
		return fmt.Errorf("unknown game data type: %d", g.GameDataType)
	}

	if example == nil {
		return fmt.Errorf("unknown game data type: %d", g.GameDataType)
	}

	bytes, err := bson.Marshal(g.GameData)
	if err != nil {
		return fmt.Errorf("failed to marshal game data: %w", err)
	}

	if err := bson.Unmarshal(bytes, example); err != nil {
		return fmt.Errorf("failed to unmarshal game data: %w", err)
	}

	g.GameData = example

	return nil
}

type HistoricGame[T interface{}] struct {
	Id primitive.ObjectID `bson:"_id"`

	StartTime time.Time `bson:"startTime"`
	EndTime   time.Time `bson:"endTime"`

	// The below data is all optional and varies by game mode
	GameData   *T                  `bson:"gameData"`
	WinnerData *HistoricWinnerData `bson:"winnerData"`
}

type HistoricWinnerData struct {
	Winners []*BasicPlayer `bson:"winners"`
	Losers  []*BasicPlayer `bson:"losers"`
}

type BasicPlayer struct {
	Id       uuid.UUID `bson:"id"`
	Username string    `bson:"username"`
}

type Team struct {
	Id           string      `bson:"id"`
	FriendlyName string      `bson:"friendlyName"`
	Color        int32       `bson:"color"`
	PlayerIds    []uuid.UUID `bson:"playerIds"`
}

func TeamFromProto(t *gametracker.Team) (*Team, error) {
	playerIds := make([]uuid.UUID, len(t.PlayerIds))
	for i, id := range t.PlayerIds {
		playerId, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}

		playerIds[i] = playerId
	}

	return &Team{
		Id:           t.Id,
		FriendlyName: t.FriendlyName,
		Color:        t.Color,
		PlayerIds:    playerIds,
	}, nil
}

func BasicPlayerFromProto(p *gametracker.BasicGamePlayer) (*BasicPlayer, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, err
	}

	return &BasicPlayer{
		Id:       id,
		Username: p.Username,
	}, nil
}

func BasicPlayersFromProto(players []*gametracker.BasicGamePlayer) ([]*BasicPlayer, error) {
	basicPlayers := make([]*BasicPlayer, len(players))
	for i, p := range players {
		basicPlayer, err := BasicPlayerFromProto(p)
		if err != nil {
			return nil, err
		}

		basicPlayers[i] = basicPlayer
	}

	return basicPlayers, nil
}
