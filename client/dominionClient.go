package main

import (
	"net"
	"strings"
	"github.com/nsf/termbox-go"
	"time"
	"fmt"
)

func handleInput(ch chan termbox.Event) {
	for {
		time.Sleep(1000000)
		evt := termbox.PollEvent()
		ch <- evt
	}
}

func handleServer(ch chan string) {
	for {
		str := make([]byte, 1000)
		//connection.SetReadDeadline(time.Now().Add(time.Millisecond))
		n, err := connection.Read(str)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Hit ESC to end program.")
			return
		}

		message := strings.Trim(string(str[:n]), " ")

		if strings.HasPrefix(message, "noconfirm ") {
			message = string(str[10:n])
		} else if !strings.HasPrefix(message, "ask ") {
			sendToServer("recieved")
		}
		ch <- message
	}
}

var CurrentInput string
var ResponseInput string
var w, h int
var messages []string
var messageColors []termbox.Attribute
var connection net.Conn

var inputPrompt = "INPUT: "
var questionPrompt = "RESPONSE: "
var prompt = inputPrompt

func RefreshScreen() {
	if w < 50 || h < 15 {
		ScreenTooSmall()
		return
	} 
	/*
	 * Draw UI
	 */
	TermboxClear()
	TermboxDrawBox(0, 0, w - 1, 2)	// Header box
	TermboxDrawBox(0, h - 3, w - 1, h - 1)	// Input box
	TermboxWrite(2, h - 2, prompt + CurrentInput)
	exitInstruction := "Press ESC to exit"
	TermboxWrite(w - (len(exitInstruction) + 2), 1, exitInstruction)
	TermboxWrite(2, 1, "DOMINION CLIENT V0.3")
	TermboxDrawColoredMessagesInBox(3, 3, w - 1, h - 4, messages, messageColors)

	termbox.Flush()
}

func ScreenTooSmall() {
	TermboxClear()
	var msg string
	switch {
	case w > 30:
		msg = "Terminal is too small"
	case w > 20:
		msg = "Terminal too small"
	case w > 15:
		msg = "Small terminal"
	case w > 7:
		msg = "Small"
	case w > 2:
		msg = "Why"
	default:
		msg = "y"
	}
	TermboxWrite((w - len(msg)) / 2, h / 2, msg)
	termbox.Flush()
}

func addMessage(msg string, color termbox.Attribute) {
	messages = append(messages, msg)
	messageColors = append(messageColors, color)
}

func sendToServer(msg string) {
	fmt.Fprintf(connection, msg)
}

func tellServer(msg string) {
	sendToServer(msg)
	addMessage(msg, termbox.ColorGreen)
}

func main() {
	var hostName string
	name := ""

	fmt.Println("Welcome to the dominion client!")
	for name == "" {
		fmt.Println("What is your name?")
		fmt.Scanf("%s", &name)

		if name == "You" || name == "you"{
			fmt.Println("Invalid name.")
			name = ""
		} else if len(name) > 15 {
			fmt.Println("Name must be 15 characters or less")
			name = ""
		}
	}

	fmt.Print("Enter your server's ip and port: ")
	fmt.Scanf("%s", &hostName)
	fmt.Println("Connecting...")

	conn, err := net.Dial("tcp", hostName)
	connection = conn

	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed to connect to host.")
		fmt.Println("Exiting.")
		return
	} else {
		fmt.Println("Connected.")
		fmt.Fprintf(conn, name)
		addMessage("Connected to host " + hostName, termbox.ColorRed)
	}

	termbox.Init()
	defer termbox.Close()

	TermboxClear()

	ch := make(chan termbox.Event)
	server := make(chan string, 1000)

	w, h = termbox.Size()

	go handleInput(ch)
	go handleServer(server)

	CurrentInput = ""

	RefreshScreen()

	for {
		select {
		case evt := <-ch:
			if evt.Width > 0 {
				w,h = evt.Width, evt.Height
			} else if evt.Key == termbox.KeyEsc {
				return
			} else if evt.Key == termbox.KeyBackspace2 ||
					  evt.Key == termbox.KeyBackspace {
				if len(CurrentInput) > 0 {
					CurrentInput = string([]byte(CurrentInput)[:len(CurrentInput) - 1])
				}	
			} else if evt.Key == termbox.KeySpace {
				CurrentInput += " "
			} else if evt.Key == termbox.KeyEnter {
				tellServer(CurrentInput)
				CurrentInput = ""
			} else if evt.Ch != 0 {
				if len(CurrentInput) < 35 {
					CurrentInput += string(evt.Ch)
				}
			}
			RefreshScreen()
		case message := <-server:
			if strings.HasPrefix(message, "ask ") {
				addMessage(string([]byte(message)[4:]), termbox.ColorYellow)
			} else {
				addMessage(message, termbox.ColorWhite)
			}
			RefreshScreen()
		default:
		}

		time.Sleep(1000000)
	}

	// fmt.Fprintf(conn, name)

	// for {
	// 	str := make([]byte, 1000)
	// 	n, err := conn.Read(str)
	// 	if (err != nil) {
	// 		fmt.Println("Error:", err)
	// 		return
	// 	}
	// 	str1 := string([]byte(str)[:n])
	// 	if strings.HasPrefix(str1, "ask") {
	// 		str1 = string([]byte(str)[4:n])
	// 		fmt.Print(str1 + "\n: ")
	// 		var returnValue string
	// 		fmt.Scanf("%s", &returnValue)
	// 		fmt.Fprintf(conn, returnValue)
	// 	} else {
	// 		fmt.Println(str1)
	// 		fmt.Fprintf(conn, "recieved")
	// 	}
	// }
}