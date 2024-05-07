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
	Board       map[string]*Region       `json:"map"`
	Units       map[int]*Unit            `json:"units"`
	Players     map[common.Address]*Team `json:"players"`
	Turn        string                   `json:"turn"`
	MoveCounter bool                     `json:"MoveCounter"`
}

// The board is built of regions wich have a name, are either occupied or not, are owned by a player, are either a base or not and are connected to other regions
type Region struct {
	Name         string    `json:"name"`
	Occupied     bool      `json:"occupied"`
	Owner        string    `json:"owner"`
	SupplyCenter bool      `json:"supplyCenter"`
	Coastal      bool      `json:"coastal"`
	Sea          bool      `json:"sea"`
	Neighbors    []*string `json:"frontiers"`
}

// Movement struct to handle move turns
type Movement struct {
	Type     string
	From     string
	To       string
	Position string
}

// The Team includes the team name, the Player address and a map of all the current armies this player has
type Team struct {
	Name   string         `json:"name"`
	Player common.Address `json:"player"`
	Armies map[int]string `json:"armies"`
	Bases  int            `json:"bases"`
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
// Type is either army or navy
// Position is the name of the region it currently is
// Owner is the name of the Team that Owns this unit
type Unit struct {
	ID           int    `json:"ID"`
	Type         string `json:"type"`
	Position     string `json:"position"`
	Owner        string `json:"owner"`
	CurrentOrder Orders `json:"currentOrder"`
	Retreating   bool   `json:"retreating"`
}

var UnitID int = 0

// BuildArmyPayload is the payload for the building army input
// Type of the army either army or navy
// Position it is been built or deleted
// Owner of  the army
// Delete is a bool indicating if the player is deleting an army
type BuildArmyPayload struct {
	Type     string `json:"type"`
	Position string `json:"Position"`
	Owner    string `json:"owner"`
	Delete   int    `json:"delete"`
}

type Orders struct {
	UnitID     int    `json:"unitID"`
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

// ConflictOutcome represents the outcome of a conflict between two units' orders
type ConflictOutcome struct {
	Winner  *Unit // The winning unit, if any
	Loser   *Unit // The losing unit, if any
	Bounced bool  // Indicates if the conflict resulted in units bouncing back to their original positions
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
	Game := GameApplication{

		RoundTime: RoundTime,
		state: GameState{
			Board:       initializeRegions(),
			Players:     initializePlayers(Austria, England, France, Germany, Italy, Russia, Turkey),
			Units:       initializeUnits(),
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
		bytes, err := json.Marshal(a.state.Units)
		if err != nil {
			return fmt.Errorf("failed to marshal: %w", err)
		}

		a.Inspect(env, bytes)
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
		bytes, err := json.Marshal(a.state.Units)
		if err != nil {
			return fmt.Errorf("failed to marshal: %w", err)
		}

		a.Inspect(env, bytes)

	default:
		return fmt.Errorf("invalid input kind: %v", input.Kind)
	}
	return a.Inspect(env, nil)
}

func (a *GameApplication) Inspect(env rollmelette.EnvInspector, payload []byte) error {

	env.Report(payload)
	return nil
}

func (a *GameApplication) InspectPosition(u int) string {
	return a.state.Units[u].Position
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
	if a.state.Board[inputPayload.Position].Occupied && inputPayload.Delete == 0 {
		return fmt.Errorf("cant build an army in occupied region")
	}
	if !a.state.Board[inputPayload.Position].Occupied && inputPayload.Delete != 0 {
		return fmt.Errorf("cant delete an army in empty region")
	}
	if a.state.Players[metadata.MsgSender].Name != a.state.Board[inputPayload.Position].Owner {
		return fmt.Errorf("cant build an army in a territory you dont own")
	}
	if inputPayload.Type == "navy" && !a.state.Board[inputPayload.Position].Coastal {
		return fmt.Errorf("cant build a navy in a landlocked territory")
	}
	if a.state.Players[metadata.MsgSender].Name != inputPayload.Owner {
		return fmt.Errorf("cant build another player's army")
	}

	if inputPayload.Delete != 0 {
		a.state.Board[inputPayload.Position].Occupied = false
		delete(a.state.Players[metadata.MsgSender].Armies, inputPayload.Delete)
	} else {
		a.state.Board[inputPayload.Position].Occupied = true
		a.state.Players[metadata.MsgSender].Armies[UnitID] = inputPayload.Position
		a.state.Units[UnitID] = &Unit{
			ID:       UnitID,
			Type:     inputPayload.Type,
			Position: inputPayload.Position,
			Owner:    inputPayload.Owner,
			CurrentOrder: Orders{
				UnitID:     UnitID,
				Ordertype:  "hold",
				OrderOwner: inputPayload.Owner,
			},
		}
		UnitID += 1
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
	if a.state.Players[metadata.MsgSender].Name != a.state.Units[inputPayload.UnitID].Owner {
		return fmt.Errorf("cant order an army that dont belong to you")
	}
	if !moveSet[inputPayload.Ordertype] {
		return fmt.Errorf("invalid order")
	}

	if inputPayload.Ordertype == "move" {
		if a.state.Units[inputPayload.UnitID].Type == "army" && a.state.Board[inputPayload.ToRegion.Name].Sea {
			return fmt.Errorf("cant send an army into the sea")
		}

		if a.state.Units[inputPayload.UnitID].Type == "navy" && !a.state.Board[inputPayload.ToRegion.Name].Sea && !a.state.Board[inputPayload.ToRegion.Name].Coastal {
			return fmt.Errorf("cant send a ship inland")
		}
		if !isConnected(a.state.Board[inputPayload.FromRegion.Name], &inputPayload.ToRegion.Name) {
			return fmt.Errorf("cant move to non adjacent territory")
		}
	}

	if inputPayload.Ordertype == "support move" {
		if !isConnected(a.state.Board[inputPayload.FromRegion.Name], &inputPayload.ToRegion.Name) ||
			!isConnected(a.state.Board[inputPayload.FromRegion.Name], &a.state.Units[inputPayload.UnitID].Position) ||
			!isConnected(a.state.Board[inputPayload.ToRegion.Name], &a.state.Units[inputPayload.UnitID].Position) {
			return fmt.Errorf("cant support move to nor from non adjacent territories")
		}
	}

	if inputPayload.Ordertype == "support hold" {
		if !isConnected(a.state.Board[a.state.Units[inputPayload.UnitID].Position], &inputPayload.ToRegion.Name) {
			return fmt.Errorf("cant support hold to non adjacent territory")
		}
	}

	if inputPayload.Ordertype == "convoy" {
		if a.state.Units[inputPayload.UnitID].Type != "navy" {
			return fmt.Errorf("cant convoy if the unit is not a navy")
		}
		if !isConnected(a.state.Board[inputPayload.FromRegion.Name], &a.state.Units[inputPayload.UnitID].Position) ||
			!isConnected(a.state.Board[inputPayload.ToRegion.Name], &a.state.Units[inputPayload.UnitID].Position) {
			return fmt.Errorf("cant convoy from or to Regions that your sea tile does not touch")
		}
	}

	a.state.Units[inputPayload.UnitID].CurrentOrder = inputPayload
	return nil
}

func isConnected(From *Region, To *string) bool {
	for _, city := range From.Neighbors {
		c := city
		if *c == *To {
			fmt.Println("Match found:", *To)
			return true
		}
	}
	return false
}

func ResolveMovementConflicts(gameState *GameState) {
	for _, unit := range gameState.Units {
		// Skip units with hold orders
		if unit.CurrentOrder.Ordertype == "hold" {
			continue
		}

		// Check if the target region is occupied
		targetRegion := gameState.Board[unit.CurrentOrder.ToRegion.Name]

		switch {
		case unit.CurrentOrder.Ordertype == "move" && !targetRegion.Occupied:
			// An army moving to an unoccupied territory alone
			outcome := ResolveMoveToUnoccupied(unit, gameState)
			UpdateGameState(outcome, gameState)

		case unit.CurrentOrder.Ordertype == "move" && targetRegion.Occupied:
			// An army trying to move to an occupied territory
			outcome := ResolveMoveToOccupied(unit, gameState)
			UpdateGameState(outcome, gameState)

		default:
			// Do nothing for other cases
		}
	}

}

func ResolveMoveToUnoccupied(unit *Unit, gameState *GameState) ConflictOutcome {
	for _, otherUnit := range gameState.Units {
		if unit.ID != otherUnit.ID && otherUnit.CurrentOrder.ToRegion.Name == unit.CurrentOrder.ToRegion.Name {
			// Two different armies trying to move to the same unoccupied territory
			return ResolveConflict(unit, otherUnit, gameState)
		}
	}
	// An army moving to an unoccupied territory alone
	return ConflictOutcome{Winner: unit}
}

func ResolveMoveToOccupied(unit *Unit, gameState *GameState) ConflictOutcome {
	var conflictingUnits []*Unit

	for _, otherUnit := range gameState.Units {
		if unit.ID != otherUnit.ID && otherUnit.CurrentOrder.ToRegion.Name == unit.CurrentOrder.ToRegion.Name {
			// More than one army trying to move to an occupied territory
			conflictingUnits = append(conflictingUnits, otherUnit)
		}
	}

	if len(conflictingUnits) == 0 {
		// An army trying to move to an occupied territory
		return ResolveConflictWithOccupied(unit, gameState)
	}

	// Resolve conflicts with units moving to an occupied territory
	outcomes := make([]ConflictOutcome, 0, len(conflictingUnits)+1)
	outcomes = append(outcomes, ResolveConflictWithOccupied(unit, gameState))
	for _, otherUnit := range conflictingUnits {
		outcomes = append(outcomes, ResolveConflict(unit, otherUnit, gameState))
	}

	// Find the winner among all conflicts
	winner := findWinnerAmongConflicts(outcomes)

	return winner
}

func ResolveConflictWithOccupied(unit *Unit, gameState *GameState) ConflictOutcome {
	for _, otherUnit := range gameState.Units {
		if unit.ID != otherUnit.ID && otherUnit.CurrentOrder.ToRegion.Name == unit.CurrentOrder.ToRegion.Name {
			// Determine the outcome of the conflict
			return ResolveConflict(unit, otherUnit, gameState)
		}
	}
	// No conflict, the unit failed to move in a 1v1
	return ConflictOutcome{Bounced: true}
}

func findWinnerAmongConflicts(outcomes []ConflictOutcome) ConflictOutcome {
	var winner ConflictOutcome
	for _, outcome := range outcomes {
		if outcome.Winner != nil {
			if winner.Winner == nil || outcome.Winner.ID < winner.Winner.ID {
				winner = outcome
			}
		}
	}
	return winner
}

// ResolveConflict resolves conflicts between two units' orders
func ResolveConflict(unit1, unit2 *Unit, gameState *GameState) ConflictOutcome {
	// Determine the strength of each unit based on their orders and support
	strength1 := DetermineStrength(unit1, gameState)
	strength2 := DetermineStrength(unit2, gameState)

	// Determine the outcome of the conflict based on strengths
	if strength1 > strength2 {
		return ConflictOutcome{Winner: unit1, Loser: unit2}
	} else if strength1 < strength2 {
		return ConflictOutcome{Winner: unit2, Loser: unit1}
	} else {
		// If strengths are equal, units bounce back to their original positions
		return ConflictOutcome{Bounced: true}
	}
}

// DetermineStrength determines the strength of a unit based on its order and support
func DetermineStrength(unit *Unit, gameState *GameState) int {
	strength := 1

	// Check the type of order the unit has
	switch unit.CurrentOrder.Ordertype {
	case "move", "hold":
		// For move or hold orders, the strength is simply 1
		// No additional logic needed
	case "support move", "support hold":
		// For support orders, check if there are other units supporting this unit
		supportingUnits := findSupportingUnits(unit, gameState)
		strength += supportingUnits
	case "convoy":
		// For convoy orders, the fleet does not add to strength
	}

	return strength
}

func findSupportingUnits(unit *Unit, gameState *GameState) int {
	support := 0
	for _, sup := range gameState.Units {
		if sup.CurrentOrder.Ordertype == "support move" && sup.CurrentOrder.FromRegion.Name == unit.CurrentOrder.FromRegion.Name && sup.CurrentOrder.ToRegion.Name == unit.CurrentOrder.ToRegion.Name {
			support += 1
		}
		if sup.CurrentOrder.Ordertype == "support hold" && sup.CurrentOrder.FromRegion.Name == unit.Position {
			support += 1
		}

	}
	return support
}

// UpdateGameState updates the game state based on the outcome of a conflict
func UpdateGameState(outcome ConflictOutcome, a *GameState) {
	if outcome.Bounced {
		return
	}

	if outcome.Winner != nil && outcome.Loser != nil {
		a.Units[outcome.Winner.ID].Position = outcome.Winner.CurrentOrder.ToRegion.Name
		a.Board[outcome.Winner.Position].Occupied = true
		a.Board[outcome.Winner.Position].Owner = outcome.Winner.Owner

		a.Units[outcome.Loser.ID].Retreating = true
	} else {
		a.Units[outcome.Winner.ID].Position = outcome.Winner.CurrentOrder.ToRegion.Name
		a.Board[outcome.Winner.Position].Occupied = true
		a.Board[outcome.Winner.Position].Owner = outcome.Winner.Owner
	}
}

func ResetOrders(a *GameApplication) error {
	for _, unit := range a.state.Units {
		unit.CurrentOrder.Ordertype = "hold"
		unit.CurrentOrder.FromRegion.Name = ""
		unit.CurrentOrder.OrderOwner = ""
		unit.CurrentOrder.ToRegion.Name = ""

	}
	return nil
}

func (a *GameApplication) passTurn() error {
	if a.state.Turn == "move" {
		ResolveMovementConflicts(&a.state)
		ResetOrders(a)
		if a.state.MoveCounter {
			a.state.Turn = "build"
			a.state.MoveCounter = false
		} else {
			a.state.MoveCounter = true
		}
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
