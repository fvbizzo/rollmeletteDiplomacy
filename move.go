package main

import (
	"fmt"

	"github.com/rollmelette/rollmelette"
)

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
