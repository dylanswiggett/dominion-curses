package game

import(
	"math/rand"
	"strconv"
	"bytes"
	"sort"
	"fmt"
	"net"
)

type Turn struct {
	MoneyAvailable int
	NumActions int
	NumBuys int
}

type Player struct {
	Deck []Card
	Discard []Card
	Hand []Card
	Name string
	CurrentTurn Turn
	Connection net.Conn
	QuestionSent chan int
	Response chan string
}

func (p *Player) AddActions(num int) {
	p.CurrentTurn.NumActions += num
}

func (p *Player) AddBuys(num int) {
	p.CurrentTurn.NumBuys += num
}

func (p *Player) AddMoney(num int) {
	p.CurrentTurn.MoneyAvailable += num
}

type Game struct {
	Players map[int]Player // Player IDs
	Cards map[Card]int
}


// Implement
func (p *Player) Ask(question string) string {
	str := p.Tell("ask " + question)
	return str
}

func (p *Player) TellWithoutConfirmation(message string) {
	fmt.Fprintf(p.Connection, "noconfirm " + message)
}

// Implement
func (p *Player) Tell(message string) string {
	p.QuestionSent <- 0
	fmt.Fprintf(p.Connection, message)
	return <-p.Response
	// str := make([]byte, 1000)
	// n, err := p.Connection.Read(str)
	// if (err != nil) {
	// 	fmt.Println("Error:", err)
	// 	return "broken"
	// } else {
	// 	return string([]byte(str)[:n])
	// }
}

func (g *Game) TellAllPlayersBut(playerId int, message string) {
	for i, otherPlayer := range g.Players {
		if i != playerId {
			otherPlayer.Tell(message)
		}
	}
}

func (g *Game) DrawCard(c Card) Card {
	if g.Cards[c] <= 0 {
		return nil
	} else {
		g.Cards[c]--
		return c
	}
}

func (p *Player) ShuffleDeck() {
	oldDeck := p.Deck
	newDeck := make([]Card, len(oldDeck))
	perm    := rand.Perm(len(oldDeck))
	for i, v := range perm {
		newDeck[v] = oldDeck[i]
	}
	p.Deck = newDeck
}

func (p *Player) DrawCards(num int) {
	for i := 0; i < num; i++ {
		if len(p.Deck) == 0 {
			p.Deck = p.Discard
			p.ShuffleDeck()
			p.Discard = make([]Card, 0)
		}
		p.Hand = append(p.Hand, p.Deck[0])
		p.Deck = p.Deck[1:]
	}
}

func (p *Player) VictoryPoints() int {
	numPoints := 0
	for _, card := range append(p.Deck, append(p.Hand, p.Discard ...) ...){
		if card.GetAttributes().CardType == LandType {
			numPoints += card.GetAttributes().Param
		}
	}
	return numPoints
}

// Assumes real card number
func (p *Player) DiscardCard(cardNum int) {
	p.Discard = append(p.Discard, p.Hand[cardNum])
	p.Hand = append(p.Hand[:cardNum], p.Hand[cardNum + 1:] ...)
}

func (p *Player) TrashCard(cardNum int) {
	p.Hand = append(p.Hand[:cardNum], p.Hand[cardNum + 1:] ...)
}

func (p *Player) DrawHand() {
	p.Discard = append(p.Discard, p.Hand ...)
	p.Hand = make([]Card, 0)
	p.DrawCards(5)
}

func (p *Player) HasActionCard() bool {
	for _, c := range p.Hand {
		if c.GetAttributes().CardType == ActionType {
			return true
		}
	}
	return false
}

func (p *Player) HasMoneyCard() bool {
	for _, c := range p.Hand {
		if c.GetAttributes().CardType == MoneyType {
			return true
		}
	}
	return false
}

func (p *Player) HasMoneyCardNotGold() bool {
	for _, c := range p.Hand {
		if c.GetAttributes().CardType == MoneyType && c.GetAttributes().Param != 3{
			return true
		}
	}
	return false
}

func (p *Player) BlockAttack() string {
	for _, card := range p.Hand {
		if card.GetAttributes().Name == "Moat" {
			for {
				response := p.Ask("You have a moat. Reveal it? (Y/n)")
				if response == "" || response == "Y" || response == "y" {
					return p.Name + " has a moat."
				} else if response == "n" {
					return ""
				}
			}
		}
	}
	return ""
}

func (p *Player) HandAsString() string {
	str := string(strconv.Itoa(len(p.Hand))) + " cards:"
	for i := 0; i < len(p.Hand); i++ {
		str += "\n    (" + strconv.Itoa(i) + ") " +
			p.Hand[i].GetAttributes().Name
	}
	return str
}

type CardSort struct {
	cards []Card
}

func (c CardSort) Len() int {
	return len(c.cards)
}

func (c CardSort) Swap(i, j int) {
	c.cards[i], c.cards[j] = c.cards[j], c.cards[i]
}

func (c CardSort) Less(i, j int) bool {
	attrI := c.cards[i].GetAttributes()
	attrJ := c.cards[j].GetAttributes()
	if attrI.CardType != attrJ.CardType {
		return attrI.CardType < attrJ.CardType
	} else if attrI.Cost != attrJ.Cost {
		return attrI.Cost < attrJ.Cost
	} else {
		return (bytes.Compare([]byte(attrI.Name),
							  []byte(attrJ.Name)) > 0)
	}
}

func (g Game) GetCardList() []Card {
	keys := make([]Card, 0)
	for card, num := range g.Cards {
		if num > 0 {
			keys = append(keys, card)
		}
	}
	sort.Sort(CardSort{keys})
	return keys
}

func (g Game) CardsAsString() string {
	str := "Cards available:"
	for i, card := range g.GetCardList() {
		c := card.GetAttributes()
		str += "\n    (" + strconv.Itoa(i) + ") " + c.Name +
			": costs " + strconv.Itoa(c.Cost) + ", " +
			strconv.Itoa(g.Cards[card]) + " left."
	}
	// for card, num := range g.Cards {
	// 	str += "\n\t" + card.GetAttributes().Name +
	// 		" (" + strconv.Itoa(num) + " available)"
	// }
	return str
}

func (g *Game) BuyCard(moneyAvailable, cardNum int) (int, string, Card) {
	cards := g.GetCardList()
	if cardNum < 0 || len(cards) <= cardNum {
		return moneyAvailable, "That card doesn't exist.", nil
	} else if cards[cardNum].GetAttributes().Cost > moneyAvailable {
		return moneyAvailable, "Not enough money.", nil
	}
	g.Cards[cards[cardNum]]--
	return moneyAvailable - cards[cardNum].GetAttributes().Cost,
		"", cards[cardNum]
}