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
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")
	fmt.Println(a.state.Units)
	fmt.Println("kkkkkkkk")
	fmt.Println("kkkkkkkk")

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
					Owner:    order.Player,
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

func (a *GameApplication) handleRetreat(
	metadata rollmelette.Metadata,
	inputPayload RetreatOrderPayload,
) error {

	fmt.Println("ttttttt")
	fmt.Println("ttttttt")
	fmt.Println("ttttttt")
	fmt.Println("ttttttt")
	fmt.Println("ttttttt")
	fmt.Println("ttttttt")
	fmt.Println("ttttttt")
	if a.state.Turn != "retreats" {
		return fmt.Errorf("can't issue a retreat order outside retreating phase")
	}
	unit, ok := a.state.Units[inputPayload.UnitID]
	if !ok {
		return fmt.Errorf("unit not found")
	}
	if metadata.MsgSender != unit.Owner {
		return fmt.Errorf("can't retreat another player's unit")
	}

	// Check if the target region is the current position or the forward (defeated from) position
	if inputPayload.ToRegion == unit.Position {
		return fmt.Errorf("can't retreat to the same place")
	}
	if inputPayload.ToRegion == unit.Retreating {
		return fmt.Errorf("can't retreat forward to the attacking region")
	}

	orderType := "move"

	// Check for delete flag
	if inputPayload.Delete {
		orderType = "delete"
	}

	// Check if target region is occupied

	if orderType == "move" {
		if a.state.Board[inputPayload.ToRegion].Occupied {
			return fmt.Errorf("can't retreat to an occupied region")
		}

		// Check if the retreating region is connected
		if !isConnected(a.state.Board[unit.Position], &inputPayload.ToRegion) {
			return fmt.Errorf("can't retreat to non-adjacent region")
		}
	}

	orders := Orders{
		UnitID:      inputPayload.UnitID,
		Ordertype:   orderType,
		ToRegion:    inputPayload.ToRegion,
		ToSubRegion: inputPayload.ToSubRegion,
	}

	a.state.Units[inputPayload.UnitID].CurrentOrder = orders

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
		"convoy move":  true,
	}

	if a.state.Turn != "move" {
		return fmt.Errorf("can't move an army outside of movement phase")
	}
	if _, ok := a.state.Players[metadata.MsgSender].Armies[inputPayload.UnitID]; !ok {
		return fmt.Errorf("can't move another player's army")
	}
	if !a.state.Board[inputPayload.FromRegion].Occupied {
		return fmt.Errorf("cant order an army to move from an empty region")
	}
	if metadata.MsgSender != a.state.Units[inputPayload.UnitID].Owner {
		return fmt.Errorf("cant order an army that dont belong to you")
	}
	if !moveSet[inputPayload.Ordertype] {
		return fmt.Errorf("invalid order")
	}

	if inputPayload.Ordertype == "move" {
		if a.state.Units[inputPayload.UnitID].Position != inputPayload.FromRegion {
			return fmt.Errorf("your army is not there")
		}
		if a.state.Units[inputPayload.UnitID].Type == "army" && a.state.Board[inputPayload.ToRegion].Sea {
			return fmt.Errorf("cant send an army into the sea")
		}

		if a.state.Units[inputPayload.UnitID].Type == "navy" && !a.state.Board[inputPayload.ToRegion].Sea && !a.state.Board[inputPayload.ToRegion].Coastal {
			return fmt.Errorf("cant send a ship inland")
		}
		if !isConnected(a.state.Board[inputPayload.FromRegion], &inputPayload.ToRegion) {
			return fmt.Errorf("cant move to non adjacent territory")
		}
		MoveHarbor := false
		for _, region := range SubRegionsList {
			if (region == inputPayload.ToRegion || region == inputPayload.FromRegion) && a.state.Units[inputPayload.UnitID].Type == "navy" {
				MoveHarbor = true
			}
		}
		//navy stationed in/moving into one of the sub regions
		if MoveHarbor {
			if (inputPayload.FromSubRegion != "" && inputPayload.ToSubRegion != "") != (inputPayload.FromSubRegion == "" && inputPayload.ToSubRegion == "") {
				return fmt.Errorf("need to specify the sub region and can't move directly between sub regions")
			}
			if inputPayload.FromSubRegion != "" {
				if !isSubRegionConnected(a.state.Board[inputPayload.FromRegion].SubRegions[inputPayload.FromSubRegion], inputPayload.ToRegion) {
					return fmt.Errorf("cant reach this region from this harbor")
				}
			} else {
				if !isSubRegionConnected(a.state.Board[inputPayload.ToRegion].SubRegions[inputPayload.ToSubRegion], inputPayload.FromRegion) {
					return fmt.Errorf("cant move to non adjacent harbor")
				}
			}
		} else {
			inputPayload.FromSubRegion = ""
			inputPayload.ToSubRegion = ""
		}

	}

	if inputPayload.Ordertype == "support move" {
		if !isConnected(a.state.Board[inputPayload.FromRegion], &inputPayload.ToRegion) ||
			!isConnected(a.state.Board[inputPayload.ToRegion], &a.state.Units[inputPayload.UnitID].Position) {
			return fmt.Errorf("cant support move to nor from non adjacent territories")
		}
	}

	if inputPayload.Ordertype == "support hold" {
		if !isConnected(a.state.Board[a.state.Units[inputPayload.UnitID].Position], &inputPayload.ToRegion) {
			return fmt.Errorf("cant support hold to non adjacent territory")
		}
	}

	if inputPayload.Ordertype == "convoy" {
		if a.state.Units[inputPayload.UnitID].Type != "navy" || !a.state.Board[a.state.Units[inputPayload.UnitID].Position].Sea {
			return fmt.Errorf("cant convoy if the unit is not at sea")
		}
		if !isConnected(a.state.Board[inputPayload.FromRegion], &a.state.Units[inputPayload.UnitID].Position) ||
			!isConnected(a.state.Board[inputPayload.ToRegion], &a.state.Units[inputPayload.UnitID].Position) {
			return fmt.Errorf("cant convoy from or to Regions that your sea tile does not touch")
		}
	}

	if inputPayload.Ordertype == "convoy move" {
		if a.state.Units[inputPayload.UnitID].Type != "army" {
			return fmt.Errorf("cant convoy another boat")
		}
		if !a.state.Board[inputPayload.FromRegion].Coastal || !a.state.Board[inputPayload.ToRegion].Coastal {
			return fmt.Errorf("cant convoy from nor to landlocked regions")
		}
		var seaConnected []string
		for _, region := range a.state.Board[inputPayload.FromRegion].Neighbors {
			if a.state.Board[*region].Sea && a.state.Board[*region].Occupied {
				seaConnected = append(seaConnected, *region)
			}
		}
		if len(seaConnected) < 1 {
			return fmt.Errorf("no available boats to convoy")
		}
		connectedBySea := false
		fmt.Println("mar con")
		for _, sea := range seaConnected {
			fmt.Println("mar con", sea)
			for _, coast := range a.state.Board[sea].Neighbors {
				fmt.Println("com quem", *coast, string(a.state.Board[inputPayload.ToRegion].Name))
				if *coast == a.state.Board[inputPayload.ToRegion].Name {
					connectedBySea = true
				}
			}
		}
		if !connectedBySea {
			return fmt.Errorf("cant convoy to a coast more than one sea tile away")
		}
	}
	orders := Orders{
		UnitID:        inputPayload.UnitID,
		Ordertype:     inputPayload.Ordertype,
		OrderOwner:    inputPayload.OrderOwner,
		ToRegion:      inputPayload.ToRegion,
		ToSubRegion:   inputPayload.ToSubRegion,
		FromRegion:    inputPayload.FromRegion,
		FromSubRegion: inputPayload.FromSubRegion,
	}
	a.state.Units[inputPayload.UnitID].CurrentOrder = orders

	return nil
}

func resolveRetreats(a *GameApplication) {
	// Iterate through the units and handle their retreat orders
	for _, unit := range a.state.Units {
		switch unit.CurrentOrder.Ordertype {
		case "hold":
			// Do nothing for hold orders
			continue
		case "delete":
			fmt.Println("Deleting unit:", unit.ID)

			// Delete unit from player's armies
			if _, exists := a.state.Players[unit.Owner].Armies[unit.ID]; exists {
				delete(a.state.Players[unit.Owner].Armies, unit.ID)
			}

			// Delete unit from the game's state
			if _, exists := a.state.Units[unit.ID]; exists {
				delete(a.state.Units, unit.ID)
			}

		case "move":
			// Execute the move order
			if !a.state.Board[unit.CurrentOrder.ToRegion].Occupied {
				fmt.Println("Moving unit", unit.ID, "to", unit.CurrentOrder.ToRegion)

				a.state.Board[unit.Position].Occupied = false
				a.state.Board[unit.CurrentOrder.ToRegion].Occupied = true

				unit.Position = unit.CurrentOrder.ToRegion
				unit.CurrentOrder.Ordertype = "hold"
				unit.Retreating = ""

				// Update Unit's SubPosition if needed
				unit.SubPosition = unit.CurrentOrder.ToSubRegion
			} else {
				fmt.Println("Failed to move unit", unit.ID, "to occupied region", unit.CurrentOrder.ToRegion)
			}
		}
	}
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

func (a *GameApplication) prepareMoves() []MoveOrder {
	var moveOrders []MoveOrder
	for _, unit := range a.state.Units {
		if unit.CurrentOrder.Ordertype == "move" || unit.CurrentOrder.Ordertype == "convoy move" {
			if unit.CurrentOrder.Ordertype == "convoy move" {
				goodConvoy := false
				convoyPosition := ""
				for _, convoyUnit := range a.state.Units {
					if convoyUnit.CurrentOrder.FromRegion == unit.CurrentOrder.FromRegion && convoyUnit.CurrentOrder.Ordertype == "convoy" && convoyUnit.CurrentOrder.ToRegion == unit.CurrentOrder.ToRegion {
						//convoy is executed
						goodConvoy = true
						convoyPosition = convoyUnit.Position
					}

				}
				if goodConvoy {
					convoyAttacked := false
					for _, otherUnit := range a.state.Units {
						if otherUnit.CurrentOrder.Ordertype == "move" && otherUnit.CurrentOrder.ToRegion == convoyPosition {
							//convoy is attacked
							convoyAttacked = true
						}
					}
					if !convoyAttacked {
						moveOrders = append(moveOrders, MoveOrder{
							Unit:       unit,
							FromRegion: unit.Position,
							ToRegion:   unit.CurrentOrder.ToRegion,
						})
					}
				}
			} else {
				moveOrders = append(moveOrders, MoveOrder{
					Unit:       unit,
					FromRegion: unit.Position,
					ToRegion:   unit.CurrentOrder.ToRegion,
					SubRegion:  unit.CurrentOrder.ToSubRegion + unit.CurrentOrder.FromSubRegion,
				})
			}
		}
	}
	return moveOrders
}

func calculateSupportCount(unit *Unit, gameState *GameState) int {
	supportCount := 0
	for _, supportingUnit := range gameState.Units {
		if supportingUnit.CurrentOrder.Ordertype == "support move" &&
			supportingUnit.CurrentOrder.ToRegion == unit.CurrentOrder.ToRegion &&
			supportingUnit.CurrentOrder.FromRegion == unit.CurrentOrder.FromRegion {
			supportCount++
		} else if supportingUnit.CurrentOrder.Ordertype == "support hold" &&
			supportingUnit.CurrentOrder.ToRegion == unit.Position {
			supportCount++
		}
	}
	return supportCount
}

func (a *GameApplication) executeMoves(moveOrders []MoveOrder) {
	destinationMap := make(map[string][]*Unit)
	for _, moveOrder := range moveOrders {
		destinationMap[moveOrder.ToRegion] = append(destinationMap[moveOrder.ToRegion], moveOrder.Unit)
	}

	for _, moveOrder := range moveOrders {
		fmt.Println("move orders", moveOrder.ToRegion)
		unitsMovingToDestination := destinationMap[moveOrder.ToRegion]
		if len(unitsMovingToDestination) == 1 && !a.state.Board[moveOrder.ToRegion].Occupied {
			// No conflict, move the unit
			fmt.Println("move to", moveOrder.ToRegion)
			fmt.Println("move from", moveOrder.ToRegion)
			moveOrder.Unit.Position = moveOrder.ToRegion
			a.state.Board[moveOrder.ToRegion].Occupied = true
			a.state.Board[moveOrder.FromRegion].Occupied = false
			a.state.Units[moveOrder.Unit.ID].Position = moveOrder.ToRegion
			a.state.Units[moveOrder.Unit.ID].SubPosition = moveOrder.SubRegion

			a.state.Units[moveOrder.Unit.ID].CurrentOrder.Ordertype = "hold"
			a.state.Units[moveOrder.Unit.ID].CurrentOrder.FromRegion = ""
			a.state.Units[moveOrder.Unit.ID].CurrentOrder.OrderOwner = ""
			a.state.Units[moveOrder.Unit.ID].CurrentOrder.ToRegion = ""

			fmt.Println("occupied", a.state.Board[moveOrder.FromRegion].Occupied, moveOrder.FromRegion)
		}
	}
}

func ResolveMovementConflicts(gameState *GameState) {
	destinationMap := make(map[string][]*Unit)

	fmt.Println("Serbia occupied: ", gameState.Board["Serbia"].Occupied)
	fmt.Println("Budapest occupied: ", gameState.Board["Budapest"].Occupied)

	for _, unit := range gameState.Units {
		// Skip units with hold orders
		if unit.CurrentOrder.Ordertype == "hold" {
			continue
		}
		// Add units to the destination map
		if unit.CurrentOrder.Ordertype == "move" {
			fmt.Println("unit move order ", unit.CurrentOrder.ToRegion, unit.ID)
			destinationMap[unit.CurrentOrder.ToRegion] = append(destinationMap[unit.CurrentOrder.ToRegion], unit)
		}
	}

	for targetRegion, units := range destinationMap {
		if len(units) > 1 {
			// Multiple units trying to move to the same region
			targetRegionData := gameState.Board[targetRegion]
			var outcome ConflictOutcome

			if targetRegionData.Occupied {
				occupyingUnit := gameState.getUnitAtPosition(targetRegion)
				outcome = ResolveMoveToOccupied(units, occupyingUnit, gameState)
			} else {
				outcome = ResolveMoveToUnoccupied(units, gameState)
			}

			UpdateGameState(outcome, gameState)
		} else if len(units) == 1 {
			// Single unit trying to move to the region
			unit := units[0]
			if !gameState.Board[unit.CurrentOrder.ToRegion].Occupied {
				//Free movement
				outcome := ConflictOutcome{
					Winner:  unit,
					Bounced: false,
				}
				UpdateGameState(outcome, gameState)
			} else {
				occupyingUnit := gameState.getUnitAtPosition(unit.CurrentOrder.ToRegion)
				outcome := ResolveMoveToOccupied([]*Unit{unit}, occupyingUnit, gameState)
				UpdateGameState(outcome, gameState)
			}
		}
	}
}

func ResolveMoveToOccupied(movingUnits []*Unit, occupyingUnit *Unit, gameState *GameState) ConflictOutcome {
	fmt.Println("Occupied: ", len(movingUnits), occupyingUnit.Position)
	occupyingSupportCount := calculateSupportCount(occupyingUnit, gameState)
	fmt.Println("sup hold: ", occupyingSupportCount)
	var maxSupport int
	var winningUnit *Unit
	tie := false

	for _, unit := range movingUnits {
		supportCount := calculateSupportCount(unit, gameState)
		if supportCount > maxSupport {
			maxSupport = supportCount
			winningUnit = unit
			tie = false
		} else if supportCount == maxSupport {
			tie = true
		}
	}

	if tie || maxSupport <= occupyingSupportCount {
		// If there's a tie or the occupying unit has equal or more support, all moves fail
		return ConflictOutcome{
			Winner:  occupyingUnit,
			Bounced: true,
		}
	} else {
		// The unit with the highest support wins and moves
		winningUnit.Position = winningUnit.CurrentOrder.ToRegion
		gameState.Board[winningUnit.CurrentOrder.ToRegion].Occupied = true
		gameState.Board[winningUnit.CurrentOrder.FromRegion].Occupied = false
		occupyingUnit.Retreating = winningUnit.CurrentOrder.FromRegion

		return ConflictOutcome{
			Winner:  winningUnit,
			Loser:   occupyingUnit,
			Bounced: false,
		}
	}
}

func ResolveMoveToUnoccupied(units []*Unit, gameState *GameState) ConflictOutcome {
	var maxSupport int
	var winningUnit *Unit
	tie := false

	for _, unit := range units {
		supportCount := calculateSupportCount(unit, gameState)
		if supportCount > maxSupport {
			maxSupport = supportCount
			winningUnit = unit
			tie = false
		} else if supportCount == maxSupport {
			tie = true
		}
	}

	if tie {
		// If there's a tie, all moves fail
		for _, unit := range units {
			gameState.Board[unit.CurrentOrder.ToRegion].Occupied = false
		}
		return ConflictOutcome{
			Bounced: true,
		}
	} else {
		// The unit with the highest support moves
		winningUnit.Position = winningUnit.CurrentOrder.ToRegion
		gameState.Board[winningUnit.CurrentOrder.ToRegion].Occupied = true
		gameState.Board[winningUnit.CurrentOrder.FromRegion].Occupied = false

		return ConflictOutcome{
			Winner:  winningUnit,
			Bounced: false,
		}
	}
}

func (g *GameState) getUnitAtPosition(position string) *Unit {
	for _, unit := range g.Units {
		if unit.Position == position {
			return unit
		}
	}
	return nil
}

// UpdateGameState updates the game state based on the outcome of a conflict
func UpdateGameState(outcome ConflictOutcome, a *GameState) {
	if outcome.Bounced {
		return
	}

	if outcome.Winner != nil && outcome.Loser != nil {
		a.Units[outcome.Winner.ID].Position = outcome.Winner.CurrentOrder.ToRegion
		a.Units[outcome.Winner.ID].SubPosition = outcome.Winner.CurrentOrder.ToSubRegion
		a.Board[outcome.Winner.Position].Occupied = true
		a.Board[outcome.Winner.Position].Owner = a.Players[outcome.Winner.Owner].Name

		if a.Units[outcome.Loser.ID].Position == a.Units[outcome.Winner.ID].CurrentOrder.ToRegion {
			a.Units[outcome.Loser.ID].Retreating = a.Units[outcome.Winner.ID].CurrentOrder.FromRegion
		}
	} else {
		a.Units[outcome.Winner.ID].Position = outcome.Winner.CurrentOrder.ToRegion
		a.Board[outcome.Winner.Position].Occupied = true
		a.Board[outcome.Winner.Position].Owner = a.Players[outcome.Winner.Owner].Name
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

func setForDelete(a *GameApplication) {
	for _, unit := range a.state.Units {
		if unit.Retreating != "" {
			unit.CurrentOrder.Ordertype = "delete"
		}
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
			}
		}
	} else if a.state.Turn == "build" {
		BuildUnits(a)
		a.state.Turn = "move"
	} else if a.state.Turn == "retreats" {
		setForDelete(a)
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
