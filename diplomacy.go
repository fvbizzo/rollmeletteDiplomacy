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
	Board      map[string]*Region `json:"map"`
	Turn       string             `json:"turn"`
	OrderStack []Orders           `json:"orderStack"`
}

// The board is built of regions wich have a name, are either occupied or not, are owned by a player, are either a base or not and are connected to other regions
type Region struct {
	Name      string            `json:"name"`
	Occupied  bool              `json:"occupied"`
	Owner     Team              `json:"owner"`
	Base      bool              `json:"base"`
	Type      string            `json:"type"`
	Frontiers map[string]string `json:"frontiers"`
}

// The Team includes the team name, the Player address and a map of all the current armies this player has
type Team struct {
	Name   string           `json:"name"`
	Player common.Address   `json:"player"`
	Armies map[string]*Army `json:"armies"`
	Bases  int              `json:"bases"`
}

type InputKind string

// Input kinds accepted
const (
	MoveArmy   InputKind = "MoveArmy"
	BuildArmy  InputKind = "BuildArmy"
	PassTurn   InputKind = "PassTurn"
	DeleteArmy InputKind = "DeleteArmy"
	Retreat    InputKind = "Retreat"
)

type Input struct {
	Kind    InputKind       `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// Give order payload is simply an Orders struct
type GiveOrderPayload = Orders

// Army struct represents an army unit
// Type is either army or boat
// Position is the name of the region it currently is
// Owner is the name of the Team that Owns this army
type Army struct {
	Type     string `json:"type"`
	Position string `json:"position"`
	Owner    string `json:"owner"`
}

// BuildArmyPayload is the payload for the building army input
// Type of the army either army or boat
// Position it is been built or deleted
// Owner of  the army
// Delete is a bool indicating if the player is deleting an army
type BuildArmyPayload struct {
	Type     string `json:"type"`
	Position string `json:"Position"`
	Owner    string `json:"owner"`
	Delete   bool   `json:"delete"`
}

type Orders struct {
	Ordertype  string `json:"orderType"`
	OrderOwner Team   `json:"orderOwner"`
	ToRegion   Region `json:"toRegion"`
	FromRegion Region `json:"fromRegion"`
}

type RetreatArmyPayload = GiveOrderPayload

type PassTurnPayload string

type GameApplication struct {
	state     GameState
	Countries map[common.Address]Team
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
		Countries: map[common.Address]Team{
			England: {
				Name:   "England",
				Player: England,
				Armies: make(map[string]*Army),
				Bases:  3,
			}},

		RoundTime: RoundTime,
		state: GameState{
			Board: map[string]*Region{
				"London": {
					Name:     "London",
					Occupied: false,
					Owner: Team{
						Name:   "England",
						Player: England,
						Bases:  3,
					},
					Base:      true,
					Type:      "coast",
					Frontiers: make(map[string]string)},
			},
			Turn:       "build",
			OrderStack: []Orders{},
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
		var inputPayload GiveOrderPayload
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
	case PassTurn:
		var inputPayload PassTurnPayload
		err = json.Unmarshal(input.Payload, &inputPayload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		err = a.passTurn(metadata)
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
	if !a.state.Board[inputPayload.Position].Base {
		return fmt.Errorf("cant build an army outside a base")
	}
	if a.state.Board[inputPayload.Position].Occupied {
		return fmt.Errorf("cant build an army in occupied region")
	}
	if metadata.MsgSender != a.state.Board[inputPayload.Position].Owner.Player {
		return fmt.Errorf("cant build an army in a territory you dont own")
	}
	if inputPayload.Type == "boat" && a.state.Board[inputPayload.Position].Type != "coast" {
		return fmt.Errorf("cant build a boat in a landlocked territory")
	}

	if inputPayload.Delete {
		a.state.Board[inputPayload.Position].Occupied = false
		delete(a.Countries[metadata.MsgSender].Armies, inputPayload.Position)
	} else {
		a.state.Board[inputPayload.Position].Occupied = true
		a.Countries[metadata.MsgSender].Armies[inputPayload.Position] = &Army{
			Type:     inputPayload.Type,
			Position: inputPayload.Position,
			Owner:    inputPayload.Owner,
		}
	}

	return nil
}

func (a *GameApplication) handleMoveArmy(
	metadata rollmelette.Metadata,
	inputPayload GiveOrderPayload,
) error {
	moveSet := map[string]bool{
		"move":         true,
		"support move": true,
		"support hold": true,
		"convoy":       true,
		"hold":         true,
	}
	if a.state.Turn != "move" {
		return fmt.Errorf("cant move an army outside move phase")
	}
	if !a.state.Board[inputPayload.FromRegion.Name].Occupied {
		return fmt.Errorf("cant order an army to move from an empty region")
	}
	if metadata.MsgSender != a.Countries[metadata.MsgSender].Player {
		return fmt.Errorf("cant order an army that dont belong to you")
	}
	if !moveSet[inputPayload.Ordertype] {
		return fmt.Errorf("invalid order")
	}

	a.state.OrderStack = append(a.state.OrderStack, inputPayload)
	return nil
}

func (a *GameApplication) passTurn(
	rollmelette.Metadata,
) error {
	if a.state.Turn == "move" {
		a.state.Turn = "build"
	} else {
		a.state.Turn = "move"
	}
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
