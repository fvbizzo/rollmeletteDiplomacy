package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rollmelette/rollmelette"
)

// State of the game Board and turn type
type GameState struct {
	Board       map[string]*Region       `json:"map"`
	Units       map[int]*Unit            `json:"units"`
	Players     map[common.Address]*Team `json:"players"`
	Turn        string                   `json:"turn"`
	MoveCounter bool                     `json:"MoveCounter"`
}

// The board is built of regions wich have a name, are either occupied or not, are owned by a player, are either a base or not and are connected to other regions
type Region struct {
	Name         string              `json:"name"`
	Occupied     bool                `json:"occupied"`
	Owner        string              `json:"owner"`
	SupplyCenter bool                `json:"supplyCenter"`
	Coastal      bool                `json:"coastal"`
	Sea          bool                `json:"sea"`
	Neighbors    []*string           `json:"frontiers"`
	SubRegions   map[string][]string `json:"subRegions"`
}

type SubRegion struct {
	Name      string   `json:"name"`
	Neighbors []string `json:"neighbors"`
}

// Movement struct to handle move turns
type Movement struct {
	Type     string
	From     string
	To       string
	Position string
}

type SupportOrder struct {
	SupportingUnit *Unit
	SupportedUnit  *Unit
	TargetRegion   string
}

// The Team includes the team name, the Player address and a map of all the current armies this player has
type Team struct {
	Name   string            `json:"name"`
	Player common.Address    `json:"player"`
	Armies map[int]string    `json:"armies"`
	Bases  int               `json:"bases"`
	Ready  bool              `json:"ready"`
	Builds []*BuildArmyInput `json:"builds"`
}

type BuildArmyInput struct {
	Info   BuildArmyPayload `json:"info"`
	Player common.Address   `json:"player"`
}

type InputKind string

// Input kinds accepted
const (
	MoveArmy    InputKind = "MoveArmy"
	BuildArmy   InputKind = "BuildArmy"
	ReadyOrders InputKind = "ReadyOrders"
	DeleteArmy  InputKind = "DeleteArmy"
	Retreat     InputKind = "Retreat"
)

type Input struct {
	Kind    InputKind       `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// Give order payload is simply an Orders struct
type GiveOrderPayload = Orders

// Unit struct represents an army unit
// Type is either army or navy
// Position is the name of the region it currently is
// Owner is the name of the Team that Owns this unit
type Unit struct {
	ID           int            `json:"ID"`
	Type         string         `json:"type"`
	Position     string         `json:"position"`
	SubPosition  string         `json:"subPosition"`
	Owner        common.Address `json:"owner"`
	CurrentOrder Orders         `json:"currentOrder"`
	Retreating   string         `json:"retreating"`
}

var UnitID int = 0

// BuildArmyPayload is the payload for the building army input
// Type of the army either army or navy
// Position it is been built or deleted
// Owner of  the army
// Delete is a bool indicating if the player is deleting an army
type BuildArmyPayload struct {
	Type        string `json:"type"`
	Position    string `json:"Position"`
	SubPosition string `json:"subPosition"`
	Owner       string `json:"owner"`
	Delete      int    `json:"delete"`
}

type Orders struct {
	UnitID        int    `json:"unitID"`
	Ordertype     string `json:"orderType"`
	OrderOwner    string `json:"orderOwner"`
	ToRegion      string `json:"toRegion"`
	ToSubRegion   string `json:"toSubRegion"`
	FromRegion    string `json:"fromRegion"`
	FromSubRegion string `json:"fromSubRegion"`
}

type MoveOrder struct {
	Unit       *Unit
	FromRegion string
	ToRegion   string
	SubRegion  string
}

type RetreatOrderPayload struct {
	UnitID      int    `json:"unitID"`
	Delete      bool   `json:"delete"`
	ToRegion    string `json:"toRegion"`
	ToSubRegion string `json:"toSubRegion"`
}

type PassTurnPayload string

type GameApplication struct {
	state     GameState
	RoundTime int
}

// ConflictOutcome represents the outcome of a conflict between two units' orders
type ConflictOutcome struct {
	Winner  *Unit // The winning unit, if any
	Loser   *Unit // The losing unit, if any
	Bounced bool  // Indicates if the conflict resulted in units bouncing back to their original positions
}

var SubRegionsList = [3]string{"Bulgaria", "St Petersburg", "Spain"}

func NewGameApplication(Austria common.Address,
	England common.Address,
	France common.Address,
	Germany common.Address,
	Italy common.Address,
	Russia common.Address,
	Turkey common.Address,
	RoundTime int,
) *GameApplication {
	Game := GameApplication{

		RoundTime: RoundTime,
		state: GameState{
			Board:       initializeRegions(),
			Players:     initializePlayers(Austria, England, France, Germany, Italy, Russia, Turkey),
			Units:       initializeUnits(Austria, England, France, Germany, Italy, Russia, Turkey),
			Turn:        "move",
			MoveCounter: false,
		},
	}
	UnitID = len(Game.state.Units) + 1
	return &Game

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
	case ReadyOrders:
		var inputPayload PassTurnPayload
		err = json.Unmarshal(input.Payload, &inputPayload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		err = a.ReadyOrders(metadata)
		if err != nil {
			return err
		}
	case Retreat:
		var inputPayload RetreatOrderPayload
		err = json.Unmarshal(input.Payload, &inputPayload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}
		err = a.handleRetreat(metadata, inputPayload)
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

func isConnected(From *Region, To *string) bool {
	for _, city := range From.Neighbors {
		c := city
		if *c == *To {
			return true
		}
	}
	return false
}

func isSubRegionConnected(from []string, to string) bool {
	for _, neighbor := range from {
		if neighbor == to {
			return true
		}
	}
	return false
}

func (a *GameApplication) ReadyOrders(
	metadata rollmelette.Metadata,
) error {
	if a.state.Players[metadata.MsgSender] == nil {
		return fmt.Errorf("msg sender is not a player")
	}

	a.state.Players[metadata.MsgSender].Ready = true
	for _, player := range a.state.Players {
		if !player.Ready {
			return nil
		}
	}
	err := a.passTurn()
	if err != nil {
		return fmt.Errorf("pass turn function not working")
	}
	return nil
}

func (a *GameApplication) processMoves() {
	moveOrders := a.prepareMoves()
	a.executeMoves(moveOrders)
	ResolveMovementConflicts(&a.state)
}

func (a *GameApplication) passTurn() error {
	for _, player := range a.state.Players {
		player.Ready = false
	}

	if a.state.Turn == "move" {
		// Register all departures
		a.processMoves()
		ResetOrders(a)
		if a.state.MoveCounter {
			a.state.Turn = "build"
			a.state.MoveCounter = false
		} else {
			a.state.MoveCounter = true
		}
		for _, unit := range a.state.Units {
			if unit.Retreating != "" {
				a.state.Turn = "retreats"
				setForDelete(a)
			}
		}
	} else if a.state.Turn == "build" {
		BuildUnits(a)
		a.state.Turn = "move"
	} else if a.state.Turn == "retreats" {
		resolveRetreats(a)
		ResetOrders(a)
		if a.state.MoveCounter {
			a.state.Turn = "move"
		} else {
			a.state.Turn = "build"
		}
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
