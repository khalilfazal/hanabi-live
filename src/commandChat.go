/*
	Sent when the user sends a text message to the lobby by pressing enter
	"data" example:
	{
		msg: 'hi',
		room: 'lobby',
		// Room can also be 'game'
	}
*/

package main

import (
	"strconv"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const (
	maxChatLength = 300
)

func commandChat(s *Session, d *CommandData) {
	// Local variables
	var userID int
	if d.Discord || d.Server {
		userID = 0
	} else {
		if s == nil {
			log.Error("Failed to send a chat message because the sender's session was nil.")
			return
		}
		userID = s.UserID()
	}
	if d.Username == "" && s != nil {
		d.Username = s.Username()
	}

	/*
		Validate
	*/

	// Validate the message
	if d.Msg == "" {
		s.Warning("You cannot send a blank message.")
		return
	}

	// Truncate long messages
	if len(d.Msg) > maxChatLength {
		d.Msg = d.Msg[0 : maxChatLength-1]
	}

	// Validate the room
	if d.Room != "lobby" && d.Room != "game" {
		s.Warning("That is not a valid room.")
		return
	}

	// Sanitize the message using the bluemonday library to stop
	// various attacks against other players
	rawMsg := d.Msg
	p := bluemonday.StrictPolicy()
	d.Msg = p.Sanitize(d.Msg)

	/*
		Chat
	*/

	// Log the message
	text := "<" + d.Username
	if d.DiscordDiscriminator != "" {
		text += "#" + d.DiscordDiscriminator
	}
	text += "> " + d.Msg
	log.Info(text)

	// Handle in-game chat in a different function; the rest of this function will be for lobby chat
	if d.Room == "game" {
		commandChatGame(s, d)
		return
	}

	// Add the message to the database
	if d.Discord {
		if err := db.ChatLog.InsertDiscord(d.Username, d.Msg, d.Room); err != nil {
			log.Error("Failed to insert a Discord chat message into the database:", err)
			s.Error("")
			return
		}
	} else {
		if err := db.ChatLog.Insert(userID, d.Msg, d.Room); err != nil {
			log.Error("Failed to insert a chat message into the database:", err)
			s.Error("")
			return
		}
	}

	// Convert Discord mentions from number to username
	d.Msg = chatFillMentions(d.Msg)

	// Convert Discord channel names from number to username
	d.Msg = chatFillChannels(d.Msg)

	// Lobby messages go to everyone
	for _, s2 := range sessions {
		s2.NotifyChat(d.Msg, d.Username, d.Discord, d.Server, time.Now(), d.Room)
	}

	// Send the chat message to the Discord "#general" channel if we are replicating a message
	to := discordLobbyChannel
	if d.Server && !d.Echo {
		// Send server messages to a separate channel
		to = discordBotChannel
	}

	// Don't send Discord messages that we are already replicating
	if !d.Discord {
		// We use "rawMsg" instead of "d.Msg" because we want to send the unsanitized message
		// The bluemonday library is intended for HTML rendering, and Discord can handle any special characters
		discordSend(to, d.Username, rawMsg)
	}

	// Check for commands
	chatCommand(s, d)
}

func commandChatGame(s *Session, d *CommandData) {
	gameID := s.CurrentGame()

	// Validate that they are in a game if they are trying to send to a game room
	if gameID == -1 {
		s.Warning("You cannot send game chat if you are not in a game.")
		return
	}

	// Get the corresponding game
	var g *Game
	if v, ok := games[gameID]; !ok {
		s.Warning("Game " + strconv.Itoa(gameID) + " does not exist.")
		return
	} else {
		g = v
	}

	// Validate that this player is in the game or spectating
	if g.GetPlayerIndex(s.UserID()) == -1 && g.GetSpectatorIndex(s.UserID()) == -1 {
		s.Warning("You are not playing or spectating game " + strconv.Itoa(gameID) + ".")
		return
	}

	// Store the chat in memory for later
	chatMsg := &GameChatMessage{
		UserID:       s.UserID(),
		Username:     d.Username, // This was prepared above in the "commandChat()" function
		Msg:          d.Msg,
		DatetimeSent: time.Now(),
	}
	g.Chat = append(g.Chat, chatMsg)

	// Send it to all of the players and spectators
	if !g.SharedReplay {
		for _, p := range g.Players {
			p.Session.NotifyChat(d.Msg, d.Username, d.Discord, d.Server, chatMsg.DatetimeSent, d.Room)
		}
	}
	for _, s2 := range g.Spectators {
		s2.NotifyChat(d.Msg, d.Username, d.Discord, d.Server, chatMsg.DatetimeSent, d.Room)
	}
}
