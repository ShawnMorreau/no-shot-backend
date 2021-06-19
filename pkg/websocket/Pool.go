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
	go pool.StartPings()
	for {
		select {
		case user := <-pool.Register:
			pool.Clients[user] = true
			if len(pool.Clients) == 1 {
				pool.Host = user
			}
			pool.Players = append(pool.Players, user.ID)
			fmt.Println("Size of connection pool: ", len(pool.Clients))
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
				pool.Judge = rand.Intn(len(pool.Clients))
				initializeDecks()
				pool.NewRound()
			case "new round":
				pool.Judge = nextPlayerToLeft(pool.Judge, len(pool.Players)-1)
				fmt.Printf(fmt.Sprintf("Judges value is %v and turn is %v", pool.Judge, pool.Turn) + "\n")
				pool.NewRound()
			case "game ended":
				pool.GameStarted = false
				for client := range pool.Clients {
					client.Conn.WriteJSON(GameResponse{Type: 99, Body: "Game Ended"})
				}
			default:
				if strings.Contains(message.Body, OP_DELIMITER) {
					cardsPlayed := strings.Split(message.Body, OP_DELIMITER)
					pool.updateGameStateAfterCardPlayed(cardsPlayed, "OP")

				} else if strings.Contains(message.Body, NO_SHOT_DELIMITER) {
					cardsPlayed := strings.Split(message.Body, NO_SHOT_DELIMITER)
					pool.updateGameStateAfterCardPlayed(cardsPlayed[:1], "noShot")

				} else if strings.Contains(message.Body, WINNER_DELIMITER) {
					for client := range pool.Clients {
						if err := client.Conn.WriteJSON(
							GameResponse{
								Type:          6,
								Body:          "Host choosing whether we continue or end",
								TurnAndAction: playerAndActionBuilder(0, "Choose option", -1),
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
