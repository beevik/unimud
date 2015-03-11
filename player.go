package unimud

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path"
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

type playerState func(p *player) playerState

func (p *player) run() {
	state := (*player).stateLogin
	for state != nil {
		state = state(p)
	}
}

func (p *player) Print(args ...interface{}) {
	fmt.Fprint(p.output, args...)
	p.output.Flush()
}

func (p *player) Println(args ...interface{}) {
	fmt.Fprintln(p.output, args...)
	p.output.Flush()
}

func (p *player) GetLine() (line string, err error) {
	p.output.Flush()
	if p.input.Scan() {
		line := p.input.Text()
		return line, nil
	}
	return "", errors.New("player: disconnected")
}

func (p *player) stateLogin() playerState {
	p.Print("login: ")
	login, err := p.GetLine()
	if err != nil {
		return nil
	}

	switch {
	case len(login) < 4:
		p.Println("login id is too short.")
		return (*player).stateLogin
	case len(login) > 32:
		p.Println("login id is too long.")
		return (*player).stateLogin
	case !validateLogin(login):
		p.Println("login id contains invalid characters.")
		return (*player).stateLogin
	}

	p.login = login
	if err := p.load(); err != nil {
		// TODO: go to create new player state
		return nil
	}

	p.Print("password: ")
	pw, err := p.GetLine()
	if err != nil {
		return nil
	}

	_ = pw

	// TODO: Check the password
	// TODO: If valid, return the playing-game state func
	return nil
}

func validateLogin(login string) bool {
	for _, c := range login {
		switch {
		case c >= 'a' && c <= 'z':
			continue
		case c >= 'A' && c <= 'Z':
			continue
		}
		return false
	}
	return true
}

func (p *player) load() error {
	filename := path.Join("players", p.login+".dat")
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	var login string
	if err := dec.Decode(&login); err != nil {
		return err
	}
	if login != p.login {
		return errors.New("player: login id mismatch")
	}

	// TODO: Decode the rest of the player's data

	return nil
}
