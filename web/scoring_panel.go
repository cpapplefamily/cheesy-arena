// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/Team254/cheesy-arena/game"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled bool
		Alliance     string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	var realtimeScore1 **field.RealtimeScore
	var realtimeScore2 **field.RealtimeScore
	if alliance == "red" {
		realtimeScore1 = &web.arena.RedRealtimeScore
		realtimeScore2 = &web.arena.BlueRealtimeScore
	} else {
		realtimeScore1 = &web.arena.BlueRealtimeScore
		realtimeScore2 = &web.arena.RedRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, _, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		score1 := &(*realtimeScore1).CurrentScore
		score2 := &(*realtimeScore2).CurrentScore
		scoreChanged := false
		if command == "`"{
			log.Print("command: ",command)
			score1.AutoGridToggle_Enabled = !score1.AutoGridToggle_Enabled
			scoreChanged = true
		}
		if command == "-"{
			log.Print("command: ",command)
			score1.EndGameChargeStationEngaged = !score1.EndGameChargeStationEngaged
			scoreChanged = true	
		}
		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if number, err := strconv.Atoi(command); err == nil && number >= 0 && number <= 9 {
			// Handle per-robot scoring fields.
			log.Print("Number: ", number)
			if number == 0 {
				score1.AutoChargeStationEngaged = !score1.AutoChargeStationEngaged
				scoreChanged = true	
			} else if number <= 3 {
				index := number - 1
				score1.TaxiStatuses[index] = !score1.TaxiStatuses[index]
				score1.MobilityStatuses[index] = !score1.MobilityStatuses[index]
				scoreChanged = true
			} else if number >= 7{ //Buttons 7-9
				index := number - 7
				score1.EndgameStatuses[index]++
				if score1.EndgameStatuses[index] == 5 {
					score1.EndgameStatuses[index] = 0
				}
				score1.ChargedUpEndgameStatuses[index]++
				if score1.ChargedUpEndgameStatuses[index] == 3 {
					score1.ChargedUpEndgameStatuses[index] = 0
				}
				scoreChanged = true
			} else { //Buttons 4-6
				index := number - 4
				score1.AutoChargeStationDockedStatuses[index] = !score1.AutoChargeStationDockedStatuses[index]
				scoreChanged = true	
			}
		} else if !web.arena.Plc.IsEnabled() {
			switch strings.ToUpper(command) {
				//Auto scoring is from a device running as Red Page
			case "RL":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score1.AutoCargoLower[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score1.TeleopCargoLower[:])
					}
				}

			case "RU":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score1.AutoCargoUpper[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score1.TeleopCargoUpper[:])
					}
				}
			case "BL":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score2.AutoCargoLower[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score2.TeleopCargoLower[:])
					}
				}
			case "BU":
				// Don't read score from counter if not in match :TODO Add TeliopPostMatch
				if web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.PreMatch {
					if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
						scoreChanged = incrementGoal(score2.AutoCargoUpper[:])
					}
					if web.arena.MatchState == field.TeleopPeriod {
						scoreChanged = incrementGoal(score2.TeleopCargoUpper[:])
					}
				}
			
			//Top Row
			case "Q":
				scoreChanged = gridscored(score1,2,0)
			case "W":
				scoreChanged = gridscored(score1,2,1)
			case "E":
				scoreChanged = gridscored(score1,2,2)
			case "R":
				scoreChanged = gridscored(score1,2,3)
			case "T":
				scoreChanged = gridscored(score1,2,4)
			case "Y":
				scoreChanged = gridscored(score1,2,5)
			case "U":
				scoreChanged = gridscored(score1,2,6)
			case "I":
				scoreChanged = gridscored(score1,2,7)
			case "O":
				scoreChanged = gridscored(score1,2,8)
			//Mid Row
			case "A":
				scoreChanged = gridscored(score1,1,0)
			case "S":
				scoreChanged = gridscored(score1,1,1)
			case "D":
				scoreChanged = gridscored(score1,1,2)
			case "F":
				scoreChanged = gridscored(score1,1,3)
			case "G":
				scoreChanged = gridscored(score1,1,4)
			case "H":
				scoreChanged = gridscored(score1,1,5)
			case "J":
				scoreChanged = gridscored(score1,1,6)
			case "K":
				scoreChanged = gridscored(score1,1,7)
			case "L":
				scoreChanged = gridscored(score1,1,8)
			//Bottom Row
			case "Z":
				scoreChanged = gridscored(score1,0,0)
			case "X":
				scoreChanged = gridscored(score1,0,1)
			case "C":
				scoreChanged = gridscored(score1,0,2)
			case "V":
				scoreChanged = gridscored(score1,0,3)
			case "B":
				scoreChanged = gridscored(score1,0,4)
			case "N":
				scoreChanged = gridscored(score1,0,5)
			case "M":
				scoreChanged = gridscored(score1,0,6)
			case ",":
				scoreChanged = gridscored(score1,0,7)
			case ".":
				scoreChanged = gridscored(score1,0,8)
			}

		}
		
		if scoreChanged {
			checkCoopStatus(score1, score2)
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

// Check Coopertition Status
func checkCoopStatus(score1 *game.Score, score2 *game.Score){
	if score1.LinksCoopertitionReady && score2.LinksCoopertitionReady {
		score1.LinksCoopertitionAchived = true
		score2.LinksCoopertitionAchived = true
	}else{
		score1.LinksCoopertitionAchived = false
		score2.LinksCoopertitionAchived = false
	}
}

// Increments the cargo count for the given goal.
func incrementGoal(goal []int) bool {
	// Use just the first hub quadrant for manual scoring.
	goal[0]++
	return true
}

// Decrements the cargo for the given goal.
func decrementGoal(goal []int) bool {
	// Use just the first hub quadrant for manual scoring.
	if goal[0] > 0 {
		goal[0]--
		return true
	}
	return false
}

func gridscored(score *game.Score, row int, column int)bool{
	if(score.AutoGridToggle_Enabled){
		score.GridAciveInAutoStatuses[row][column] = !score.GridAciveInAutoStatuses[row][column] 
		score.GridGamePeiceStatuses[row][column]  = score.GridAciveInAutoStatuses[row][column] 
	}else{
		score.GridGamePeiceStatuses[row][column]  = !score.GridGamePeiceStatuses[row][column] 
	}
	return true
}
