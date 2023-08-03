package main

import (
	"context"
	"fmt"
	"game-tracker/internal/config"
	"github.com/emortalmc/proto-specs/gen/go/message/gametracker"
	pbmodel "github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"
)

var (
	serverId = "tower-defence-test-3wd6ws34-wc3463"
	gameId   = primitive.NewObjectID()

	expectationalId = uuid.MustParse("8d36737e-1c0a-4a71-87de-9906f577845e")
	emortaldevId    = uuid.MustParse("7bd5b459-1e6b-4753-8274-1fbd2fe9a4d5")
)

type app struct {
	w *kafka.Writer
}

func main() {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		panic(err)
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(fmt.Sprintf("%s:%d", cfg.Kafka.Host, cfg.Kafka.Port)),
		Topic:    "games",
		Balancer: &kafka.LeastBytes{},
		Async:    false,
	}

	a := &app{
		w: w,
	}

	a.writeStartMessage()
	log.Printf("wrote start message, waiting 5 seconds before sending update")
	time.Sleep(5 * time.Second)
	a.writeUpdateMessage()
	log.Printf("wrote update message, waiting 5 seconds before sending end")

	if err != nil {
		panic(err)
	}
}

func (a *app) writeStartMessage() {
	tdStartData, err := anypb.New(&pbmodel.TowerDefenceStartData{
		HealthData: &pbmodel.TowerDefenceHealthData{
			MaxHealth:  1000,
			BlueHealth: 1000,
			RedHealth:  1000,
		},
	})
	if err != nil {
		panic(err)
	}

	teamStartData, err := anypb.New(&pbmodel.CommonGameStartTeamData{Teams: []*pbmodel.Team{
		{
			Id:           "blue",
			FriendlyName: "Blue",
			Color:        0x0000FF,
			PlayerIds:    []string{expectationalId.String()},
		},
		{
			Id:           "red",
			FriendlyName: "Red",
			Color:        0xFF0000,
			PlayerIds:    []string{emortaldevId.String()},
		},
	}})
	if err != nil {
		panic(err)
	}

	message := &gametracker.GameStartMessage{
		CommonData: &gametracker.CommonGameData{
			GameModeId: "tower-defence",
			GameId:     gameId.Hex(),
			ServerId:   serverId,
			Players: []*pbmodel.BasicGamePlayer{
				{
					Id:       expectationalId.String(),
					Username: "Expectational",
				},
				{
					Id:       emortaldevId.String(),
					Username: "emortaldev",
				},
			},
		},
		StartTime: timestamppb.Now(),
		MapId:     "test-map",
		Content:   []*anypb.Any{tdStartData, teamStartData},
	}

	a.writeMessages(message)
}

func (a *app) writeUpdateMessage() {
	tdUpdateData, err := anypb.New(&pbmodel.TowerDefenceUpdateData{
		HealthData: &pbmodel.TowerDefenceHealthData{
			MaxHealth:  1000,
			BlueHealth: 750,
			RedHealth:  500,
		},
	})
	if err != nil {
		panic(err)
	}

	message := &gametracker.GameUpdateMessage{
		CommonData: &gametracker.CommonGameData{
			GameModeId: "tower-defence",
			GameId:     gameId.Hex(),
			ServerId:   serverId,
			Players: []*pbmodel.BasicGamePlayer{
				{
					Id:       expectationalId.String(),
					Username: "Expectational",
				},
				{
					Id:       emortaldevId.String(),
					Username: "emortaldev",
				},
			},
		},
		Content: []*anypb.Any{tdUpdateData},
	}

	a.writeMessages(message)
}

func (a *app) writeMessages(messages ...proto.Message) {
	kMessages := make([]kafka.Message, len(messages))

	for i, m := range messages {
		bytes, err := proto.Marshal(m)
		if err != nil {
			panic(err)
		}

		kMessages[i] = kafka.Message{
			Headers: []kafka.Header{
				{
					Key:   "X-Proto-Type",
					Value: []byte(m.ProtoReflect().Descriptor().FullName()),
				},
			},
			Value: bytes,
		}
	}

	if err := a.w.WriteMessages(context.Background(), kMessages...); err != nil {
		panic(err)
	}
}
