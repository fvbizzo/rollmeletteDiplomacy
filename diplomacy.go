package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gligneul/rollmelette"
)

// State of the game Board and turn type
type GameState struct {
	Board map[string]Region `json:"map"`
	Turn  string            `json:"turn"`
}

// The board is built of regions wich have a name, are either occupied or not, are owned by a player, are either a base or not and are connected to other regions
type Region struct {
	Name      string            `json:"name"`
	Ocuppied  bool              `json:"occupied"`
	Owner     Team              `json:"owner"`
	Base      bool              `json:"base"`
	Type      string            `json:"type"`
	Frontiers map[string]Region `json:"frontiers"`
}

type Team struct {
	Name   string         `json:"name"`
	Player common.Address `json:"player"`
}

type InputKind string

const (
	MoveArmy  InputKind = "MoveArmy"
	BuildArmy InputKind = "BuildArmy"
)

type Input struct {
	Kind    InputKind       `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

type MoveArmyPayload struct {
	CurrentRegion Region `json:"currentRegion"`
	Destiny       Region `json:"destiny"`
}

type Army struct {
	Type     string `jsno:"type"`
	Position Region `json:"position"`
	Owner    Team   `json:"owner"`
}

type BuildArmyPayload struct {
	Base     Region `json:"base"`
	ArmyType string `json:"armyType"`
}

type GameApplication struct {
	state     GameState
	Austria   common.Address
	England   common.Address
	France    common.Address
	Germany   common.Address
	Italy     common.Address
	Russia    common.Address
	Turkey    common.Address
	RoundTime int
}

func NewGameApplication(Austria common.Address,
	England common.Address,
	France common.Address,
	Germany common.Address,
	Italy common.Address,
	Russia common.Address,
	Turkey common.Address,
	RoundTime int,
) *GameApplication {
	return &GameApplication{
		Austria:   Austria,
		England:   England,
		France:    France,
		Germany:   Germany,
		Italy:     Italy,
		Russia:    Russia,
		Turkey:    Turkey,
		RoundTime: RoundTime,
		state: GameState{
			Turn: "move",
		},
	}
}

func (a *GameApplication) Advance(
	env rollmelette.Env,
	metadata rollmelette.Metadata,
	deposit rollmelette.Deposit,
	payload []byte,
) error {
	var input Input
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}

	switch input.Kind {
	case MoveArmy:
		var inputPayload MoveArmyPayload
		err = json.Unmarshal(input.Payload, &inputPayload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		err = a.handleMoveArmy(metadata, inputPayload)
		if err != nil {
			return err
		}
	case BuildArmy:
		var inputPayload BuildArmyPayload
		err = json.Unmarshal(input.Payload, &inputPayload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		err = a.handleBuildArmy(metadata, inputPayload)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid input kind: %v", input.Kind)
	}
	return a.Inspect(env, nil)
}

func (a *GameApplication) Inspect(env rollmelette.EnvInspector, payload []byte) error {
	bytes, err := json.Marshal(a.state)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	env.Report(bytes)
	return nil
}

func (a *GameApplication) handleBuildArmy(
	metadata rollmelette.Metadata,
	inputPayload BuildArmyPayload,
) error {
	if a.state.Turn != "build" {
		return fmt.Errorf("cant build an army outside build phase")
	}
	if !inputPayload.Base.Base {
		return fmt.Errorf("cant build an army outside a base")
	}
	if inputPayload.Base.Ocuppied {
		return fmt.Errorf("cant build an army in occupied region")
	}
	if metadata.MsgSender != a.state.Board[inputPayload.Base.Name].Owner.Player {
		return fmt.Errorf("cant build an army in a territory you dont own")
	}
	if inputPayload.ArmyType == "boat" && inputPayload.Base.Type != "coast" {
		return fmt.Errorf("cant build a boat in a landlocked territory")
	}

	return nil
}

func (a *GameApplication) handleMoveArmy(
	metadata rollmelette.Metadata,
	inputPayload MoveArmyPayload,
) error {

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
