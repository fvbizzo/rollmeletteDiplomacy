package main

import (
	"fmt"

	"github.com/rollmelette/rollmelette"
)

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
