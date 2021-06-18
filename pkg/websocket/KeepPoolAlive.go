package websocket

import (
	"fmt"
	"time"
)

const KEEP_ALIVE_TYPE = 55
const HOST = 0

func (pool *Pool) StartPings() {
	for {
		fmt.Println("called")
		time.Sleep(54 * time.Second)
		pool.KeepPoolAlive()
	}
}
func (pool *Pool) KeepPoolAlive() {
	if len(pool.Players) == 0 {
		return
	} else {
		client := pool.getClientFromName(pool.Players[HOST])
		if err := client.Conn.WriteJSON(GameResponse{
			Host: pool.Host.ID,
			Type: KEEP_ALIVE_TYPE,
			Body: "...",
		}); err != nil {
			fmt.Println(err)
			return
		}
	}

}
