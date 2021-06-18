package websocket

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
