package websocket

import (
	"fmt"
	"math/rand"
	"strings"
)

const MAX_PLAYERS = 5
const NO_SHOT_DELIMITER = "@#@@$@#@"
const OP_DELIMITER = "~+~...$"
const WINNER_DELIMITER = "(k(*3@#"

type Pool struct {
	Host        *Client
	Register    chan *Client
	Unregister  chan *Client
	Clients     map[*Client]bool
	Broadcast   chan Message
	GameStarted bool
	Judge       int
	Turn        int
	Players     []string
	Table       []Option
}

func NewPool() *Pool {
	return &Pool{
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[*Client]bool),
		Broadcast:   make(chan Message),
		GameStarted: false,
	}
}

//this should be separated
type GameResponse struct {
	GameStarted   bool
	Players       []string
	TurnAndAction playerAndActionRequired
	Judge         int
	Type          int
	Body          string
	ID            string
	Host          string
	MyOpCards     []string
	MyNoShotCards []string
	CardsPlayed   []Option
	Winner        string
}

func (pool *Pool) getClientFromName(name string) *Client {
	for client := range pool.Clients {
		if client.ID == name {
			return client
		}
	}
	return nil
}
func (pool *Pool) removeUserFromPool(user string) {
	var i int
	for idx, player := range pool.Players {
		if player == user {
			i = idx
		}
	}
	copy(pool.Players[i:], pool.Players[i+1:])
	pool.Players[len(pool.Players)-1] = ""
	pool.Players = pool.Players[:len(pool.Players)-1]
}
func (pool *Pool) Start() {
	for {
		select {
		case user := <-pool.Register:
			pool.Clients[user] = true
			//Right now host is purely to start the game.
			if len(pool.Clients) == 1 {
				pool.Host = user
			}
			pool.Players = append(pool.Players, user.ID)
			poolSize := len(pool.Clients)
			fmt.Println("Size of connection pool: ", poolSize)
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(GameResponse{
					Players: pool.Players, Host: pool.Host.ID, Type: 2, Body: "User has joined...", ID: user.ID}); err != nil {
					fmt.Println(err)
					return
				}
			}

		case user := <-pool.Unregister:
			delete(pool.Clients, user)
			fmt.Println("Size of connection pool", len(pool.Clients))
			pool.removeUserFromPool(user.ID)
			for client := range pool.Clients {
				client.Conn.WriteJSON(GameResponse{Players: pool.Players, Host: pool.Host.ID, Type: 3, Body: "User has disconnected...", ID: user.ID})
			}
		case message := <-pool.Broadcast:
			switch message.Body {
			case "start game":
				pool.GameStarted = true
				randJudge := rand.Intn(len(pool.Clients))
				pool.Judge = randJudge
				pool.Turn = pool.Judge
				playerAndAction := pool.getNextPlayerAndTheirRequiredAction()
				pool.Turn = playerAndAction.Turn
				initializeDecks()
				pool.initializeTable()
				for client := range pool.Clients {
					client.getCards()
					if err := client.Conn.WriteJSON(
						GameResponse{
							Players:       pool.Players,
							Host:          pool.Host.ID,
							Type:          5, //idk why 5 right now... idk if it even needs a type tbh but we'll include it for now
							Body:          "New Game Starting",
							Judge:         pool.Judge,
							TurnAndAction: playerAndAction,
							MyOpCards:     client.WhiteCards,
							MyNoShotCards: client.RedCards,
							GameStarted:   pool.GameStarted,
							CardsPlayed:   pool.Table,
						}); err != nil {
						fmt.Println(err)
						return
					}
				}
			case "new round":
				pool.Judge = nextPlayerToLeft(pool.Judge, len(pool.Players))
				pool.Turn = pool.Judge
				playerAndAction := pool.getNextPlayerAndTheirRequiredAction()
				pool.Turn = playerAndAction.Turn
				initializeDecks()
				pool.initializeTable()
				for client := range pool.Clients {
					client.getCards()
					if err := client.Conn.WriteJSON(
						GameResponse{
							Players:       pool.Players,
							Host:          pool.Host.ID,
							Type:          5, //idk why 5 right now... idk if it even needs a type tbh but we'll include it for now
							Body:          "New Game Starting",
							Judge:         pool.Judge,
							TurnAndAction: playerAndAction,
							MyOpCards:     client.WhiteCards,
							MyNoShotCards: client.RedCards,
							GameStarted:   pool.GameStarted,
							CardsPlayed:   pool.Table,
						}); err != nil {
						fmt.Println(err)
						return
					}
				}
			case "game ended":
				pool.GameStarted = false
				for client := range pool.Clients {
					client.Conn.WriteJSON(GameResponse{Type: 99, Body: "Game Ended"})
				}
			default:
				if strings.Contains(message.Body, OP_DELIMITER) {
					cardsPlayed := strings.Split(message.Body, OP_DELIMITER)
					player := pool.getClientFromName(message.ID)
					player.CardsOnTable.WhiteCards = cardsPlayed
					playerAndAction := pool.getNextPlayerAndTheirRequiredAction()
					pool.Turn = playerAndAction.Turn
					player.removeCards(cardsPlayed, "OP")
					pool.updateTable()
					for client := range pool.Clients {
						if err := client.Conn.WriteJSON(
							GameResponse{
								Players:       pool.Players,
								Host:          pool.Host.ID,
								Type:          5, //idk why 5 right now... idk if it even needs a type tbh but we'll include it for now
								Body:          "Something Happened",
								Judge:         pool.Judge,
								MyOpCards:     client.WhiteCards,
								MyNoShotCards: client.RedCards,
								GameStarted:   pool.GameStarted,
								CardsPlayed:   pool.Table,
								TurnAndAction: playerAndAction,
							}); err != nil {
							fmt.Println(err)
							return
						}
					}

				} else if strings.Contains(message.Body, NO_SHOT_DELIMITER) {
					cardsPlayed := strings.Split(message.Body, NO_SHOT_DELIMITER)
					player := pool.getClientFromName(pool.Players[pool.Turn])
					player.removeCards(cardsPlayed[:1], "noShot")
					playerAndAction := pool.getNextPlayerAndTheirRequiredAction()
					pool.Turn = playerAndAction.Turn
					//this is the player to the left so we can now set the played Card here.
					var newPlayer *Client
					if playerAndAction.FirstToLeft == -1 {
						newPlayer = pool.getClientFromName(pool.Players[pool.Turn])
					} else {
						newPlayer = pool.getClientFromName(pool.Players[playerAndAction.FirstToLeft])
					}
					newPlayer.CardsOnTable.RedCard = cardsPlayed[0]

					pool.updateTable()
					for client := range pool.Clients {
						if err := client.Conn.WriteJSON(
							GameResponse{
								Players:       pool.Players,
								Host:          pool.Host.ID,
								Type:          5, //idk why 5 right now... idk if it even needs a type tbh but we'll include it for now
								Body:          "Something Happened",
								Judge:         pool.Judge,
								MyOpCards:     client.WhiteCards,
								MyNoShotCards: client.RedCards,
								GameStarted:   pool.GameStarted,
								CardsPlayed:   pool.Table,
								TurnAndAction: playerAndAction,
							}); err != nil {
							fmt.Println(err)
							return
						}
					}
				} else if strings.Contains(message.Body, WINNER_DELIMITER) {
					for client := range pool.Clients {
						if err := client.Conn.WriteJSON(
							GameResponse{
								Players:       pool.Players,
								Host:          pool.Host.ID,
								Type:          6, //idk why 6 right now... idk if it even needs a type tbh but we'll include it for now
								Body:          "Host choosing whether we continue or end",
								Judge:         pool.Judge,
								TurnAndAction: playerAndActionBuilder(0, "Choose option", -1),
								MyOpCards:     client.WhiteCards,
								MyNoShotCards: client.RedCards,
								GameStarted:   pool.GameStarted,
								CardsPlayed:   pool.Table,
								Winner:        strings.Split(message.Body, WINNER_DELIMITER)[0],
							}); err != nil {
							fmt.Println(err)
							return
						}
					}
				} else {
					for client := range pool.Clients {
						if err := client.Conn.WriteJSON(message); err != nil {
							fmt.Println(err)
							return
						}
					}
				}
			}
		}
	}
}
