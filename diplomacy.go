package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gligneul/rollmelette"
)

type GameState struct {
	Board map[string]Region `json:"map"`
}

type Region struct {
	Name     string `json:"name"`
	Ocuppied bool   `json:"occupied"`
	Color    string `json:"color"`
}

type InputKind string

const (
	MoveArmy   InputKind = "MoveArmy"
	CreateArmy InputKind = "CreateArmy"
)

type Input struct {
	Kind    InputKind       `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

type MoveArmyPayload struct {
	CurrentRegion Region `json:"currentRegion"`
	Destiny       Region `json:"destiny"`
}

type GameApplication struct {
	state   GameState
	Austria common.Address
	England common.Address
	France  common.Address
	Germany common.Address
	Italy   common.Address
	Russia  common.Address
	Turkey  common.Address
}

func (a *GameApplication) Advance(
	env rollmelette.Env,
	metadata rollmelette.Metadata,
	deposit rollmelette.Deposit,
	payload []byte,
) error {
	// Handle advance input
	return nil
}

func (a *GameApplication) Inspect(env rollmelette.EnvInspector, payload []byte) error {
	// Handle inspect input
	return nil
}

func main() {
	ctx := context.Background()
	opts := rollmelette.NewRunOpts()
	app := new(GameApplication)
	err := rollmelette.Run(ctx, opts, app)
	if err != nil {
		slog.Error("application error", "error", err)
	}
}
