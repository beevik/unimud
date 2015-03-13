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

// Create a new player associated with the Game g.
func newPlayer(g *Game) *player {
	return &player{
		game:   g,
		input:  bufio.NewScanner(os.Stdin),
		output: bufio.NewWriter(os.Stdout),
	}
}

// playerState is a function type that operates on a player
// and returns the next state the player should enter.
type playerState func(p *player) playerState

// run launches the player's state machine.
func (p *player) run() {
	state := (*player).stateLogin
	for state != nil {
		state = state(p)
	}
}

// Print outputs arguments to the player's output writer
// without appending a trailing carriage return.
func (p *player) Print(args ...interface{}) {
	fmt.Fprint(p.output, args...)
	p.output.Flush()
}

// Println outputs arguments to the player's output writer
// and appends a trailing carriage return.
func (p *player) Println(args ...interface{}) {
	fmt.Fprintln(p.output, args...)
	p.output.Flush()
}

// GetLine reads a CR-terminated line of text from the
// player's input reader. It returns an error if the read
// fails.
func (p *player) GetLine() (line string, err error) {
	p.output.Flush()
	if p.input.Scan() {
		line = p.input.Text()
		return line, nil
	}
	return "", errors.New("player: disconnected")
}

// stateLogin handles player I/O when the player is in
// the login state.
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

// validateLogin checks a login id string for invalid
// characters and returns true if validation succeeds.
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

// load reads the player's data from disk and returns
// an error if the load fails.
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
