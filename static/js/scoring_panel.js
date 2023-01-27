// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the scoring interface.

var websocket;
var alliance;

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  $("#matchName").text(data.MatchType + " " + data.Match.DisplayName);
  if (alliance === "red") {
    $("#team1").text(data.Match.Red1);
    $("#team2").text(data.Match.Red2);
    $("#team3").text(data.Match.Red3);
  } else {
    $("#team1").text(data.Match.Blue1);
    $("#team2").text(data.Match.Blue2);
    $("#team3").text(data.Match.Blue3);
  }
};

// Handles a websocket message to update the match status.
var handleMatchTime = function(data) {
  switch (matchStates[data.MatchState]) {
    case "PRE_MATCH":
      // Pre-match message state is set in handleRealtimeScore().
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
      break;
    case "POST_MATCH":
      $("#postMatchMessage").hide();
      $("#commitMatchScore").css("display", "flex");
      break;
    default:
      $("#postMatchMessage").hide();
      $("#commitMatchScore").hide();
  }
};

// Handles a websocket message to update the realtime scoring fields.
var handleRealtimeScore = function(data) {
  var realtimeScore1;
  var realtimeScore2;
  if (alliance === "red") {
    realtimeScore1 = data.Red;
    realtimeScore2 = data.Blue;
  } else {
    realtimeScore1 = data.Blue;
    realtimeScore2 = data.Red;
  }
  var score1 = realtimeScore1.Score;
  var score2 = realtimeScore2.Score;

  $("#autoGridToggleEnabled>.value").text(score1.AutoGridToggle_Enabled ? "Auto Score Toggel Enabled" : "Auto Score Toggel Disabled");
  $("#autoGridToggleEnabled").attr("data-value", score1.AutoGridToggle_Enabled);
  //For Debuging Scoring
  $("#currentScore").text("Current Score: " + realtimeScore1.ScoreSummary.Score);

  //Group One Score
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    $("#taxiStatus" + i1 + ">.value").text(score1.TaxiStatuses[i] ? "Yes" : "No");
    $("#taxiStatus" + i1).attr("data-value", score1.TaxiStatuses[i]);
    $("#mobilityStatus" + i1 + ">.value").text(score1.MobilityStatuses[i] ? "Yes" : "No");
    $("#mobilityStatus" + i1).attr("data-value", score1.MobilityStatuses[i]);
    $("#endgameStatus" + i1 + ">.value").text(getEndgameStatusText(score1.EndgameStatuses[i]));
    $("#endgameStatus" + i1).attr("data-value", score1.EndgameStatuses[i]);
    $("#chargedUpEndgameStatus" + i1 + ">.value").text(getChargedUpEndgameStatusText(score1.ChargedUpEndgameStatuses[i]));
    $("#chargedUpEndgameStatus" + i1).attr("data-value", score1.ChargedUpEndgameStatuses[i]);
    $("#autoDockedStatus" + i1 + ">.value").text(score1.AutoChargeStationDockedStatuses[i] ? "Yes" : "No");
    $("#autoDockedStatus" + i1).attr("data-value", score1.AutoChargeStationDockedStatuses[i]);
    $("#autoCargoLower").text(score1.AutoCargoLower[0]);
    $("#autoCargoUpper").text(score1.AutoCargoUpper[0]);
    $("#teleopCargoLower").text(score1.TeleopCargoLower[0]);
    $("#teleopCargoUpper").text(score1.TeleopCargoUpper[0]);
  }
  
  $("#autoEngagedStatus>.value").text(score1.AutoChargeStationEngaged ? "Yes" : "No");
  $("#autoEngagedStatus").attr("data-value", score1.AutoChargeStationEngaged);
  $("#endGameChargeStationEngaged>.value").text(score1.EndGameChargeStationEngaged ? "Yes" : "No");
  $("#endGameChargeStationEngaged").attr("data-value", score1.EndGameChargeStationEngaged);
  

  for (var i = 0; i < 3; i++){
    for (var j = 0; j < 9; j++){
      $("#autoGridStatusR" + i + "C" + j + ">.value").text(score1.GridAciveInAutoStatuses[i][j] ? "Auto" : "--");
      $("#autoGridStatusR" + i + "C" + j).attr("data-value", score1.GridAciveInAutoStatuses[i][j]);
      $("#scoreR" + i + "C" + j).attr("data-value", score1.GridGamePeiceStatuses[i][j]);
    }
  }
  
  //Group Two Score 
  for (var i = 0; i < 3; i++) {
    var i1 = i + 1;
    $("#taxiStatus2" + i1 + ">.value").text(score2.TaxiStatuses[i] ? "Yes" : "No");
    $("#taxiStatus2" + i1).attr("data-value", score2.TaxiStatuses[i]);
    $("#mobilityStatus2" + i1 + ">.value").text(score2.MobilityStatuses[i] ? "Yes" : "No");
    $("#mobilityStatus2" + i1).attr("data-value", score2.MobilityStatuses[i]);
    $("#endgameStatus2" + i1 + ">.value").text(getEndgameStatusText(score2.EndgameStatuses[i]));
    $("#endgameStatus2" + i1).attr("data-value", score2.EndgameStatuses[i]);
    $("#chargedUpEndgameStatus2" + i1 + ">.value").text(getChargedUpEndgameStatusText(score2.ChargedUpEndgameStatuses[i]));
    $("#chargedUpEndgameStatus2" + i1).attr("data-value", score2.ChargedUpEndgameStatuses[i]);
    $("#autoDockedStatus2" + i1 + ">.value").text(score2.AutoChargeStationDockedStatuses[i] ? "Yes" : "No");
    $("#autoDockedStatus2" + i1).attr("data-value", score2.AutoChargeStationDockedStatuses[i]);
    $("#autoCargoLower2").text(score2.AutoCargoLower[0]);
    $("#autoCargoUpper2").text(score2.AutoCargoUpper[0]);
    $("#teleopCargoLower2").text(score2.TeleopCargoLower[0]);
    $("#teleopCargoUpper2").text(score2.TeleopCargoUpper[0]);
    $("#currentScore2").text(realtimeScore2.ScoreSummary.Score);
  }
};

// Handles a keyboard event and sends the appropriate websocket message.
var handleKeyPress = function(event) {
  websocket.send(String.fromCharCode(event.keyCode));
};

// Handles an element click and sends the appropriate websocket message.
var handleClick = function(shortcut) {
  console.log(shortcut)
  websocket.send(shortcut);
};

// Sends a websocket message to indicate that the score for this alliance is ready.
var commitMatchScore = function() {
  websocket.send("commitMatch");
  $("#postMatchMessage").css("display", "flex");
  $("#commitMatchScore").hide();
};

// Returns the display text corresponding to the given integer endgame status value.
var getEndgameStatusText = function(level) {
  switch (level) {
    case 1:
      return "Low";
    case 2:
      return "Mid";
    case 3:
      return "High";
    case 4:
      return "Traversal";
    default:
      return "None";
  }
};
var getChargedUpEndgameStatusText = function(level) {
  //console.log(level);
  switch (level) {
    case 1:
      return "Parked";
    case 2:
      return "Docked";
    default:
      return "None";
  }
};

$(function() {
  alliance = window.location.href.split("/").slice(-1)[0];
  $("#alliance").attr("data-alliance", alliance);

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/scoring/" + alliance + "/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data); },
    matchTime: function(event) { handleMatchTime(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
  });

  $(document).keypress(handleKeyPress);
});
