package unimud

import (
	"bufio"
	"fmt"
	"os"
)

// A player represents a user playing the unimud Game instance.
type player struct {
	game   *Game
	input  *bufio.Scanner
	output *bufio.Writer
	login  string
}

func newPlayer(g *Game) *player {
	return &player{
		game:   g,
		input:  bufio.NewScanner(os.Stdin),
		output: bufio.NewWriter(os.Stdout),
	}
}

func (p *player) run() {
	for {
		fmt.Fprint(p.output, "login: ")
		p.output.Flush()

		if p.input.Scan() {
			login := p.input.Text()
			fmt.Fprintln(p.output, "Hi", login)
			p.output.Flush()
		}
	}
}
