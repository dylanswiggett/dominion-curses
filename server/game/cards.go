package game

import (
	"strconv"
)

type CardAttributes struct {
	CardType int
	Cost int
	Param int // General purpose
	Name string // Must be unique
	Desc string
}

const (
	MoneyType = 0
	LandType  = 1
	ActionType= 2
)

type Card interface {
	GetAttributes() CardAttributes
	Act(sourcePlayerId int, game *Game) Player
}

/*
 * Actual card implementations go here:
 */

type Money struct {
	Value int
}

func (m Money) GetAttributes() CardAttributes {
	switch m.Value {
	case 3:
		return CardAttributes{MoneyType, 6, 3, "Gold", "Treasure with 3 buying power"}
	case 2:
		return CardAttributes{MoneyType, 3, 2, "Silver", "Treasure with 2 buying power"}
	default:
		return CardAttributes{MoneyType, 0, 1, "Copper", "Treasure with 1 buying power"}
	}
}

func (m Money) Act(sourcePlayerId int, game *Game) Player {
	return game.Players[sourcePlayerId]
}

type Land struct {
	Points int
}

func (l Land) GetAttributes() CardAttributes {
	switch l.Points {
	case 6:
		return CardAttributes{LandType, 8, 6, "Province", "Land worth 6 Victory point"}
	case 3:
		return CardAttributes{LandType, 5, 3, "Duchy", "Land worth 3 Victory points"}
	default:
		return CardAttributes{LandType, 2, 1, "Estate", "Land worth 1 Victory points"}
	}
}

func (l Land) Act(sourcePlayerId int, game *Game) Player {
	return game.Players[sourcePlayerId]
}

type Cellar struct {
}

func (c Cellar) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 2, 0, "Cellar", "+1 Action\n    Discard any number of cards.\n    +1 Card per card discarded"}
}

func (c Cellar) Act(sourcePlayerId int, game *Game) Player{
	player := game.Players[sourcePlayerId]
	numDiscarded := 0
	for {
		player.Tell("Your hand contains " + player.HandAsString())
		response := player.Ask("Select one to discard, or type 'end' \n" +
			                   "to finish and draw your new cards.")
		if response == "end" {
			break
		}
		cardNum, err := strconv.Atoi(response)
		if err != nil {
			player.Tell("That's not a number.")
		} else if cardNum < 0 || cardNum >= len(player.Hand){
			player.Tell("That's not a card.")
		} else {
			card := player.Hand[cardNum]
			player.DiscardCard(cardNum)
			player.Tell("Discarded a " + card.GetAttributes().Name)
			game.TellAllPlayersBut(sourcePlayerId, player.Name + " discarded a " + card.GetAttributes().Name)
			numDiscarded++
		}
	}
	player.Tell("Drawing " + strconv.Itoa(numDiscarded) + " new cards.")
	player.DrawCards(numDiscarded)
	player.Tell("Your hand now contains " + player.HandAsString())
	game.TellAllPlayersBut(sourcePlayerId, player.Name + " drew " + strconv.Itoa(numDiscarded) + " new cards.")
	player.AddActions(1)
	return player
}

type Market struct {
}

func (m Market) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 5, 0, "Market", "+1 Card\n    +1 Action\n    +1 Buy\n    +1 Money"}
}

func (m Market) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	player.AddActions(1)
	player.AddBuys(1)
	player.AddMoney(1)
	player.DrawCards(1)
	player.Tell("+1 Action, +1 Buy, +1 Money, +1 Card")
	return player
}

type Militia struct {
}

func (m Militia) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 4, 0, "Militia", "+2 Money\n    Each other player discards\n    down to 3 cards in his hand."}
}

func (m Militia) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	for i, player := range game.Players {
		if i == sourcePlayerId || len(player.Hand) <= 3 {
			continue
		}
		game.TellAllPlayersBut(i, player.Name + " is discarding cards.")
		canBlock := player.BlockAttack()
		if canBlock != "" {
			game.TellAllPlayersBut(i, canBlock)
			continue
		}
		for len(player.Hand) > 3 {
			player.Tell("Your hand contains " + player.HandAsString())
			response := player.Ask("Select a card to discard")
			cardNum, err := strconv.Atoi(response)
			if err != nil {
				player.Tell("That's not a number.")
			} else if cardNum < 0 || cardNum >= len(player.Hand){
				player.Tell("That's not a card.")
			} else {
				card := player.Hand[cardNum]
				player.DiscardCard(cardNum)
				player.Tell("Discarded a " + card.GetAttributes().Name)
				game.TellAllPlayersBut(i, player.Name + " discarded a " + card.GetAttributes().Name)
			}
		}
		game.Players[i] = player
	}
	player.AddMoney(2)
	return player
}

type Mine struct {
}

func (m Mine) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 5, 0, "Mine", "Trash a Treasure card from your hand.\n    Gain a Treasure card costing up to\n    3 more; put it into your hand."}
}

func (m Mine) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	if !player.HasMoneyCard() {
		player.Tell("You have no upgradeable money cards.")
		return player
	}
	for {
		player.Tell("You have " + player.HandAsString())
		response := player.Ask("Select a money card to upgrade, or type 'skip'")
		if response == "skip" {
			break
		}
		cardNum, err := strconv.Atoi(response)
		if err != nil {
			player.Tell("That's not a number.")
		} else if cardNum < 0 || cardNum >= len(player.Hand){
			player.Tell("That's not a card.")
		} else {
			card := player.Hand[cardNum]
			if card.GetAttributes().CardType != MoneyType {
				player.Tell("That's not money")
			} else if card.GetAttributes().Param == 3 {
				player.Tell("Can't upgrade gold")
			} else {
				upgrade := Money{card.GetAttributes().Param + 1}
				if game.Cards[upgrade] == 0 {
					player.Tell("No " + upgrade.GetAttributes().Name + " available.")
				} else {
					player.TrashCard(cardNum)
					player.Hand = append(player.Hand, upgrade)
					game.Cards[upgrade]--
					information := " upgraded a " + card.GetAttributes().Name + " to a " + upgrade.GetAttributes().Name
					player.Tell("You've" + information)
					game.TellAllPlayersBut(sourcePlayerId, player.Name + information)
					break
				}
			}
		}
	}
	return player
}

type Moat struct {
}

func (m Moat) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 2, 0, "Moat", "+2 Cards\n    When another player plays an Attack\n    card, you may reveal this from your\n    hand. If you do, you are unaffected\n    by that Attack."}
}

func (m Moat) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	player.Tell("You draw 2 cards")
	player.DrawCards(2)
	return player
}

type Remodel struct {
}

func (r Remodel) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 4, 0, "Remodel", "Trash a card from your hand.\n    Gain a card costing up to 2 more\n    than the trashed card."}
}

func (r Remodel) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	for {
		player.Tell("Your hand contains " + player.HandAsString())
		response := player.Ask("Select a card to remodel, or type 'skip' to abort.")
		if response == "skip" {
			break
		}
		cardNum, err := strconv.Atoi(response)
		if err != nil {
			player.Tell("That's not a number.")
		} else if cardNum < 0 || cardNum >= len(player.Hand){
			player.Tell("That's not a card.")
		} else {
			card := player.Hand[cardNum]
			maxValue := card.GetAttributes().Cost + 2
			for {
				player.Tell(game.CardsAsString())
				response1 := player.Ask("Select a new card worth up to " + strconv.Itoa(maxValue) + ",\nor type 'back' to select a different card to remodel.")
				if response1 == "back" {
					break
				}
				newCardNum, err := strconv.Atoi(response1)
				if err != nil {
					player.Tell("That's not a number.")
				}  else {
					_, msg, newCard := game.BuyCard(maxValue, newCardNum)
					if msg != "" {
						player.Tell(msg)
						continue
					}
					player.TrashCard(cardNum)
					player.Discard = append(player.Discard, newCard)
					infoMessage := " remodeled a " + card.GetAttributes().Name + " into a " + newCard.GetAttributes().Name + "."
					player.Tell("You" + infoMessage)
					game.TellAllPlayersBut(sourcePlayerId, player.Name + infoMessage)
					return player
				}
			}
		}
	}
	return player
}

type Smithy struct {
}

func (s Smithy) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 4, 0, "Smithy", "+3 cards"}
}

func (s Smithy) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	player.DrawCards(3)
	player.Tell("You draw 3 cards.")
	game.TellAllPlayersBut(sourcePlayerId, player.Name + " draws 3 cards.")
	return player
}

type Village struct {
}

func (v Village) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 3, 0, "Village", "+1 Card\n    +2 Actions"}
}

func (v Village) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	player.DrawCards(1)
	player.AddActions(2)
	player.Tell("You draw 1 card and gain 2 actions.")
	game.TellAllPlayersBut(sourcePlayerId, player.Name + " draws 1 card and gains 2 actions.")
	return player
}

type Woodcutter struct {
}

func (w Woodcutter) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 3, 0, "Woodcutter", "+1 Buy\n    +2 Money"}
}

func (w Woodcutter) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	player.AddBuys(1)
	player.AddMoney(2)
	player.Tell("You gain 1 buy and 2 money.")
	game.TellAllPlayersBut(sourcePlayerId, player.Name + " gains 1 buy and 2 money.")
	return player
}

type Workshop struct {
}

func (w Workshop) GetAttributes() CardAttributes {
	return CardAttributes{ActionType, 3, 0, "Workshop", "Gain a card costing up to 4"}
}

func (w Workshop) Act(sourcePlayerId int, game *Game) Player {
	player := game.Players[sourcePlayerId]
	for {
		player.Tell(game.CardsAsString())
		response := player.Ask("Select a card worth up to 4 to buy, or 'skip' to cancel")
		if response == "skip" {
			break
		}
		cardNum, err := strconv.Atoi(response)
		if err != nil {
			player.Tell("That's not a number.")
		} else {
			_, err, card := game.BuyCard(4, cardNum)
			if err != "" {
				player.Tell(err)
				continue
			}
			player.Discard = append(player.Discard, card)
			info :=  " bought a " + card.GetAttributes().Name + "."
			player.Tell("You" + info)
			game.TellAllPlayersBut(sourcePlayerId, player.Name + info)
			break
		}
	}
	return player
}