package main

import (
	"fmt"

	"github.com/rollmelette/rollmelette"
)

func (a *GameApplication) handleRetreat(
	metadata rollmelette.Metadata,
	inputPayload RetreatOrderPayload,
) error {

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

	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println(orders)
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")
	fmt.Println("ahahahaha")

	a.state.Units[inputPayload.UnitID].CurrentOrder = orders

	fmt.Println(a.state.Units[inputPayload.UnitID].CurrentOrder)

	return nil
}

func resolveRetreats(a *GameApplication) {
	// Iterate through the units and handle their retreat orders
	for _, unit := range a.state.Units {
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println(unit.CurrentOrder.Ordertype)
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		fmt.Println("ahahahaha")
		switch unit.CurrentOrder.Ordertype {
		case "hold":
			// Do nothing for hold orders
			continue
		case "delete":
			fmt.Println("Deleting unit:", unit.ID)

			// Delete unit from player's armies
			delete(a.state.Players[unit.Owner].Armies, unit.ID)

			// Delete unit from the game's state
			delete(a.state.Units, unit.ID)

		case "move":
			//check for bouncing retreats
			for _, other := range a.state.Units {
				//both retreating units tryed to move to the same place and are deleted instead
				if unit.CurrentOrder.Ordertype == "move" && unit.ID != other.ID {
					delete(a.state.Players[unit.Owner].Armies, unit.ID)
					delete(a.state.Players[unit.Owner].Armies, other.ID)
					delete(a.state.Units, unit.ID)
					delete(a.state.Units, other.ID)
				}
			}
			// Execute the move order
			if !a.state.Board[unit.CurrentOrder.ToRegion].Occupied {
				fmt.Println("Moving unit", unit.ID, "to", unit.CurrentOrder.ToRegion)

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

func setForDelete(a *GameApplication) {
	for _, unit := range a.state.Units {
		if unit.Retreating != "" {
			unit.CurrentOrder.Ordertype = "delete"
		}
	}
}
