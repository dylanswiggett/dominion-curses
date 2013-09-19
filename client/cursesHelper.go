package main

import (
	"github.com/nsf/termbox-go"
	"fmt"
	"strings"
)

func TermboxDraw(x, y int, c rune) {
	termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
}

func TermboxDrawBox(x1, y1, x2, y2 int) {
	for x := x1 + 1; x < x2; x++ {
		TermboxDraw(x, y1, '-')
		TermboxDraw(x, y2, '-')
	}
	for y := y1 + 1; y < y2; y++ {
		TermboxDraw(x1, y, '|')
		TermboxDraw(x2, y, '|')
	}
	TermboxDraw(x1, y1, '+')
	TermboxDraw(x2, y1, '+')
	TermboxDraw(x1, y2, '+')
	TermboxDraw(x2, y2, '+')
}

func TermboxClear() {
	err := termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
	if err != nil {
		fmt.Println("Error:",err)
		return
	}
}

func TermboxWrite(x, y int, text string) {
	for line, text := range strings.Split(text, "\n") {
		for i := 0; i < len(text); i++ {
			TermboxDraw(x + i, y + line, rune([]byte(text)[i]))
		}
	}
}

func TermboxWriteColored(x, y int, text string, color termbox.Attribute) {
	for line, text := range strings.Split(text, "\n") {
		for i := 0; i < len(text); i++ {
			termbox.SetCell(x + i, y + line, rune([]byte(text)[i]), color, termbox.ColorDefault)
		}
	}
}

func TermboxDrawColoredMessagesInBox(x1, y1, x2, y2 int, messages []string, colors []termbox.Attribute) {
	currentY := y2
	for i := len(messages) - 1; i >= 0; i-- {
		messageParts := strings.Split(messages[i], "\n")
		for j := len(messageParts) - 1; j >= 0; j-- {
			if currentY < y1 {
				return
			}
			if len(messageParts[j]) > (x2 - x1) {
				messageParts[j] = string([]byte(messageParts[j])[:(x2-x1)])
			}
			TermboxWriteColored(x1, currentY, messageParts[j], colors[i])
			currentY --
		}
	}
}