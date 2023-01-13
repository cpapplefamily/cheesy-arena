// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

import "log"

type Score struct {
	TaxiStatuses                    [3]bool
	AutoCargoLower                  [4]int
	AutoCargoUpper                  [4]int
	TeleopCargoLower                [4]int
	TeleopCargoUpper                [4]int
	EndgameStatuses                 [3]EndgameStatus
	MobilityStatuses                [3]bool
	AutoChargeStationDockedStatuses [3]bool
	AutoChargeStationEngaged        bool
	GridAciveInAutoStatuses         [3][9]bool
	GridStatuses                    [3][9]bool
	LinksStatuses                   [3][7]int
	Links                           int
	ChargedUpEndgameStatuses        [3]ChargedUpEndgameStatus
	EndGameChargeStationEngaged     bool
	Fouls                           []Foul
	ElimDq                          bool
}

var QuintetThreshold = 5
var CargoBonusRankingPointThresholdWithoutQuintet = 20
var CargoBonusRankingPointThresholdWithQuintet = 18
var HangarBonusRankingPointThreshold = 16
var DoubleBonusRankingPointThreshold = 0

var CoopititionThreshold = 5
var LinksRankingPointThresholdWithoutCoopertition = 5
var LinksRankingPointThresholdWithCoopertition = 4
var ChargeStationRankingPointThreshold = 26

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameLow
	EndgameMid
	EndgameHigh
	EndgameTraversal
)

type ChargedUpEndgameStatus int

const (
	Endgame_None ChargedUpEndgameStatus = iota
	Endgame_Parked
	Endgame_Docked
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, taxied := range score.TaxiStatuses {
		if taxied {
			summary.TaxiPoints += 2
		}
	}
	//*** Start Charged Up Mobility Points
	for _, mobility := range score.MobilityStatuses {
		if mobility {
			summary.MobilityPoints += 3
		}
	}
	//*** End Charged Up Mobility Points

	//*** Start Charged Up Auto Charge Station Points
	for i := 0; i < 3; i++ {
		if score.AutoChargeStationDockedStatuses[i] {
			summary.AutoChargePoints = 8
			if score.AutoChargeStationEngaged {
				summary.AutoChargePoints = summary.AutoChargePoints + 4
			}
			//once we find one robot Docked Quit Looking
			break
		}
	}
	//*** End Charged Up Auto Charge Station Points

	//*** Start Charged Up Tally AutoPoints for TieBreaker
	summary.AutoPoints += (summary.MobilityPoints + summary.AutoChargePoints)
	//*** End Charged Up Tally AutoPoints for TieBreaker

	for i := 0; i < 4; i++ {
		summary.AutoCargoCount += score.AutoCargoLower[i] + score.AutoCargoUpper[i]
		summary.AutoCargoPoints += 2 * score.AutoCargoLower[i]
		summary.AutoCargoPoints += 4 * score.AutoCargoUpper[i]
	}

	//*** Start Charged Up Apply Bouns point if the Grid was populated in auto and had a active game Piece
	for i := 0; i < 9; i++ {
		//Low Goal points
		if score.GridStatuses[0][i] {
			summary.GridPoints += 2
		}
		//Low Goal Auto Bonus
		if score.GridStatuses[0][i] && score.GridAciveInAutoStatuses[0][i] {
			summary.GridPoints += 1
			summary.AutoPoints += 3 // Auto Points only used for Tie Breaker Otherwise its in the GridPoints already
		}
		//Mid Goal points
		if score.GridStatuses[1][i] {
			summary.GridPoints += 3
		}
		//Mid Goal Auto Bonus
		if score.GridStatuses[1][i] && score.GridAciveInAutoStatuses[1][i] {
			summary.GridPoints += 1
			summary.AutoPoints += 4 // Auto Points only used for Tie Breaker Otherwise its in the GridPoints already
		}
		//Hight Goal points
		if score.GridStatuses[2][i] {
			summary.GridPoints += 5
		}
		//High Goal Auto Bonus
		if score.GridStatuses[2][i] && score.GridAciveInAutoStatuses[2][i] {
			summary.GridPoints += 1
			summary.AutoPoints += 6 // Auto Points only used for Tie Breaker Otherwise its in the GridPoints already
		}
	}
	//*** End Charged Up Apply Bouns point if the Grid was populated in auto and had a active game Piece

	// Calculate teleoperated period cargo points.
	summary.CargoCount = summary.AutoCargoCount
	summary.CargoPoints = summary.AutoCargoPoints
	for i := 0; i < 4; i++ {
		summary.CargoCount += score.TeleopCargoLower[i] + score.TeleopCargoUpper[i]
		summary.CargoPoints += 1 * score.TeleopCargoLower[i]
		summary.CargoPoints += 2 * score.TeleopCargoUpper[i]
	}

	//*** Start Charged Up Calculate Links Statuse
	//Fist step finds all overlaping groups of 3
	for i := 0; i < 3; i++ {
		for j := 0; j < 7; j++ {
			if score.GridStatuses[i][j] {
				score.LinksStatuses[i][j] = score.LinksStatuses[i][j] + 1
			}
			if score.GridStatuses[i][j+1] {
				score.LinksStatuses[i][j] = score.LinksStatuses[i][j] + 1
			}
			if score.GridStatuses[i][j+2] {
				score.LinksStatuses[i][j] = score.LinksStatuses[i][j] + 1
			}
		}
	}
	// Second step removes Overlaping
	// Column 0
	// Never Removed
	// Column 1
	// removed if Column 0 = 3
	for i := 0; i < 3; i++ {
		if score.LinksStatuses[i][0] >= 3 {
			score.LinksStatuses[i][1] = 0
		}
	}
	// Column 2-7
	// Removed if either previous 2 columns are = 3
	for i := 0; i < 3; i++ {
		for j := 2; j < 7; j++ {
			if score.LinksStatuses[i][j-1] >= 3 || score.LinksStatuses[i][j-2] >= 3 {
				score.LinksStatuses[i][j] = 0
			}
		}
	}

	//Count the remaining Links
	for i := 0; i < 3; i++ {
		for j := 2; j < 7; j++ {
			if score.LinksStatuses[i][j] >= 3 {
				summary.LinksCount += 1
			}
		}
	}

	log.Print("summary.LinksCount: ", summary.LinksCount)

	//Total Links Points
	summary.LinksPoints = summary.LinksCount * 5
	//*** End Charged Up Calculate Links Statuse

	// Calculate endgame points.
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgameLow:
			summary.HangarPoints += 4
		case EndgameMid:
			summary.HangarPoints += 6
		case EndgameHigh:
			summary.HangarPoints += 10
		case EndgameTraversal:
			summary.HangarPoints += 15
		}
	}

	//*** Start Charged Up EndGame Points
	for _, status := range score.ChargedUpEndgameStatuses {
		switch status {
		case Endgame_Parked:
			summary.Endgame_ParkedPoints += 2
		case Endgame_Docked:
			if score.EndGameChargeStationEngaged{
				summary.Endgame_DockedPoints += 10
			}else{
				summary.Endgame_DockedPoints += 6
			}
		}
	}

	summary.ChargeStationPoints = summary.AutoChargePoints +
		summary.Endgame_DockedPoints +
		summary.Endgame_EngagedPoints
	//*** End Charged Up EndGame Points

	// Calculate bonus ranking points.
	summary.CargoGoal = CargoBonusRankingPointThresholdWithoutQuintet
	// A QuintetThreshold of 0 disables the Quintet.
	if QuintetThreshold > 0 && summary.AutoCargoCount >= QuintetThreshold {
		summary.CargoGoal = CargoBonusRankingPointThresholdWithQuintet
		summary.QuintetAchieved = true
	}
	if summary.CargoCount >= summary.CargoGoal {
		summary.CargoBonusRankingPoint = true
	}
	summary.HangarBonusRankingPoint = summary.HangarPoints >= HangarBonusRankingPointThreshold

	// The "double bonus" ranking point is an offseason-only addition which grants an additional RP if either the total
	// cargo count or the hangar points is over the certain threshold. A threshold of 0 disables this RP.
	if DoubleBonusRankingPointThreshold > 0 {
		summary.DoubleBonusRankingPoint = summary.CargoCount >= DoubleBonusRankingPointThreshold ||
			summary.HangarPoints >= DoubleBonusRankingPointThreshold
	}

	//** Start Ranking Point Calculations
	//++ Start Links RP
	summary.LinksGoal = LinksRankingPointThresholdWithoutCoopertition

	var OpponentCoopAchived = false
	var MatchCoopAchived = false
	if OpponentCoopAchived && summary.LinksCount >= CoopititionThreshold {
		MatchCoopAchived = true
	}

	if CoopititionThreshold > 0 && MatchCoopAchived && summary.LinksCount >= CoopititionThreshold {
		summary.LinksGoal = LinksRankingPointThresholdWithCoopertition
		//summary.QuintetAchieved = true
	}
	if summary.LinksCount >= summary.LinksGoal {
		summary.LinksRankingPoint = true
	}
	//++ End Links RP
	//++ Start Charging Station RP
	summary.ChargeStationPoints = summary.AutoChargePoints + summary.Endgame_DockedPoints + summary.Endgame_EngagedPoints
	if summary.ChargeStationPoints >= ChargeStationRankingPointThreshold {
		summary.ChargeStationRankingPoint = true
	}
	//++ End Charging Station RP
	//** End Ranking Point Calculations

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	// Check for the opponent fouls that automatically trigger a ranking point.
	// Note: There are no such fouls in the 2022 game; leaving this comment for future years.
	log.Print("summary.MobilityPoints: ", summary.MobilityPoints)
	log.Print("summary.GridPoints: ", summary.GridPoints)
	log.Print("summary.LinksPoints: ", summary.LinksPoints)
	log.Print("summary.ChargeStationPoints: ", summary.ChargeStationPoints)
	summary.MatchPoints = summary.MobilityPoints +
		summary.GridPoints +
		summary.LinksPoints +
		summary.ChargeStationPoints +
		summary.Endgame_ParkedPoints

	summary.Score = summary.MatchPoints + summary.FoulPoints

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.TaxiStatuses != other.TaxiStatuses ||
		score.AutoCargoLower != other.AutoCargoLower ||
		score.AutoCargoUpper != other.AutoCargoUpper ||
		score.TeleopCargoLower != other.TeleopCargoLower ||
		score.TeleopCargoUpper != other.TeleopCargoUpper ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.ElimDq != other.ElimDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
