package websocket

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
)

var noShotDeck Deck
var OPDeck Deck

const MaxNoShotCards = 3
const MaxOPCards = 5

type Deck struct {
	Deck []Card `json:"Cards"`
}

type Card struct {
	Value  string `json:"value"`
	Indeck bool   `json:"inDeck"`
}

// Consider switching this to take in the file as an argument from a table obj or something
// and then that way it could be ran as a goroutine and get the decks initialized immediately
//for now this will work.
func initializeDecks() {
	f, err := os.Open("Flaws.json")
	if err != nil {
		log.Print(err)
	}
	f2, err := os.Open("Perks.json")
	if err != nil {
		log.Print(err)
	}
	d := json.NewDecoder(f)
	if err := d.Decode(&noShotDeck); err != nil {
		log.Println(err)
	}
	d2 := json.NewDecoder(f2)
	if err := d2.Decode(&OPDeck); err != nil {
		log.Println(err)
	}
}
func getRandomCardsFromDeck(numCards int, deck Deck) []string {
	var cards []string
	for len(cards) < numCards {
		if randCard := deck.Deck[rand.Intn(len(deck.Deck))]; randCard.Indeck {
			randCard.Indeck = false
			cards = append(cards, randCard.Value)
		}
	}
	return cards
}
