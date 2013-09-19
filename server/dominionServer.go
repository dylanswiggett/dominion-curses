package main

import(
	"fmt"
	"dylanswiggett.com/dominion/server/game"
	"strconv"
	"math/rand"
	"time"
	"net"
	"strings"
)

var gameData game.Game

var numTestPlayers = 1

var players = make(map[int]game.Player)

var helpMessage = 	"Commands (type any time, even when asked a question):\n" +
					"    'help'              - Display this help message\n" +
					"    'name'              - Output your name\n" +
					"    'describe <number>' - Describe a card (type 'deck' for numbers)\n" +
					"    'hand'              - List cards in your hand\n" +
					"    'deck'              - List unbought cards"

func handlePlayerInput(playerId int, questionSent chan int, response chan string) {
	for {
		str := make([]byte, 1000)
		n, err := players[playerId].Connection.Read(str)
		if err != nil {
			fmt.Println("Error:", err)
		}
		input := strings.Split(strings.Trim(string([]byte(str)[:n]), " "), " ")
		player := players[playerId]
		switch input[0] {
		case "help":
			player.TellWithoutConfirmation(helpMessage)
		case "name":
			player.TellWithoutConfirmation("Your name is " +
					players[playerId].Name)
		case "hand":
			player.TellWithoutConfirmation("Your hand contains " +
				player.HandAsString())
		case "deck":
			player.TellWithoutConfirmation(gameData.CardsAsString() +
				"\nType 'describe <number>' for more information")
		case "describe":
			if len(input) < 2 {
				player.TellWithoutConfirmation("Please ask 'describe <card number>'")
			} else {
				cardNum, err := strconv.Atoi(input[1])
				if (err != nil) {
					player.TellWithoutConfirmation("Please ask 'describe <card number>'")
				} else {
					cards := gameData.GetCardList()
					if cardNum < 0 || cardNum >= len(cards) {
						player.TellWithoutConfirmation("Not a valid card number!")
					} else {
						card := cards[cardNum].GetAttributes()
						player.TellWithoutConfirmation(card.Name +
							":\n    " + card.Desc)
					}
				}
			}
		default:
			select {
			case i := <- questionSent:
				if i == 0 {
					response <- input[0]
				}
			default:
				player.TellWithoutConfirmation("What? Type 'help' for a list of commands.")
			}
		}
		time.Sleep(100000)
	}
}

func main() {
	rand.Seed( time.Now().UTC().UnixNano())

	fmt.Println("Running Dominion server...")

	cards := make(map[game.Card]int)
	cards[game.Money{1}] = 60
	cards[game.Money{2}] = 40
	cards[game.Money{3}] = 30
	cards[game.Land{1}] = 24
	cards[game.Land{3}] = 12
	cards[game.Land{6}] = 12

	cards[game.Cellar{}] = 10
	cards[game.Market{}] = 10
	cards[game.Militia{}] = 10
	cards[game.Mine{}] = 10
	cards[game.Moat{}] = 10
	cards[game.Remodel{}] = 10
	cards[game.Smithy{}] = 10
	cards[game.Village{}] = 10
	cards[game.Woodcutter{}] = 10
	cards[game.Workshop{}] = 10

	fmt.Println("Selecting action cards")
	// TODO: Implement card selection here

	// for i := 0; i < numTestPlayers; i++ {
	// 	players[i] = game.Player{nil, nil, nil,
	// 		"Player " + strconv.Itoa(i), game.Turn{-1, 0, 0}}
	// }

	numSupplyPiles := len(cards)

	fmt.Println("Getting players (hit enter to begin immediately)...")

	ln, err := net.Listen("tcp", ":4000")
	if err != nil {
		fmt.Println(err)
		return
	}
	i := 0
	for i < numTestPlayers {
		fmt.Println("Waiting for player.")
		conn, err := ln.Accept()
		fmt.Println("Found connection.")
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		fmt.Println("Waiting for name...")
		name := make([]byte, 100)
		n, err := conn.Read(name)
		if (err != nil) {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Player's name is " + string(name[:n]))

		questionSent := make(chan int, 1)
		response := make(chan string, 1000)

		player := game.Player{nil, nil, nil, string(name[:n]),
			game.Turn{-1, 0, 0}, conn, questionSent, response}
		players[i] = player

		go handlePlayerInput(i, questionSent, response)

	 	player.Tell("You're in the game! Your name is " +
	 		players[i].Name + "\n\nType help at any time for a list of commands.")
		

		gameData = game.Game{players, cards}
		gameData.TellAllPlayersBut(i, players[i].Name + " has joined the game")

		i++
	}

	gameOver := false

	fmt.Println("Dealing initial cards...")
	for p, _ := range gameData.Players {
		player := gameData.Players[p]
		player.Deck = make([]game.Card, 10)
		for i := 0; i < 10; i++ {
			switch {
			case i < 7:
				player.Deck[i] = gameData.DrawCard(game.Money{1})
			default:
				player.Deck[i] = gameData.DrawCard(game.Land{1})
			}
		}
		player.ShuffleDeck()
		player.DrawHand()
		player.Tell("You've drawn " + player.HandAsString())
		// fmt.Println(gameData.CardsAsString())
		gameData.Players[p] = player
	}
	for !gameOver {
		for i, player := range gameData.Players {
			player.Tell("IT'S YOUR TURN!")
			player.CurrentTurn = game.Turn{0, 1, 1}
			gameData.TellAllPlayersBut(i, "It's " + player.Name + "'s turn!")

			player.Tell("ACTION PHASE:")
			gameData.TellAllPlayersBut(i, "It's " + player.Name + "'s action phase.")
			for player.CurrentTurn.NumActions > 0 {
				numActionsLeft := strconv.Itoa(player.CurrentTurn.NumActions)
				player.Tell("You have " + numActionsLeft + " actions")
				gameData.TellAllPlayersBut(i, player.Name + " has " + numActionsLeft + " actions.")
				player.Tell("Your hand contains " + player.HandAsString())
				playerSelection := player.Ask("Type a number to play, or 'skip' to go to Buy phase")
				if playerSelection == "skip" {
					break
				}
				cardNum, err := strconv.Atoi(playerSelection)
				if cardNum < 0 || len(player.Hand) <= cardNum || err != nil {
					player.Tell("That card doesn't exist.")
					continue
				} else if player.Hand[cardNum].GetAttributes().CardType !=
						game.ActionType {
					player.Tell("That's not an action card.")
					continue
				}

				cardPlayed := player.Hand[cardNum]
				cardName := cardPlayed.GetAttributes().Name
				player.Tell("You played " + cardName + ".")
				gameData.TellAllPlayersBut(i, player.Name + " played " + cardName + ".")
				player.DiscardCard(cardNum)
				gameData.Players[i] = player
				player = cardPlayed.Act(i, &gameData)
				player.CurrentTurn.NumActions--
				gameData.Players[i] = player
			}
			if player.CurrentTurn.NumActions == 0 {
				player.Tell("No more actions")
			}
			for _, card := range player.Hand {
				if card.GetAttributes().CardType == game.MoneyType {
					player.CurrentTurn.MoneyAvailable +=
						card.GetAttributes().Param
				}
			}

			player.Tell("BUY PHASE:")
			gameData.TellAllPlayersBut(i, "It's " + player.Name + "'s buy phase.")
			for player.CurrentTurn.NumBuys > 0 {
				buyingPower := strconv.Itoa(player.CurrentTurn.MoneyAvailable)
				numBuysLeft := strconv.Itoa(player.CurrentTurn.NumBuys)
				player.Tell("You have " + buyingPower + " buying power and " +
					numBuysLeft + " buys")
				gameData.TellAllPlayersBut(i, player.Name + " has " + numBuysLeft +
					" buys and " + buyingPower + " buying power.")
				player.Tell(gameData.CardsAsString())
				playerSelection := player.Ask("Type a number to buy, or 'skip' to end turn.")
				if playerSelection == "skip" {
					break
				}
				cardNum, err := strconv.Atoi(playerSelection)
				if err != nil {
					player.Tell("That card doesn't exist.")
					continue
				}
				newMoney, err1, cardBought := gameData.BuyCard(
					player.CurrentTurn.MoneyAvailable, cardNum)
				if err1 != "" {
					player.Tell(err1)
					continue
				}
				
				player.Tell("You bought a " + cardBought.GetAttributes().Name + ".")
				gameData.TellAllPlayersBut(i, player.Name + " bought a " + cardBought.GetAttributes().Name + ".")
				player.Discard = append(player.Discard, cardBought)
				player.CurrentTurn.MoneyAvailable = newMoney
				player.CurrentTurn.NumBuys--
			}

			player.Tell("You draw a new hand.")
			player.DrawHand()
			player.Tell("Your hand contains " + player.HandAsString())

			gameData.Players[i] = player

			if gameData.Cards[game.Land{6}] <= 0 ||
				len(gameData.GetCardList()) + 3 <= numSupplyPiles {
				player.Tell("You've ended the game!")
				gameData.TellAllPlayersBut(i, "The game has ended!")
				maxPoints := -10000000
				bestPlayer := player
				bestPlayerIndex := 0
				for playerIndex, curPlayer := range gameData.Players {
					playerPoints := curPlayer.VictoryPoints()
					pointStr := strconv.Itoa(playerPoints)
					curPlayer.Tell("You have " + pointStr + " victory points.")
					gameData.TellAllPlayersBut(playerIndex, curPlayer.Name + " has " + pointStr + " victory points.")
					if playerPoints > maxPoints {
						maxPoints = playerPoints
						bestPlayer = curPlayer
						bestPlayerIndex = playerIndex
					}
				}
				bestPlayer.Tell("You win!")
				gameData.TellAllPlayersBut(bestPlayerIndex, bestPlayer.Name + " has won!")
				return
			}
		}
	}
}