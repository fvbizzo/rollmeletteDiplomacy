package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gligneul/rollmelette"
	"github.com/stretchr/testify/suite"
)

var payload = common.Hex2Bytes("deadbeef")
var msgSender = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafafa")
var Austria = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf1")
var England = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf2")
var France = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf3")
var Germany = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf4")
var Italy = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf5")
var Russia = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf6")
var Turkey = common.HexToAddress("0xfafafafafafafafafafafafafafafafafafafaf7")

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

// Testing the build army function
func (s *MyApplicationSuite) TestBuildArmy() {
	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": false}}`
	result := s.tester.Advance(England, []byte(input))
	s.Nil(result.Err)
}

// Testing a player trying to build an army outside build phase
func (s *MyApplicationSuite) TestBuildArmyOutsideBuildPhase() {
	input1 := `{"kind": "PassTurn", "payload": ""}`
	s.tester.Advance(England, []byte(input1))
	input := `{"kind": "BuildArmy", "payload" : {"Type": "army", "Position": "London", "Owner": "England", "Delete": false}}`
	result := s.tester.Advance(England, []byte(input))
	s.ErrorContains(result.Err, "cant build an army outside build")

}

func (s *MyApplicationSuite) TestInspect() {
	result := s.tester.Inspect(payload)
	s.Nil(result.Err)
}
