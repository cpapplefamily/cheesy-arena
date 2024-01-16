// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	MobilityStatuses          [3]bool
	AmplificationCount			int
	AutoAmpNotes				int
	TeleopAmpNotes				int
	AutoSpeakerNotes			int
	TeleopSpeakerNotesNotAmplified	int
	TeleopSpeakerNotesAmplified	int
	TeleopSpeaderNotesAmplifiedLimitCount int
	TrapNotes					int
	HarmonyStatuses				[3]bool
	StageStatuses				[3]StageStatus
	AmplificationActive			bool
	CoopertitionStatus			bool
	///
	Grid                      Grid
	AutoDockStatuses          [3]bool
	AutoChargeStationLevel    bool
	EndgameStatuses           [3]EndgameStatus
	EndgameChargeStationLevel bool
	Fouls                     []Foul
	PlayoffDq                 bool
}

var SustainabilityBonusLinkThresholdWithoutCoop = 6
var SustainabilityBonusLinkThresholdWithCoop = 5
var ActivationBonusPointThreshold = 26

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameParked
	EndgameDocked
)

type StageStatus int

const (
	EndStageNone StageStatus = iota
	EndStageParked
	EndStageOnstageNotSpotlit
	EndStageOnstageSpotlit
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, mobility := range score.MobilityStatuses {
		if mobility {
			//summary.MobilityPoints += 3
			summary.MobilityPoints += 2
		}
	}
	autoGridPoints := score.Grid.AutoGamePiecePoints()
	autoChargeStationPoints := 0
	for i := 0; i < 3; i++ {
		if score.AutoDockStatuses[i] {
			autoChargeStationPoints += 8
			if score.AutoChargeStationLevel {
				autoChargeStationPoints += 4
			}
			break
		}
	}
	autoAmpPoints := score.AutoAmpNotes * 2
	autoSpeakerPoints := score.AutoSpeakerNotes * 5
	//summary.AutoPoints = summary.MobilityPoints + autoGridPoints + autoChargeStationPoints
	summary.AutoPoints = summary.MobilityPoints + autoAmpPoints + autoSpeakerPoints

	// Calculate teleoperated period points.
	teleopGridPoints := score.Grid.TeleopGamePiecePoints() + score.Grid.LinkPoints() + score.Grid.SuperchargedPoints()
	teleopChargeStationPoints := 0
	for i := 0; i < 3; i++ {
		switch score.EndgameStatuses[i] {
		case EndgameParked:
			summary.ParkPoints += 2
		case EndgameDocked:
			teleopChargeStationPoints += 6
			if score.EndgameChargeStationLevel {
				teleopChargeStationPoints += 4
			}
		}
	}

	summary.GridPoints = autoGridPoints + teleopGridPoints
	summary.ChargeStationPoints = autoChargeStationPoints + teleopChargeStationPoints
	summary.EndgamePoints = teleopChargeStationPoints + summary.ParkPoints
	//summary.MatchPoints = summary.MobilityPoints + 
	//						summary.GridPoints + 
	//						summary.ChargeStationPoints + 
	//						summary.ParkPoints
	
	summary.AmpPoints = score.AutoAmpNotes * 2 + 
						score.TeleopAmpNotes * 1
	summary.SpeakerPoints = score.AutoSpeakerNotes * 5 + 
							score.TeleopSpeakerNotesNotAmplified * 2 + 
							score.TeleopSpeakerNotesAmplified * 5

	summary.TrapPoints = score.TrapNotes * 5

	//summary.OnstagePoints = 0
	for i := 0; i < 3; i++ {
		switch score.StageStatuses[i] {
		case EndStageParked:
			summary.ParkPoints += 1
		case EndStageOnstageNotSpotlit:
			summary.OnstagePoints += 3
			summary.RobotsOnstage += 1
		case EndStageOnstageSpotlit:
			summary.RobotsOnstage += 1
			summary.OnstagePoints += 4
		}
	}

	harmonyCount := 0
	for i := 0; i < 3; i++ {
		if score.HarmonyStatuses[i]{
			harmonyCount += 1
		}
	}
	if harmonyCount >= 2 {
		for i := 0; i < 3; i++ {
			if score.HarmonyStatuses[i] && (score.StageStatuses[i]>1){
				summary.HarmonyPoints += 2
			}
		}
	}
	
	summary.EndStagePoints = 	summary.ParkPoints +
							summary.OnstagePoints + 
							summary.HarmonyPoints +
							summary.TrapPoints

	summary.MatchPoints = 	summary.MobilityPoints + 
							summary.AmpPoints + 
							summary.SpeakerPoints +
							summary.EndStagePoints
							


	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of tech fouls since it is used to break ties in playoffs.
		if foul.IsTechnical {
			summary.NumOpponentTechFouls++
		}

		rule := foul.Rule()
		if rule != nil {
			// Check for the opponent fouls that automatically trigger a ranking point.
			if rule.IsRankingPoint {
				summary.SustainabilityBonusRankingPoint = true
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	totalNotes := 	score.AutoAmpNotes +
					score.TeleopAmpNotes +
					score.AutoSpeakerNotes +
					score.TeleopSpeakerNotesNotAmplified +
					score.TeleopSpeakerNotesAmplified
	
	//Set Melody Threshold for Melody Ranking Point
	melodyThreshold := 18
	if score.CoopertitionStatus && opponentScore.CoopertitionStatus{
		melodyThreshold = 15
	}

	if totalNotes >= melodyThreshold{
		summary.MelodyRankingPoint = true
	}else{
		summary.MelodyRankingPoint = false
	}

	//Emsemble Ranking Point
	if (summary.EndStagePoints >= 10) && (summary.RobotsOnstage >= 2) {
		summary.EmsembleRankingPoint = true
	}else{
		summary.EmsembleRankingPoint = false
	}

	// Calculate bonus ranking points.
/* 	summary.CoopertitionBonus = score.Grid.IsCoopertitionThresholdAchieved() &&
		opponentScore.Grid.IsCoopertitionThresholdAchieved()
	summary.NumLinks = len(score.Grid.Links())
	summary.NumLinksGoal = SustainabilityBonusLinkThresholdWithoutCoop */
	// A SustainabilityBonusLinkThresholdWithCoop of 0 disables the coopertition bonus.
/* 	if SustainabilityBonusLinkThresholdWithCoop > 0 && summary.CoopertitionBonus {
		summary.NumLinksGoal = SustainabilityBonusLinkThresholdWithCoop
	}
	if summary.NumLinks >= summary.NumLinksGoal {
		summary.SustainabilityBonusRankingPoint = true
	}
	summary.ActivationBonusRankingPoint = summary.ChargeStationPoints >= ActivationBonusPointThreshold

	if summary.SustainabilityBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.ActivationBonusRankingPoint {
		summary.BonusRankingPoints++
	} */

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.MobilityStatuses != other.MobilityStatuses ||
		score.Grid != other.Grid ||
		score.AutoDockStatuses != other.AutoDockStatuses ||
		score.AutoChargeStationLevel != other.AutoChargeStationLevel ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.EndgameChargeStationLevel != other.EndgameChargeStationLevel ||
		score.PlayoffDq != other.PlayoffDq ||
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
