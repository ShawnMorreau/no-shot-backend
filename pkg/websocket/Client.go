package websocket

import (
	"fmt"
	"log"

	"github.com/Pallinder/go-randomdata"
	"github.com/gorilla/websocket"
)

//we can probably separate this so a Client has a Hand
type Client struct {
	ID           string
	Conn         *websocket.Conn
	Pool         *Pool
	RedCards     []string
	WhiteCards   []string
	CardsOnTable Option
	RoundsWon    int
}
type Option struct {
	Player     string
	RedCard    string
	WhiteCards []string
}

type Message struct {
	Type int
	Body string
	ID   string
}

func (c *Client) Read() {
	c.ID = randomdata.SillyName()
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		message := Message{Type: messageType, Body: string(p), ID: c.ID}
		c.Pool.Broadcast <- message
		fmt.Printf("message Recieved: %v\n", message)
	}
}
func (c *Client) getCards() {
	if cardsNeeded := MaxOPCards - len(c.WhiteCards); cardsNeeded != 0 {
		c.WhiteCards = append(c.WhiteCards, getRandomCardsFromDeck(cardsNeeded, OPDeck)...)
	}
	if cardsNeeded := MaxNoShotCards - len(c.RedCards); cardsNeeded != 0 {
		c.RedCards = append(c.RedCards, getRandomCardsFromDeck(cardsNeeded, noShotDeck)...)
	}
}
func (c *Client) removeCards(cards []string, kind string) {
	if kind == "OP" {
		for _, card := range cards {
			c.WhiteCards[c.getIndexOfCardInHand(card, kind)] = c.WhiteCards[len(c.WhiteCards)-1]
			c.WhiteCards[len(c.WhiteCards)-1] = ""
			c.WhiteCards = c.WhiteCards[:len(c.WhiteCards)-1]
		}
	} else {
		for _, card := range cards {
			c.RedCards[c.getIndexOfCardInHand(card, kind)] = c.RedCards[len(c.RedCards)-1]
			c.RedCards[len(c.RedCards)-1] = ""
			c.RedCards = c.RedCards[:len(c.RedCards)-1]
		}
	}

}
func (c *Client) getIndexOfCardInHand(card string, kind string) int {
	if kind == "OP" {
		for i, c := range c.WhiteCards {
			if c == card {
				return i
			}
		}
	} else {
		for i, c := range c.RedCards {
			if c == card {
				return i
			}
		}
	}
	return -1
}
