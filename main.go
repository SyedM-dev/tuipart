package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}
	defer screen.Fini()

	err = screen.Init()
	if err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}

	screen.Clear()

	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	screen.SetContent(10, 5, 'H', nil, style)
	screen.SetContent(11, 5, 'i', nil, style)
	draw_box(screen, 10, 10, 40, 20, "")
	draw_box(screen, 45, 10, 85, 20, "═║╔╚╗╝")
	screen.Show()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Rune() == 'q' {
				return
			}
			if ev.Key() == tcell.KeyCtrlC {
				return
			}
		case *tcell.EventResize:
			screen.Sync()
		}
	}
}

func draw_box(s tcell.Screen, x1, y1, x2, y2 int, chars_i string) {
	if chars_i == "" {
		chars_i = "─│╭╰╮╯" // "━┃┏┗┓┛"
	}
	chars := []rune(chars_i)
	for x := x1; x <= x2; x++ {
		s.SetContent(x, y1, chars[0], nil, tcell.StyleDefault)
	}
	for x := x1; x <= x2; x++ {
		s.SetContent(x, y2, chars[0], nil, tcell.StyleDefault)
	}
	for x := y1 + 1; x < y2; x++ {
		s.SetContent(x1, x, chars[1], nil, tcell.StyleDefault)
	}
	for x := y1 + 1; x < y2; x++ {
		s.SetContent(x2, x, chars[1], nil, tcell.StyleDefault)
	}
	s.SetContent(x1, y1, chars[2], nil, tcell.StyleDefault)
	s.SetContent(x1, y2, chars[3], nil, tcell.StyleDefault)
	s.SetContent(x2, y1, chars[4], nil, tcell.StyleDefault)
	s.SetContent(x2, y2, chars[5], nil, tcell.StyleDefault)
}
