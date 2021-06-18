package websocket

import (
	"fmt"
)

type playerAndActionRequired struct {
	Turn        int
	Action      string
	FirstToLeft int
}
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

func nextPlayerToLeft(idx int, poolSize int) int {
	if idx-1 < 0 {
		return poolSize
	}
	return idx - 1
}
func (pool *Pool) getNextPlayerAndTheirRequiredAction() playerAndActionRequired {
	poolSize := len(pool.Clients) - 1
	next := nextPlayerToLeft(pool.Turn, poolSize)
	nextNext := nextPlayerToLeft(next, poolSize)
	if pool.Judge == next {
		playersRedCards := pool.getClientFromName(pool.Players[pool.Turn]).RedCards
		if len(playersRedCards) == MaxNoShotCards {
			return playerAndActionBuilder(nextNext, "Choose noShot cards", -1)
		} else {
			return playerAndActionBuilder(pool.Judge, "Pick a winner", nextNext)
		}

	} else {
		if len(pool.getClientFromName(pool.Players[next]).CardsOnTable.WhiteCards) == 0 {
			return playerAndActionBuilder(next, "Choose OP cards", -1)
		} else {
			return playerAndActionBuilder(next, "Choose noShot cards", -1)
		}
	}
}

func playerAndActionBuilder(player int, action string, nextNext int) playerAndActionRequired {
	return playerAndActionRequired{Turn: player, Action: action, FirstToLeft: nextNext}
}

func (pool *Pool) NewRound() {
	initializeDecks()
	pool.initializeTable()
	pool.Turn = pool.Judge
	playerAndAction := pool.getNextPlayerAndTheirRequiredAction()
	pool.Turn = playerAndAction.Turn
	pool.GameStarted = true
	for client := range pool.Clients {
		client.getCards()
		if err := client.Conn.WriteJSON(
			GameResponse{
				Players:       pool.Players,
				Host:          pool.Host.ID,
				Type:          5, //idk why 5 right now... idk if it even needs a type tbh but we'll include it for now
				Body:          "New Round Starting",
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
}

func (pool *Pool) updateGameStateAfterCardPlayed(cardsPlayed []string, cardType string) {
	player := pool.getClientFromName(pool.Players[pool.Turn])
	player.removeCards(cardsPlayed, cardType)
	playerAndAction := pool.getNextPlayerAndTheirRequiredAction()
	pool.Turn = playerAndAction.Turn
	if cardType == "noShot" {
		var newPlayer *Client
		if playerAndAction.FirstToLeft == -1 {
			newPlayer = pool.getClientFromName(pool.Players[pool.Turn])
		} else {
			newPlayer = pool.getClientFromName(pool.Players[playerAndAction.FirstToLeft])
		}
		newPlayer.CardsOnTable.RedCard = cardsPlayed[0]
	} else {
		player.CardsOnTable.WhiteCards = cardsPlayed
	}
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
}
