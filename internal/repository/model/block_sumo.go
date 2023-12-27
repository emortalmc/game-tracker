package model

import (
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/model/gametracker"
	"github.com/google/uuid"
)

type LiveBlockSumoData struct {
	Scoreboard *BlockSumoScoreboard `bson:"scoreboard"`
}

type HistoricBlockSumoData struct {
	Scoreboard *BlockSumoScoreboard `bson:"scoreboard"`
}

func CreateLiveBlockSumoDataFromUpdate(data *gametracker.BlockSumoUpdateData) (*LiveBlockSumoData, error) {
	scoreboard, err := CreateBlockSumoScoreboard(data.Scoreboard)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scoreboard: %w", err)
	}

	return &LiveBlockSumoData{
		Scoreboard: scoreboard,
	}, nil
}

func (d *LiveBlockSumoData) Update(data *gametracker.BlockSumoUpdateData) error {
	scoreboard, err := CreateBlockSumoScoreboard(data.Scoreboard)
	if err != nil {
		return fmt.Errorf("failed to parse scoreboard: %w", err)
	}

	d.Scoreboard = scoreboard

	return nil
}

func CreateBlockSumoScoreboard(data *gametracker.BlockSumoScoreboard) (*BlockSumoScoreboard, error) {
	entries := make([]*BlockSumoScoreboardEntry, len(data.Entries))

	for i, e := range data.Entries {
		entry, err := CreateBlockSumoScoreboardEntry(e)
		if err != nil {
			return nil, fmt.Errorf("failed to parse scoreboard entry: %w", err)
		}

		entries[i] = entry
	}

	return &BlockSumoScoreboard{Entries: entries}, nil
}

type BlockSumoScoreboard struct {
	Entries []*BlockSumoScoreboardEntry `bson:"entries"`
}

type BlockSumoScoreboardEntry struct {
	PlayerID uuid.UUID `bson:"playerId"`

	RemainingLives int32 `bson:"remainingLives"`
	Kills          int32 `bson:"kills"`
	FinalKills     int32 `bson:"finalKills"`
}

func CreateBlockSumoScoreboardEntry(e *gametracker.BlockSumoScoreboard_Entry) (*BlockSumoScoreboardEntry, error) {
	id, err := uuid.Parse(e.Player.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse player id: %w", err)
	}

	return &BlockSumoScoreboardEntry{
		PlayerID:       id,
		RemainingLives: e.RemainingLives,
		Kills:          e.Kills,
		FinalKills:     e.FinalKills,
	}, nil
}

func CreateHistoricBlockSumoDataFromFinish(data *gametracker.BlockSumoFinishData) (*HistoricBlockSumoData, error) {
	scoreboard, err := CreateBlockSumoScoreboard(data.Scoreboard)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scoreboard: %w", err)
	}

	return &HistoricBlockSumoData{
		Scoreboard: scoreboard,
	}, nil
}
