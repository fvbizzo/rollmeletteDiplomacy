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
	Board      map[string]*Region       `json:"map"`
	Units      []*Unit                  `json:"units"`
	Players    map[common.Address]*Team `json:"players"`
	Turn       string                   `json:"turn"`
	OrderStack []Orders                 `json:"orderStack"`
}

// The board is built of regions wich have a name, are either occupied or not, are owned by a player, are either a base or not and are connected to other regions
type Region struct {
	Name         string    `json:"name"`
	Occupied     bool      `json:"occupied"`
	Owner        string    `json:"owner"`
	SupplyCenter bool      `json:"supplyCenter"`
	Coastal      bool      `json:"coastal"`
	Sea          bool      `json:"sea"`
	Neighbors    []*Region `json:"frontiers"`
}

// The Team includes the team name, the Player address and a map of all the current armies this player has
type Team struct {
	Name   string           `json:"name"`
	Player common.Address   `json:"player"`
	Armies map[string]*Unit `json:"armies"`
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

// Unit struct represents an army unit
// Type is either army or boat
// Position is the name of the region it currently is
// Owner is the name of the Team that Owns this unit
type Unit struct {
	Type     string  `json:"type"`
	Position *Region `json:"position"`
	Owner    string  `json:"owner"`
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
	OrderOwner string `json:"orderOwner"`
	ToRegion   Region `json:"toRegion"`
	FromRegion Region `json:"fromRegion"`
}

type RetreatArmyPayload = GiveOrderPayload

type PassTurnPayload string

type GameApplication struct {
	state     GameState
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

		RoundTime: RoundTime,
		state: GameState{
			Board:      initializeRegions(),
			Players:    initializePlayers(Austria, England, France, Germany, Italy, Russia, Turkey),
			Turn:       "build",
			OrderStack: []Orders{},
		},

		//state.Players := initializePlayers(Austria, Engalnd, France, Germany, Italy, Russia, Turkey)
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
		err = a.passTurn()
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
	if !a.state.Board[inputPayload.Position].SupplyCenter {
		return fmt.Errorf("cant build an army outside a suply center")
	}
	if a.state.Board[inputPayload.Position].Occupied {
		return fmt.Errorf("cant build an army in occupied region")
	}
	if a.state.Players[metadata.MsgSender].Name != a.state.Board[inputPayload.Position].Owner {
		return fmt.Errorf("cant build an army in a territory you dont own")
	}
	if inputPayload.Type == "boat" && !a.state.Board[inputPayload.Position].Coastal {
		return fmt.Errorf("cant build a boat in a landlocked territory")
	}
	if a.state.Players[metadata.MsgSender].Name != inputPayload.Owner {
		return fmt.Errorf("cant build another player's army")
	}

	if inputPayload.Delete {
		a.state.Board[inputPayload.Position].Occupied = false
		delete(a.state.Players[metadata.MsgSender].Armies, inputPayload.Position)
	} else {
		a.state.Board[inputPayload.Position].Occupied = true
		a.state.Players[metadata.MsgSender].Armies[inputPayload.Position] = &Unit{
			Type:     inputPayload.Type,
			Position: a.state.Board[inputPayload.Position],
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
	if metadata.MsgSender != a.state.Players[metadata.MsgSender].Player {
		return fmt.Errorf("cant order an army that dont belong to you")
	}
	if !moveSet[inputPayload.Ordertype] {
		return fmt.Errorf("invalid order")
	}

	a.state.OrderStack = append(a.state.OrderStack, inputPayload)
	return nil
}

func (a *GameApplication) resolveConflict(
	rollmelette.Metadata,
) error {
	// Map to store the strength of each player's units in the conflict
	strength := make(map[string]int)

	// Count the strength of each player's units in the conflict
	for _, unit := range a.state.OrderStack {
		strength[unit.OrderOwner]++
	}

	// Find the player with the highest strength
	maxStrength := 0
	var winner string
	for player, s := range strength {
		if s > maxStrength {
			maxStrength = s
			winner = player
		}
	}

	// Remove units of losing players from the territory
	for _, unit := range a.state.OrderStack {
		if unit.OrderOwner != winner {
			unit.ToRegion.Name = "" // Clear the occupant of the territory
			// You might want to remove the defeated unit from the player's controlled units list
		}
	}

	return nil
}

func (a *GameApplication) passTurn() error {
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
