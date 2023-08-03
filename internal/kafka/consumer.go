package kafka

import (
	"context"
	"fmt"
	"game-tracker/internal/config"
	"game-tracker/internal/repository"
	"game-tracker/internal/repository/model"
	"github.com/emortalmc/proto-specs/gen/go/message/gametracker"
	pbmodel "github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"github.com/emortalmc/proto-specs/gen/go/nongenerated/kafkautils"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"strings"
	"sync"
	"time"
)

const gamesTopic = "games"

var parsers = map[proto.Message]func(data proto.Message, g *model.LiveGame) error{
	&pbmodel.TowerDefenceStartData{}:   handleTowerDefenceStartData,
	&pbmodel.TowerDefenceUpdateData{}:  handleTowerDefenceUpdateData,
	&pbmodel.CommonGameStartTeamData{}: parseGameStartTeamData,
}

type consumer struct {
	logger *zap.SugaredLogger
	repo   repository.Repository

	reader *kafka.Reader
}

func NewConsumer(ctx context.Context, wg *sync.WaitGroup, cfg *config.KafkaConfig, logger *zap.SugaredLogger,
	repo repository.Repository) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{cfg.Host},
		GroupID:     "game-tracker",
		GroupTopics: []string{gamesTopic},

		Logger: kafka.LoggerFunc(func(format string, args ...interface{}) {
			logger.Infow(fmt.Sprintf(format, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(format string, args ...interface{}) {
			logger.Errorw(fmt.Sprintf(format, args...))
		}),
	})

	c := &consumer{
		logger: logger,
		repo:   repo,

		reader: reader,
	}

	handler := kafkautils.NewConsumerHandler(logger, reader)
	handler.RegisterHandler(&gametracker.GameStartMessage{}, c.handleGameStartMessage)
	handler.RegisterHandler(&gametracker.GameUpdateMessage{}, c.handleGameUpdateMessage)

	logger.Infow("started listening for kafka messages", "topics", reader.Config().GroupTopics)

	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.Run(ctx) // Run is blocking until the context is cancelled
		if err := reader.Close(); err != nil {
			logger.Errorw("failed to close kafka reader", err)
		}
	}()
}

func (c *consumer) handleGameStartMessage(ctx context.Context, _ *kafka.Message, uncastMsg proto.Message) {
	m := uncastMsg.(*gametracker.GameStartMessage)
	commonData := m.CommonData

	id, err := primitive.ObjectIDFromHex(commonData.GameId)
	if err != nil {
		c.logger.Errorw("failed to parse game id", "gameId", commonData.GameId)
		return
	}

	players, err := model.BasicPlayersFromProto(commonData.Players)
	if err != nil {
		c.logger.Errorw("failed to parse players", "players", commonData.Players)
		return
	}

	liveGame := &model.LiveGame{
		Id:          id,
		GameModeId:  commonData.GameModeId,
		Stage:       model.StageInProgress, // todo listen for matchmaker creating it
		ServerId:    commonData.ServerId,
		LastUpdated: time.Now(),
		Players:     players,
		StartTime:   m.StartTime.AsTime(),
	}

	if err := c.handleParsers(m.Content, liveGame); err != nil {
		c.logger.Errorw("failed to handle game content", "game", commonData.GameId, "content", m.Content)
		return
	}

	if err := c.repo.SaveLiveGame(ctx, liveGame); err != nil {
		c.logger.Errorw("failed to save live game", "game", liveGame, "error", err)
		return
	}
}

func (c *consumer) handleGameUpdateMessage(ctx context.Context, _ *kafka.Message, uncastMsg proto.Message) {
	m := uncastMsg.(*gametracker.GameUpdateMessage)
	commonData := m.CommonData

	id, err := primitive.ObjectIDFromHex(commonData.GameId)
	if err != nil {
		c.logger.Errorw("failed to parse game id", "gameId", commonData.GameId)
		return
	}

	liveGame, err := c.repo.GetLiveGame(ctx, id)
	if err != nil {
		c.logger.Errorw("failed to get live game", "game", id, "error", err)
		return
	}

	// common data start

	players, err := model.BasicPlayersFromProto(commonData.Players)
	if err != nil {
		c.logger.Errorw("failed to parse players", "players", commonData.Players)
		return
	}

	liveGame.Players = players
	liveGame.LastUpdated = time.Now()

	// common data end

	if err := c.handleParsers(m.Content, liveGame); err != nil {
		c.logger.Errorw("failed to handle game content", "game", commonData.GameId, "content", m.Content)
		return
	}

	if err := c.repo.SaveLiveGame(ctx, liveGame); err != nil {
		c.logger.Errorw("failed to save live game", "game", liveGame, "error", err)
	}
}

func (c *consumer) handleParsers(content []*anypb.Any, g *model.LiveGame) error {
	unhandledIndexes := make([]bool, len(content)) // every index is false by default

	for i, anyPb := range content {
		anyFullName := strings.SplitAfter(anyPb.TypeUrl, "type.googleapis.com/")[1]

		for key, parser := range parsers {
			if anyFullName == string(key.ProtoReflect().Descriptor().FullName()) {
				unmarshaled := key
				if err := anyPb.UnmarshalTo(unmarshaled); err != nil {
					return fmt.Errorf("failed to unmarshal game content: %w", err)
				}

				if err := parser(unmarshaled, g); err != nil {
					return fmt.Errorf("failed to parse game content: %w", err)
				}

				unhandledIndexes[i] = true
				break
			}
		}
	}

	// check every index has been processed and warn if not
	for i, b := range unhandledIndexes {
		if !b {
			c.logger.Warnw("unhandled game content index", "index", i, "game", g.Id, "gameMode", g.GameModeId,
				"content", content)
		}
	}

	return nil
}
