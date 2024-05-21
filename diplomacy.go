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
	ID           int    `json:"ID"`
	Type         string `json:"type"`
	Position     string `json:"position"`
	Owner        string `json:"owner"`
	CurrentOrder Orders `json:"currentOrder"`
	Retreating   string `json:"retreating"`
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
	ToRegion   string `json:"toRegion"`
	FromRegion string `json:"fromRegion"`
}

type RetreatOrderPayload struct {
	UnitID   int    `json:"unitID"`
	Delete   bool   `json:"delete"`
	ToRegion string `json:"toRegion"`
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
		err = a.Retreat(metadata, inputPayload)
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
	if len(a.state.Players[metadata.MsgSender].Armies) >= a.state.Players[metadata.MsgSender].Bases && inputPayload.Delete == 0 {
		return fmt.Errorf(("cant build another army without extra supply centers"))
	}

	build := BuildArmyInput{
		Info:   inputPayload,
		Player: metadata.MsgSender,
	}
	a.state.Players[metadata.MsgSender].Builds = append(a.state.Players[metadata.MsgSender].Builds, &build)

	return nil
}

func BuildUnits(a *GameApplication) {

	for _, player := range a.state.Players {
		if len(player.Builds) == 0 {
			continue
		}
		for _, order := range player.Builds {
			if order.Info.Delete != 0 {
				a.state.Board[order.Info.Position].Occupied = false
				delete(a.state.Players[order.Player].Armies, order.Info.Delete)
				delete(a.state.Units, order.Info.Delete)
			} else {
				a.state.Board[order.Info.Position].Occupied = true
				a.state.Players[order.Player].Armies[UnitID] = order.Info.Position
				a.state.Units[UnitID] = &Unit{
					ID:       UnitID,
					Type:     order.Info.Type,
					Position: order.Info.Position,
					Owner:    order.Info.Owner,
					CurrentOrder: Orders{
						UnitID:     UnitID,
						Ordertype:  "hold",
						OrderOwner: "",
					},
				}

				UnitID += 1
			}
		}
		player.Builds = nil
	}
}

func (a *GameApplication) Retreat(
	metadata rollmelette.Metadata,
	inputPayload RetreatOrderPayload,
) error {

	if a.state.Turn != "retreats" {
		return fmt.Errorf("cant issue a retreat order outside retreating phase")
	}
	if a.state.Players[metadata.MsgSender].Name != a.state.Units[inputPayload.UnitID].Owner {
		return fmt.Errorf("cant retreat another player's unit")
	}

	return nil
}

func handleRetreat(a *GameState) {

}

func (a *GameApplication) handleMoveArmy(
	metadata rollmelette.Metadata,
	inputPayload GiveOrderPayload,
) error {
	if a.state.Turn != "move" {
		return fmt.Errorf("can't move an army outside of movement phase")
	}
	if _, ok := a.state.Players[metadata.MsgSender].Armies[inputPayload.UnitID]; !ok {
		return fmt.Errorf("can't move another player's army")
	}
	orders := Orders{
		UnitID:     inputPayload.UnitID,
		Ordertype:  inputPayload.Ordertype,
		OrderOwner: inputPayload.OrderOwner,
		ToRegion:   inputPayload.ToRegion,
		FromRegion: inputPayload.FromRegion,
	}
	a.state.Units[inputPayload.UnitID].CurrentOrder = orders

	return nil
}

func (a *GameApplication) ResolveMovementConflicts() {
	supportMoveOrders := a.processSupportMoveOrders()
	supportHoldOrders := a.processSupportHoldOrders()

	moveOrders := make([]*Unit, 0)
	for _, unit := range a.state.Units {
		if unit.CurrentOrder.Ordertype == "move" {
			moveOrders = append(moveOrders, unit)
		}
	}

	executedMoves := make(map[string]bool)
	for len(moveOrders) > 0 {
		remainingMoves := make([]*Unit, 0)
		for _, unit := range moveOrders {
			if executedMoves[unit.CurrentOrder.FromRegion] || executedMoves[unit.CurrentOrder.ToRegion] {
				remainingMoves = append(remainingMoves, unit)
				continue
			}
			a.executeMove(unit, supportMoveOrders, supportHoldOrders)
			executedMoves[unit.CurrentOrder.FromRegion] = true
			executedMoves[unit.CurrentOrder.ToRegion] = true
		}
		if len(remainingMoves) == len(moveOrders) {
			break
		}
		moveOrders = remainingMoves
	}

	for _, unit := range moveOrders {
		a.executeMove(unit, supportMoveOrders, supportHoldOrders)
	}
}

func (a *GameApplication) executeMove(unit *Unit, supportMoveOrders, supportHoldOrders map[string]int) {
	targetRegion := unit.CurrentOrder.ToRegion
	attackers := []*Unit{unit}

	// Check for other units attempting to move to the same region
	for _, otherUnit := range a.state.Units {
		if otherUnit.CurrentOrder.Ordertype == "move" && otherUnit.CurrentOrder.ToRegion == targetRegion && otherUnit.ID != unit.ID {
			// If two units attempt to move to the same region, both bounce
			unit.Position = unit.CurrentOrder.FromRegion
			a.state.Board[unit.CurrentOrder.FromRegion].Occupied = true
			return
		}
	}

	// Handle normal movement logic
	if len(attackers) == 1 && !a.state.Board[targetRegion].Occupied {
		unit.Position = targetRegion
		a.state.Board[targetRegion].Occupied = true
		a.state.Board[unit.CurrentOrder.FromRegion].Occupied = false
	} else {
		outcome := a.ResolveMoveToOccupied(attackers, supportMoveOrders, supportHoldOrders, targetRegion)
		if outcome.Bounced {
			for _, attacker := range attackers {
				attacker.Position = attacker.CurrentOrder.FromRegion
			}
		} else {
			outcome.Winner.Position = targetRegion
			a.state.Board[targetRegion].Occupied = true
			a.state.Board[outcome.Winner.CurrentOrder.FromRegion].Occupied = false
			if outcome.Loser != nil {
				outcome.Loser.Retreating = outcome.Loser.Position
				a.state.Board[outcome.Loser.Position].Occupied = false
			}
		}
	}
}

func (a *GameApplication) ResolveMoveToOccupied(attackers []*Unit, supportMoveOrders, supportHoldOrders map[string]int, targetRegion string) ConflictOutcome {
	defenderUnit := a.getUnitAtPosition(targetRegion)
	if defenderUnit != nil {
		attackers = append(attackers, defenderUnit)
	}

	strongest := attackers[0]
	for _, attacker := range attackers[1:] {
		strongest = a.ResolveConflict(strongest, attacker, supportMoveOrders, supportHoldOrders, targetRegion)
	}

	if defenderUnit != nil && strongest == defenderUnit {
		return ConflictOutcome{Bounced: true}
	}

	var loser *Unit
	for _, attacker := range attackers {
		if attacker != strongest {
			loser = attacker
			break
		}
	}

	return ConflictOutcome{
		Winner: strongest,
		Loser:  loser,
	}
}

func (a *GameApplication) ResolveConflict(unit1, unit2 *Unit, supportMoveOrders, supportHoldOrders map[string]int, targetRegion string) *Unit {
	unit1Strength := 1 + supportMoveOrders[unit1.CurrentOrder.FromRegion]
	unit2Strength := 1 + supportMoveOrders[unit2.CurrentOrder.FromRegion]

	if unit1.Position == targetRegion {
		unit1Strength += supportHoldOrders[targetRegion]
	}
	if unit2.Position == targetRegion {
		unit2Strength += supportHoldOrders[targetRegion]
	}

	if unit1Strength > unit2Strength {
		return unit1
	}
	if unit2Strength > unit1Strength {
		return unit2
	}

	if unit1.Position == targetRegion {
		return unit1
	}
	if unit2.Position == targetRegion {
		return unit2
	}

	return nil
}

func (a *GameApplication) getUnitAtPosition(position string) *Unit {
	for _, unit := range a.state.Units {
		if unit.Position == position {
			return unit
		}
	}
	return nil
}

func (a *GameApplication) processSupportMoveOrders() map[string]int {
	supportOrders := make(map[string]int)

	for _, unit := range a.state.Units {
		if unit.CurrentOrder.Ordertype == "support move" {
			fromRegion := unit.CurrentOrder.FromRegion
			supportOrders[fromRegion]++
		}
	}

	return supportOrders
}

func (a *GameApplication) processSupportHoldOrders() map[string]int {
	supportOrders := make(map[string]int)

	for _, unit := range a.state.Units {
		if unit.CurrentOrder.Ordertype == "support hold" {
			toRegion := unit.CurrentOrder.ToRegion
			supportOrders[toRegion]++
		}
	}

	return supportOrders
}

func (a *GameApplication) ExecuteOrders() {
	supportMoveOrders := a.processSupportMoveOrders()
	supportHoldOrders := a.processSupportHoldOrders()

	for _, unit := range a.state.Units {
		if unit.CurrentOrder.Ordertype == "move" {
			a.executeMove(unit, supportMoveOrders, supportHoldOrders)
		}
	}
}

// UpdateGameState updates the game state based on the outcome of a conflict
func UpdateGameState(outcome ConflictOutcome, a *GameState) {
	if outcome.Bounced {
		return
	}

	if outcome.Winner != nil && outcome.Loser != nil {
		a.Units[outcome.Winner.ID].Position = outcome.Winner.CurrentOrder.ToRegion
		a.Board[outcome.Winner.Position].Occupied = true
		a.Board[outcome.Winner.Position].Owner = outcome.Winner.Owner

		if a.Units[outcome.Loser.ID].Position == a.Units[outcome.Winner.ID].CurrentOrder.ToRegion {
			a.Units[outcome.Loser.ID].Retreating = a.Units[outcome.Winner.ID].Position
		}
	} else {
		a.Units[outcome.Winner.ID].Position = outcome.Winner.CurrentOrder.ToRegion
		a.Board[outcome.Winner.Position].Occupied = true
		a.Board[outcome.Winner.Position].Owner = outcome.Winner.Owner
	}
}

func ResetOrders(a *GameApplication) {
	for _, unit := range a.state.Units {
		unit.CurrentOrder.Ordertype = "hold"
		unit.CurrentOrder.FromRegion = ""
		unit.CurrentOrder.OrderOwner = ""
		unit.CurrentOrder.ToRegion = ""

	}
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
	a.passTurn()
	return nil
}

func (a *GameApplication) passTurn() error {
	for _, player := range a.state.Players {
		player.Ready = false
	}
	if a.state.Turn == "move" {
		a.ResolveMovementConflicts()
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
			}
		}
	} else if a.state.Turn == "build" {
		BuildUnits(a)
		a.state.Turn = "move"
	} else if a.state.Turn == "retreats" {
		handleRetreat(&a.state)
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
