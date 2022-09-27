// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Defines the tournament structure for a double-elimination bracket culminating in a best-of-three final.

package bracket

import (
	"fmt"
	"log"
)

// Creates an unpopulated double-elimination bracket. Only supports having exactly eight alliances.
func NewDoubleEliminationBracket(numAlliances int) (*Bracket, error) {
	if (numAlliances == 8) {
		return newBracket(doubleEliminationBracketMatchupTemplates_8, newMatchupKey(6, 1), numAlliances)
	}else if (numAlliances == 4) {
		return newBracket(doubleEliminationBracketMatchupTemplates_4, newMatchupKey(6, 1), numAlliances)
	}else{
		log.Printf("NewDoubleEliminationBracket numAlliances: %s", numAlliances)
		return nil, fmt.Errorf("Must have exactly 4 or 8 alliances")	
	}
}

var doubleEliminationBracketMatchupTemplates_8 = []matchupTemplate{
	{
		matchupKey:         newMatchupKey(1, 1),
		displayName:        "1",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: allianceSource{allianceId: 8},
	},
	{
		matchupKey:         newMatchupKey(1, 2),
		displayName:        "2",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 4},
		blueAllianceSource: allianceSource{allianceId: 5},
	},
	{
		matchupKey:         newMatchupKey(1, 3),
		displayName:        "3",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 3},
		blueAllianceSource: allianceSource{allianceId: 6},
	},
	{
		matchupKey:         newMatchupKey(1, 4),
		displayName:        "4",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 2},
		blueAllianceSource: allianceSource{allianceId: 7},
	},
	{
		matchupKey:         newMatchupKey(2, 1),
		displayName:        "5",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 1),
		blueAllianceSource: newLoserAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 2),
		displayName:        "6",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 3),
		blueAllianceSource: newLoserAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(2, 3),
		displayName:        "7",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 1),
		blueAllianceSource: newWinnerAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 4),
		displayName:        "8",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 3),
		blueAllianceSource: newWinnerAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(3, 1),
		displayName:        "9",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 2),
	},
	{
		matchupKey:         newMatchupKey(3, 2),
		displayName:        "10",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 4),
		blueAllianceSource: newWinnerAllianceSource(2, 1),
	},
	{
		matchupKey:         newMatchupKey(4, 1),
		displayName:        "11",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(3, 2),
		blueAllianceSource: newWinnerAllianceSource(3, 1),
	},
	{
		matchupKey:         newMatchupKey(4, 2),
		displayName:        "12",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 4),
	},
	{
		matchupKey:         newMatchupKey(5, 1),
		displayName:        "13",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(4, 2),
		blueAllianceSource: newWinnerAllianceSource(4, 1),
	},
	{
		matchupKey:         newMatchupKey(6, 1),
		displayName:        "F",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(4, 2),
		blueAllianceSource: newWinnerAllianceSource(5, 1),
	},
}
var doubleEliminationBracketMatchupTemplates_4 = []matchupTemplate{
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(1, 1),
		displayName:        "1",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 1},
		blueAllianceSource: allianceSource{allianceId: 1},
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(1, 2),
		displayName:        "2",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 4},
		blueAllianceSource: allianceSource{allianceId: 4},
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(1, 3),
		displayName:        "3",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 2},
		blueAllianceSource: allianceSource{allianceId: 2},
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(1, 4),
		displayName:        "4",
		NumWinsToAdvance:   1,
		redAllianceSource:  allianceSource{allianceId: 3},
		blueAllianceSource: allianceSource{allianceId: 3},
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(2, 1),
		displayName:        "5",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 1),
		blueAllianceSource: newLoserAllianceSource(1, 2),
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(2, 2),
		displayName:        "6",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(1, 3),
		blueAllianceSource: newLoserAllianceSource(1, 4),
	},
	{
		matchupKey:         newMatchupKey(2, 3),
		displayName:        "M1",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 1),
		blueAllianceSource: newWinnerAllianceSource(1, 2),
	},
	{
		matchupKey:         newMatchupKey(2, 4),
		displayName:        "M2",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(1, 3),
		blueAllianceSource: newWinnerAllianceSource(1, 4),
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(3, 1),
		displayName:        "9",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 2),
	},
	{	//Fource Alliance RED to win
		matchupKey:         newMatchupKey(3, 2),
		displayName:        "10",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(2, 4),
		blueAllianceSource: newWinnerAllianceSource(2, 1),
	},
	{
		matchupKey:         newMatchupKey(4, 1),
		displayName:        "M4",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(3, 1),
		blueAllianceSource: newWinnerAllianceSource(3, 2),
	},
	{
		matchupKey:         newMatchupKey(4, 2),
		displayName:        "M3",
		NumWinsToAdvance:   1,
		redAllianceSource:  newWinnerAllianceSource(2, 3),
		blueAllianceSource: newWinnerAllianceSource(2, 4),
	},
	{
		matchupKey:         newMatchupKey(5, 1),
		displayName:        "M5",
		NumWinsToAdvance:   1,
		redAllianceSource:  newLoserAllianceSource(4, 2),
		blueAllianceSource: newWinnerAllianceSource(4, 1),
	},
	{
		matchupKey:         newMatchupKey(6, 1),
		displayName:        "F",
		NumWinsToAdvance:   2,
		redAllianceSource:  newWinnerAllianceSource(4, 2),
		blueAllianceSource: newWinnerAllianceSource(5, 1),
	},
}