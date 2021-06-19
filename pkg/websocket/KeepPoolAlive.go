package websocket

import (
	"fmt"
	"time"
)

const KEEP_ALIVE_TYPE = 55
const HOST = 0
const SLEEP_TIME = 54

/*
	Heroku kills any websocket connections that hasn't had anything pushed through
	for 55 seconds, so I have a goroutine that sleeps for 54 seconds and then sends a
	blank piece of info over that I can just ignore by checking Type==KEEP_ALIVE_TYPE
*/
func (pool *Pool) StartPings() {
	for {
		fmt.Println("called")
		time.Sleep(SLEEP_TIME * time.Second)
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
