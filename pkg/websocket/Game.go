package websocket

type playerAndActionRequired struct {
	Turn        int
	Action      string
	FirstToLeft int
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

func (pool *Pool) updateTable() {
	var newArr []Option
	for c := range pool.Clients {
		if c != pool.getClientFromName(pool.Players[pool.Judge]) {
			newArr = append(newArr, c.CardsOnTable)
		}
	}
	pool.Table = newArr
}

func (pool *Pool) initializeTable() {
	var newArr []Option
	for c := range pool.Clients {
		if c != pool.getClientFromName(pool.Players[pool.Judge]) {
			c.CardsOnTable.Player = c.ID
			c.CardsOnTable.RedCard = ""
			c.CardsOnTable.WhiteCards = []string{}
			newArr = append(newArr, c.CardsOnTable)
		}
	}
	pool.Table = newArr
}
