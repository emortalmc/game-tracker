package model

import (
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/model/gametracker"
)

type LiveBlockSumoData struct {
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
	BasicPlayer *BasicPlayer `bson:"basicPlayer"`

	RemainingLives int32 `bson:"remainingLives"`
	Kills          int32 `bson:"kills"`
	FinalKills     int32 `bson:"finalKills"`
}

func CreateBlockSumoScoreboardEntry(e *gametracker.BlockSumoScoreboard_Entry) (*BlockSumoScoreboardEntry, error) {
	basicPlayer, err := BasicPlayerFromProto(e.Player)
	if err != nil {
		return nil, err
	}

	return &BlockSumoScoreboardEntry{
		BasicPlayer:    basicPlayer,
		RemainingLives: e.RemainingLives,
		Kills:          e.Kills,
		FinalKills:     e.FinalKills,
	}, nil
}
