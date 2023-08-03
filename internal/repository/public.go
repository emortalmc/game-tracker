package repository

import (
	"context"
	"game-tracker/internal/repository/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	GetLiveGame(ctx context.Context, id primitive.ObjectID) (*model.LiveGame, error)
	// SaveLiveGame saves a game (with upsert)
	SaveLiveGame(ctx context.Context, game *model.LiveGame) error
}
