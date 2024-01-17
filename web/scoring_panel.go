// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
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
		PlcIsEnabled        bool
		Alliance            string
		ValidGridNodeStates map[game.Row]map[int]map[game.NodeState]string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance, game.ValidGridNodeStates()}
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

	var realtimeScore **field.RealtimeScore
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
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
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}
		score := &(*realtimeScore).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else {
			args := struct {
				TeamPosition int
				GridRow      int
				GridNode     int
				NodeState    game.NodeState
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			switch strings.ToUpper(command) {
			//case "mobilityStatus":
			case "MOBILITYSTATUS":
				if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
					score.MobilityStatuses[args.TeamPosition-1] = !score.MobilityStatuses[args.TeamPosition-1]
					scoreChanged = true
				}
			//case "autoDockStatus":
			case "AUTODOCKSTATUS":
				if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
					score.AutoDockStatuses[args.TeamPosition-1] = !score.AutoDockStatuses[args.TeamPosition-1]
					scoreChanged = true
				}
			//case "endgameStatus":
			case "ENDGAMESTATUS":
				if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
					score.EndgameStatuses[args.TeamPosition-1]++
					if score.EndgameStatuses[args.TeamPosition-1] > 2 {
						score.EndgameStatuses[args.TeamPosition-1] = 0
					}
					scoreChanged = true
				}
			//case "stageStatus":
			case "STAGESTATUS":
				if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
					score.StageStatuses[args.TeamPosition-1]++
					if score.StageStatuses[args.TeamPosition-1] > 3 {
						score.StageStatuses[args.TeamPosition-1] = 0
					}
					scoreChanged = true
				}
			//case "harmonyStatus":
			case "HARMONYSTATUS":
				if args.TeamPosition >= 1 && args.TeamPosition <= 3 {
					score.HarmonyStatuses[args.TeamPosition-1] = !score.HarmonyStatuses[args.TeamPosition-1]
					scoreChanged = true
				}
			//case "coopertitionStatus":
			case "COOPERTITIONSTATUS":
				if score.AmplificationCount > 0 && !score.CoopertitionStatus {
					score.CoopertitionStatus = !score.CoopertitionStatus
					score.AmplificationCount = decrementAmplification(score.AmplificationCount)
					scoreChanged = true
				}
			//case "amplificationActive":
			case "AMPLIFICATIONACTIVE":
				if score.AmplificationCount > 1 {//&& !score.AmplificationActive {
					score.AmplificationActive = !score.AmplificationActive
					score.AmplificationCount = 0
					scoreChanged = true
				}
			//Notes Auto
			case "Q":
				score.AutoAmpNotes--
				score.AmplificationCount = decrementAmplification(score.AmplificationCount)
				if score.AutoAmpNotes <= 0 {
					score.AutoAmpNotes = 0
				}
				scoreChanged = true
			case "W":
				score.AutoAmpNotes++
				if !score.AmpAccumulatorDisable{
					score.AmplificationCount = incrementAmplification(score.AmplificationCount)
				}
				scoreChanged = true
			case "A":
				score.AutoSpeakerNotes--
				if score.AutoSpeakerNotes <= 0 {
					score.AutoSpeakerNotes = 0
				}
				scoreChanged = true
			case "S":
				score.AutoSpeakerNotes++
				scoreChanged = true
			//Notes Teleop
			case "E":
				score.TeleopAmpNotes--
				score.AmplificationCount = decrementAmplification(score.AmplificationCount)
				if score.TeleopAmpNotes <= 0 {
					score.TeleopAmpNotes = 0
				}
				scoreChanged = true
			case "R":
				score.TeleopAmpNotes++
				if !score.AmpAccumulatorDisable{
					score.AmplificationCount = incrementAmplification(score.AmplificationCount)
				}
				scoreChanged = true
			case "D":
				score.TeleopSpeakerNotesNotAmplified--
				if score.TeleopSpeakerNotesNotAmplified <= 0 {
					score.TeleopSpeakerNotesNotAmplified = 0
				}
				scoreChanged = true
			case "F":
				if !score.AmplificationActive{
				score.TeleopSpeakerNotesNotAmplified++
				}else{
				score.TeleopSpeakerNotesAmplified++	
				score.TeleopSpeaderNotesAmplifiedLimitCount++
				}
				scoreChanged = true
			case "G":
				score.TeleopSpeakerNotesAmplified--
				if score.TeleopSpeakerNotesAmplified <= 0 {
					score.TeleopSpeakerNotesAmplified = 0
				}
				scoreChanged = true
			case "H":
				if score.AmplificationActive{
					score.TeleopSpeakerNotesAmplified++
					score.TeleopSpeaderNotesAmplifiedLimitCount++
					scoreChanged = true
				}
			case "Y":
				score.TrapNotes--
				if score.TrapNotes <= 0 {
					score.TrapNotes = 0
				}
				scoreChanged = true
			case "T":
				score.TrapNotes++
				if score.TrapNotes >= 3 {
					score.TrapNotes = 3
				}
				scoreChanged = true
			//2023
			case "autoChargeStationLevel":
				score.AutoChargeStationLevel = !score.AutoChargeStationLevel
				scoreChanged = true
			case "endgameChargeStationLevel":
				score.EndgameChargeStationLevel = !score.EndgameChargeStationLevel
				scoreChanged = true
			case "gridAutoScoring":
				if args.GridRow >= 0 && args.GridRow <= 2 && args.GridNode >= 0 && args.GridNode <= 8 {
					score.Grid.AutoScoring[args.GridRow][args.GridNode] =
						!score.Grid.AutoScoring[args.GridRow][args.GridNode]
					scoreChanged = true
				}
			case "gridNode":
				if args.GridRow >= 0 && args.GridRow <= 2 && args.GridNode >= 0 && args.GridNode <= 8 {
					currentState := score.Grid.Nodes[args.GridRow][args.GridNode]
					if currentState == args.NodeState {
						score.Grid.Nodes[args.GridRow][args.GridNode] = game.Empty
						if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
							score.Grid.AutoScoring[args.GridRow][args.GridNode] = false
						}
					} else {
						score.Grid.Nodes[args.GridRow][args.GridNode] = args.NodeState
						if web.arena.MatchState == field.AutoPeriod || web.arena.MatchState == field.PausePeriod {
							score.Grid.AutoScoring[args.GridRow][args.GridNode] = true
						}
					}
					scoreChanged = true
				}
			}

			if scoreChanged {
				web.arena.RealtimeScoreNotifier.Notify()
			}
		}

		
	}

}

func incrementAmplification(amplificationCount int) int {
	if amplificationCount < 2{
		amplificationCount++
	}
	return amplificationCount
}

func decrementAmplification(amplificationCount int) int {
	if amplificationCount > 0{
		amplificationCount--
	}
	return amplificationCount
} 
