package main

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gligneul/rollmelette"
	"github.com/stretchr/testify/suite"
)

var payload = common.Hex2Bytes("deadbeef")

// var msgSender = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafafa")
var Austria = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf1")
var England = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf2")
var France = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf3")
var Germany = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf4")
var Italy = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf5")
var Russia = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf6")
var Turkey = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf7")

var PassTurnPayloadSetup = []byte(`{"kind": "ReadyOrders", "payload": ""}`)

var currentState GameState

func TestMyApplicationSuite(t *testing.T) {
	suite.Run(t, new(MyApplicationSuite))
}

type MyApplicationSuite struct {
	suite.Suite
	tester *rollmelette.Tester
}

func (s *MyApplicationSuite) SetupTest() {
	app := NewGameApplication(Austria, England, France, Germany, Italy, Russia, Turkey, 5)
	s.tester = rollmelette.NewTester(app)
}

func (s *MyApplicationSuite) PassTurn() ([]byte, error) {
	result := s.tester.Advance(Austria, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	result = s.tester.Advance(England, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	result = s.tester.Advance(France, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	result = s.tester.Advance(Germany, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	result = s.tester.Advance(Italy, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	result = s.tester.Advance(Russia, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	result = s.tester.Advance(Turkey, PassTurnPayloadSetup)
	if result.Err != nil {
		return result.Reports[0].Payload, result.Err
	}
	return result.Reports[0].Payload, nil
}

func (s *MyApplicationSuite) TestDeleteArmy() {

	s.PassTurn()
	s.PassTurn()

	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 4}}`
	result := s.tester.Advance(England, []byte(input))
	s.Nil(result.Err)
}

func (s *MyApplicationSuite) TestPassMoveTurn() {
	result := s.tester.Advance(England, PassTurnPayloadSetup)
	s.Nil(result.Err)
}

func (s *MyApplicationSuite) TestDeleteArmyWhereThereIsNone() {

	s.PassTurn()
	s.PassTurn()

	preinput := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 4}}`
	s.tester.Advance(England, []byte(preinput))
	s.PassTurn()
	s.PassTurn()
	s.PassTurn()

	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 4}}`
	result := s.tester.Advance(England, []byte(input))
	s.ErrorContains(result.Err, "cant delete an army in empty region")
}

// Testing the build army function
func (s *MyApplicationSuite) TestBuildArmy() {

	s.PassTurn()
	s.PassTurn()

	preinput := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 4}}`
	s.tester.Advance(England, []byte(preinput))

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	//check if the unit has been deleted
	_, ok := currentState.Units[4]

	s.Equal(false, ok)
	s.Nil(result)

	s.PassTurn()
	s.PassTurn()

	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 0}}`
	r := s.tester.Advance(England, []byte(input))
	s.Nil(r.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	expect := Unit{
		ID:       23,
		Type:     "army",
		Position: "London",
		Owner:    "England",
		CurrentOrder: Orders{
			UnitID:     23,
			Ordertype:  "hold",
			OrderOwner: "",
			ToRegion:   "",
			FromRegion: "",
		},
	}

	s.Equal(&expect, currentState.Units[23])
	s.Nil(result)
}

// Testing a player trying to build an army outside build phase
func (s *MyApplicationSuite) TestBuildArmyOutsideBuildPhase() {
	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 0}}`
	result := s.tester.Advance(England, []byte(input))
	s.ErrorContains(result.Err, "cant build an army outside build")

}

// Trying to build another player Army
func (s *MyApplicationSuite) TestBuildanotherPlayerArmy() {

	s.PassTurn()
	s.PassTurn()

	preinput := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": 4}}`
	s.tester.Advance(England, []byte(preinput))

	s.PassTurn()
	s.PassTurn()
	s.PassTurn()

	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "France", "Delete": 0}}`
	result := s.tester.Advance(England, []byte(input))
	s.ErrorContains(result.Err, "cant build another player's army")
}

func (s *MyApplicationSuite) TestMoveArmy() {
	input := `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "move", "OrderOwner": "England", "ToRegion": "Wales", "FromRegion": "London"}}`
	s.tester.Advance(England, []byte(input))

	report, result := s.PassTurn()
	s.Nil(result)

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Wales", currentState.Units[4].Position)
}

func (s *MyApplicationSuite) TestUnitBounce() {

	input := `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "move", "OrderOwner": "England", "ToRegion": "English Channel", "FromRegion": "London"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 8, "OrderType": "move", "OrderOwner": "France", "ToRegion": "English Channel", "FromRegion": "Brest"}}`

	r1 := s.tester.Advance(England, []byte(input))
	r2 := s.tester.Advance(France, []byte(input2))

	s.Nil(r1.Err)
	s.Nil(r2.Err)

	json.Unmarshal([]byte(r2.Reports[0].Payload), &currentState)

	//checking if the orders were issued
	expectedUnitsOrders1 := Orders{
		UnitID:     4,
		Ordertype:  "move",
		OrderOwner: "England",
		ToRegion:   "English Channel",
		FromRegion: "London",
	}

	expectedUnitsOrders2 := Orders{
		UnitID:     8,
		Ordertype:  "move",
		OrderOwner: "France",
		ToRegion:   "English Channel",
		FromRegion: "Brest",
	}

	s.Equal(expectedUnitsOrders1, currentState.Units[4].CurrentOrder)
	s.Equal(expectedUnitsOrders2, currentState.Units[8].CurrentOrder)

	report, result := s.PassTurn()

	s.Nil(result)

	json.Unmarshal([]byte(string(report)), &currentState)

	//Both units should bounce and be at the place they started
	s.Equal("London", currentState.Units[4].Position)
	s.Equal("Brest", currentState.Units[8].Position)
}

func (s *MyApplicationSuite) TestSupportMove() {
	//Setting the first unit to tyrolia and trying to invade venice 1x1
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Tyrolia", "FromRegion": "Vienna"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 3, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Trieste"}}`

	r1 := s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)
	r2 := s.tester.Advance(Austria, []byte(input2))
	s.Nil(r2.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Tyrolia", currentState.Units[1].Position)
	s.Equal("Trieste", currentState.Units[3].Position)
	s.Nil(result)

	//invading Venice from Tyrolia and getting support from Trieste
	input3 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Tyrolia"}}`
	input4 := `{"kind": "MoveArmy", "payload" : {"UnitID": 3, "OrderType": "support move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Tyrolia"}}`

	r3 := s.tester.Advance(Austria, []byte(input3))
	s.Nil(r3.Err)
	r4 := s.tester.Advance(Austria, []byte(input4))
	s.Nil(r4.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Venice", currentState.Units[1].Position)
	s.Equal("Tyrolia", currentState.Units[14].Retreating)
	s.Nil(result)

}

func (s *MyApplicationSuite) TestSupportHoldSuccess() {
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Tyrolia", "FromRegion": "Vienna"}}`

	r1 := s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Tyrolia", currentState.Units[1].Position)
	s.Nil(result)

	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Tyrolia"}}`
	input3 := `{"kind": "MoveArmy", "payload" : {"UnitID": 13, "OrderType": "support hold", "OrderOwner": "Italy", "ToRegion": "Venice", "FromRegion": "Rome"}}`
	input4 := `{"kind": "MoveArmy", "payload" : {"UnitID": 3, "OrderType": "support move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Tyrolia"}}`

	r2 := s.tester.Advance(Austria, []byte(input2))
	s.Nil(r2.Err)
	r3 := s.tester.Advance(Italy, []byte(input3))
	s.Nil(r3.Err)
	r4 := s.tester.Advance(Austria, []byte(input4))
	s.Nil(r4.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Tyrolia", currentState.Units[1].Position)
	s.Equal("Trieste", currentState.Units[3].Position)
	s.Equal("Rome", currentState.Units[13].Position)
	s.Equal("Venice", currentState.Units[14].Position)
	s.Equal("", currentState.Units[1].Retreating)
	s.Equal("", currentState.Units[3].Retreating)
	s.Equal("", currentState.Units[13].Retreating)
	s.Equal("", currentState.Units[14].Retreating)
	s.Nil(result)

}

func (s *MyApplicationSuite) TestMoveToPositionWithLeavingUnit() {
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Budapest", "FromRegion": "Vienna"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 2, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Serbia", "FromRegion": "Budapest"}}`

	r1 := s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)
	r2 := s.tester.Advance(Austria, []byte(input2))
	s.Nil(r2.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Budapest", currentState.Units[1].Position)
	s.Equal("Serbia", currentState.Units[2].Position)

	s.Nil(result)
}

func (s *MyApplicationSuite) TestMultipleSimultaniousAttacks() {
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Budapest", "FromRegion": "Vienna"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 2, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Serbia", "FromRegion": "Budapest"}}`
	input3 := `{"kind": "MoveArmy", "payload" : {"UnitID": 16, "OrderType": "move", "OrderOwner": "Russia", "ToRegion": "Ukraine", "FromRegion": "Moscow"}}`
	input4 := `{"kind": "MoveArmy", "payload" : {"UnitID": 18, "OrderType": "move", "OrderOwner": "Russia", "ToRegion": "Galicia", "FromRegion": "Warsaw"}}`
	input5 := `{"kind": "MoveArmy", "payload" : {"UnitID": 20, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Bulgaria", "FromRegion": "Constantinople"}}`

	r1 := s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)
	r2 := s.tester.Advance(Austria, []byte(input2))
	s.Nil(r2.Err)
	r3 := s.tester.Advance(Russia, []byte(input3))
	s.Nil(r3.Err)
	r4 := s.tester.Advance(Russia, []byte(input4))
	s.Nil(r4.Err)
	r5 := s.tester.Advance(Turkey, []byte(input5))
	s.Nil(r5.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Budapest", currentState.Units[1].Position)
	s.Equal("Serbia", currentState.Units[2].Position)
	s.Equal("Ukraine", currentState.Units[16].Position)
	s.Equal("Galicia", currentState.Units[18].Position)
	s.Equal("Bulgaria", currentState.Units[20].Position)
	s.Nil(result)

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Rumania", "FromRegion": "Budapest"}}`
	input2 = `{"kind": "MoveArmy", "payload" : {"UnitID": 2, "OrderType": "support move", "OrderOwner": "Austria", "ToRegion": "Rumania", "FromRegion": "Budapest"}}`
	input3 = `{"kind": "MoveArmy", "payload" : {"UnitID": 16, "OrderType": "support move", "OrderOwner": "Russia", "ToRegion": "Rumania", "FromRegion": "Galicia"}}`
	input4 = `{"kind": "MoveArmy", "payload" : {"UnitID": 18, "OrderType": "move", "OrderOwner": "Russia", "ToRegion": "Rumania", "FromRegion": "Galicia"}}`
	input5 = `{"kind": "MoveArmy", "payload" : {"UnitID": 19, "OrderType": "support move", "OrderOwner": "Russia", "ToRegion": "Rumania", "FromRegion": "Galicia"}}`
	input6 := `{"kind": "MoveArmy", "payload" : {"UnitID": 20, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Rumania", "FromRegion": "Bulgaria"}}`

	r1 = s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)
	r2 = s.tester.Advance(Austria, []byte(input2))
	s.Nil(r2.Err)
	r3 = s.tester.Advance(Russia, []byte(input3))
	s.Nil(r3.Err)
	r4 = s.tester.Advance(Russia, []byte(input4))
	s.Nil(r4.Err)
	r5 = s.tester.Advance(Russia, []byte(input5))
	s.Nil(r5.Err)
	r6 := s.tester.Advance(Turkey, []byte(input6))
	s.Nil(r6.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Budapest", currentState.Units[1].Position)
	s.Equal("Serbia", currentState.Units[2].Position)
	s.Equal("Ukraine", currentState.Units[16].Position)
	s.Equal("Rumania", currentState.Units[18].Position)
	s.Equal("Sevastopol", currentState.Units[19].Position)
	s.Equal("Bulgaria", currentState.Units[20].Position)
	s.Equal(true, currentState.Board["Rumania"].Occupied)
	s.Nil(result)

}

func (s *MyApplicationSuite) TestConvoy() {

	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "convoy", "OrderOwner": "England", "ToRegion": "Holland", "FromRegion": "Liverpool"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Holland", "FromRegion": "Liverpool"}}`

	r1 := s.tester.Advance(England, []byte(input1))
	s.ErrorContains(r1.Err, "cant convoy if the unit is not at sea")
	r2 := s.tester.Advance(England, []byte(input2))
	s.ErrorContains(r2.Err, "no available boats to convoy")

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Liverpool", currentState.Units[5].Position)
	s.Equal("London", currentState.Units[4].Position)
	s.Nil(result)

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "move", "OrderOwner": "England", "ToRegion": "North Sea", "FromRegion": "London"}}`

	r1 = s.tester.Advance(England, []byte(input1))
	s.Nil(r1.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Liverpool", currentState.Units[5].Position)
	s.Equal("North Sea", currentState.Units[4].Position)
	s.Nil(result)

	s.PassTurn()

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Liverpool"}}`
	r1 = s.tester.Advance(England, []byte(input1))
	s.ErrorContains(r1.Err, "no available boats to convoy")

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "move", "OrderOwner": "England", "ToRegion": "Yorkshire", "FromRegion": "Liverpool"}}`
	r1 = s.tester.Advance(England, []byte(input1))
	s.Nil(r1.Err)
	s.PassTurn()

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Sweden", "FromRegion": "Yorkshire"}}`
	r1 = s.tester.Advance(England, []byte(input1))
	s.ErrorContains(r1.Err, "cant convoy to a coast more than one sea tile away")

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`
	r1 = s.tester.Advance(England, []byte(input1))
	s.Nil(r1.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Yorkshire", currentState.Units[5].Position)
	s.Nil(result)

	s.PassTurn()

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "convoy", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`
	input2 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`

	r1 = s.tester.Advance(England, []byte(input1))
	r2 = s.tester.Advance(England, []byte(input2))
	s.Nil(r1.Err)
	s.Nil(r2.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Norway", currentState.Units[5].Position)
	s.Nil(result)
}

func (s *MyApplicationSuite) TestConvoyAttacked() {

	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "move", "OrderOwner": "England", "ToRegion": "North Sea", "FromRegion": "London"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "move", "OrderOwner": "England", "ToRegion": "Yorkshire", "FromRegion": "Liverpool"}}`
	input3 := `{"kind": "MoveArmy", "payload" : {"UnitID": 8, "OrderType": "move", "OrderOwner": "France", "ToRegion": "English Channel", "FromRegion": "Brest"}}`

	r1 := s.tester.Advance(England, []byte(input1))
	r2 := s.tester.Advance(England, []byte(input2))
	r3 := s.tester.Advance(France, []byte(input3))
	s.Nil(r1.Err)
	s.Nil(r2.Err)
	s.Nil(r3.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Yorkshire", currentState.Units[5].Position)
	s.Equal("North Sea", currentState.Units[4].Position)
	s.Equal("English Channel", currentState.Units[8].Position)
	s.Nil(result)

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`
	input2 = `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "convoy", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`
	input3 = `{"kind": "MoveArmy", "payload" : {"UnitID": 8, "OrderType": "move", "OrderOwner": "France", "ToRegion": "North Sea", "FromRegion": "English Channel"}}`

	r1 = s.tester.Advance(England, []byte(input1))
	r2 = s.tester.Advance(England, []byte(input2))
	r3 = s.tester.Advance(France, []byte(input3))
	s.Nil(r1.Err)
	s.Nil(r2.Err)
	s.Nil(r3.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Yorkshire", currentState.Units[5].Position)
	s.Equal("North Sea", currentState.Units[4].Position)
	s.Equal("English Channel", currentState.Units[8].Position)
	s.Nil(result)
}

func (s *MyApplicationSuite) TestConvoyDebug() {

	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "move", "OrderOwner": "England", "ToRegion": "North Sea", "FromRegion": "London"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "move", "OrderOwner": "England", "ToRegion": "Yorkshire", "FromRegion": "Liverpool"}}`

	r1 := s.tester.Advance(England, []byte(input1))
	r2 := s.tester.Advance(England, []byte(input2))
	s.Nil(r1.Err)
	s.Nil(r2.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Yorkshire", currentState.Units[5].Position)
	s.Equal("North Sea", currentState.Units[4].Position)
	s.Nil(result)

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 5, "OrderType": "convoy move", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`
	input2 = `{"kind": "MoveArmy", "payload" : {"UnitID": 4, "OrderType": "convoy", "OrderOwner": "England", "ToRegion": "Norway", "FromRegion": "Yorkshire"}}`
	r1 = s.tester.Advance(England, []byte(input1))
	r2 = s.tester.Advance(England, []byte(input2))
	s.Nil(r1.Err)
	s.Nil(r2.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Norway", currentState.Units[5].Position)
	s.Nil(result)
}

func (s *MyApplicationSuite) TestMovingFromSubRegions() {
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 17, "OrderType": "move", "OrderOwner": "Russia", "ToRegion": "Norway", "FromRegion": "St Petersburg", "FromSubRegion": "South Coast"}}`
	r1 := s.tester.Advance(Russia, []byte(input1))
	s.ErrorContains(r1.Err, "cant reach this region from this harbor")

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 17, "OrderType": "move", "OrderOwner": "Russia", "ToRegion": "Finland", "FromRegion": "St Petersburg", "FromSubRegion": "South Coast"}}`
	r1 = s.tester.Advance(Russia, []byte(input1))
	s.Nil(r1.Err)

}

func (s *MyApplicationSuite) TestMovingIntoSubRegions() {
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 22, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Black Sea", "FromRegion": "Ankara", "FromSubRegion": ""}}`
	r1 := s.tester.Advance(Turkey, []byte(input1))
	s.Nil(r1.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Black Sea", currentState.Units[22].Position)
	s.Nil(result)

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 22, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Bulgaria", "FromRegion": "Black Sea", "ToSubRegion": ""}}`
	r1 = s.tester.Advance(Turkey, []byte(input1))
	s.ErrorContains(r1.Err, "need to specify the sub region and can't move directly between sub regions")

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 22, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Bulgaria", "FromRegion": "Black Sea", "ToSubRegion": "South Coast"}}`
	r1 = s.tester.Advance(Turkey, []byte(input1))
	s.ErrorContains(r1.Err, "cant move to non adjacent harbor")

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 22, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Bulgaria", "FromRegion": "Black Sea", "FromSubRegion": "", "ToSubRegion": "North Coast"}}`
	r1 = s.tester.Advance(Turkey, []byte(input1))
	s.Nil(result)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Bulgaria", currentState.Units[22].Position)
	s.Equal("North Coast", currentState.Units[22].SubPosition)
	s.Nil(result)

}

func (s *MyApplicationSuite) TestRetreat() {
	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Tyrolia", "FromRegion": "Vienna"}}`

	r1 := s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)

	report, result := s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Tyrolia", currentState.Units[1].Position)
	s.Nil(result)

	input1 = `{"kind": "MoveArmy", "payload" : {"UnitID": 1, "OrderType": "move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Tyrolia"}}`
	input2 := `{"kind": "MoveArmy", "payload" : {"UnitID": 3, "OrderType": "support move", "OrderOwner": "Austria", "ToRegion": "Venice", "FromRegion": "Tyrolia"}}`

	r1 = s.tester.Advance(Austria, []byte(input1))
	s.Nil(r1.Err)
	r2 := s.tester.Advance(Austria, []byte(input2))
	s.Nil(r2.Err)

	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	s.Equal("Venice", currentState.Units[1].Position)
	s.Equal("Trieste", currentState.Units[3].Position)
	s.Equal("Venice", currentState.Units[14].Position)
	s.Equal("", currentState.Units[1].Retreating)
	s.Equal("", currentState.Units[3].Retreating)
	s.Equal("", currentState.Units[13].Retreating)
	s.Equal("Tyrolia", currentState.Units[14].Retreating)
	s.Equal("retreats", currentState.Turn)
	s.Nil(result)

	input1 = `{"kind": "Retreat", "payload" : {"UnitID": 14, "OrderType": "move", "OrderOwner": "Italy", "ToRegion": "Venice", "FromRegion": "Venice"}}`
	r1 = s.tester.Advance(Italy, []byte(input1))
	s.ErrorContains(r1.Err, "cant retreat to the same place")

	input1 = `{"kind": "Retreat", "payload" : {"UnitID": 14, "OrderType": "move", "OrderOwner": "Italy", "ToRegion": "Tyrolia", "FromRegion": "Venice"}}`
	r1 = s.tester.Advance(Italy, []byte(input1))
	s.ErrorContains(r1.Err, "cant retreat forward")

	//giving no retreat orders should delete the unit

	//report, result = s.PassTurn()
	report, result = s.PassTurn()

	json.Unmarshal([]byte(string(report)), &currentState)

	_, ok := currentState.Units[14]

	s.Equal("build", currentState.Turn)
	s.Equal(false, ok)
	s.Nil(result)

}

func (s *MyApplicationSuite) TestMoveFromFlasePosition() {

	input1 := `{"kind": "MoveArmy", "payload" : {"UnitID": 17, "OrderType": "move", "OrderOwner": "Turkey", "ToRegion": "Black Sea", "FromRegion": "Ankara", "ToSubRegion": ""}}`
	r1 := s.tester.Advance(Russia, []byte(input1))
	s.ErrorContains(r1.Err, "your army is not there")

}

func (s *MyApplicationSuite) TestInspect() {
	result := s.tester.Inspect(payload)
	s.Nil(result.Err)
}
