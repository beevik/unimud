package unimud

import (
	"encoding/gob"
	"errors"
	"os"
	"path"
)

// A player represents a user playing the unimud Game instance.
type player struct {
	*conn
	game       *Game
	resumeChan chan bool
	login      string
	properties map[string]string // all known player properties
}

// Create a new player associated with the Game g.
func newPlayer(g *Game, c *conn) *player {
	return &player{
		conn:       c,
		game:       g,
		resumeChan: make(chan bool),
		properties: make(map[string]string),
	}
}

// playerState is a function type that operates on a player
// and returns the next state the player should enter.
type playerState func(p *player) playerState

// run launches the player's state machine.
func (p *player) run() {
	p.resume()
	defer p.yield()

	p.game.playerAdd(p)

	state := (*player).stateLogin
	for state != nil {
		state = state(p)
	}

	p.game.playerRemove(p)
}

// GetLine reads a CR-terminated line of text from the
// player's input reader. It returns an error if the read
// fails.
func (p *player) GetLine() (line string, err error) {
	p.yield()
	defer p.resume()

	p.Flush()
	if p.conn.input.Scan() {
		line = p.conn.input.Text()
		return line, nil
	}

	return "", p.conn.input.Err()
}

// yield control of the Game's state back to the game's Run
// goroutine.
func (p *player) yield() {
	p.game.yieldChan <- true
}

// resume control of the Game's state by requesting it from
// the Game's Run goroutine.
func (p *player) resume() {
	// On the game's "resume request" channel, send the channel
	// on which to report resumption.
	p.game.resumeReqChan <- p.resumeChan

	// Wait until the game's Run goroutine signals resumption
	// on our resume channel.
	<-p.resumeChan
}

// stateLogin handles player I/O when the player is in
// the login state.
func (p *player) stateLogin() playerState {
	p.Print("login: ")
	login, err := p.GetLine()
	if err != nil {
		return nil
	}

	// Check for invalid login id
	switch {
	case len(login) == 0:
		return (*player).stateLogin
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

	// Attempt to load the player from disk (or db).
	p.login = login
	if err := p.load(); err != nil {
		return (*player).stateCreateNew
	}

	// Request the password
	p.Print("password: ")
	pw, err := p.GetLine()
	if err != nil {
		return nil
	}

	// Check the password
	if pw != p.properties["pw"] {
		p.Println("incorrect password.")
		return (*player).stateLogin
	}

	return (*player).stateEnteringGame
}

// stateCreateNew handles the creation of new player
// account data.
func (p *player) stateCreateNew() playerState {
	p.Print("enter password: ")
	pw, err := p.GetLine()
	if err != nil {
		return nil
	}

	if len(pw) < 4 {
		p.Println("password too short.")
		return (*player).stateLogin
	}

	p.Print("re-enter password: ")
	rpw, err := p.GetLine()
	if err != nil {
		return nil
	}

	if pw != rpw {
		p.Println("passwords don't match.")
		return (*player).stateLogin
	}

	p.properties["pw"] = pw

	if err := p.save(); err != nil {
		p.Println("error: player couldn't be saved.")
		return nil
	}

	return (*player).stateEnteringGame
}

// stateEnteringGame is a transitional state that is
// entered just before the player starts playing the game.
func (p *player) stateEnteringGame() playerState {
	return (*player).statePlaying
}

// statePlaying is the state a player enters while playing
// the game itself.
func (p *player) statePlaying() playerState {
	p.Print("> ")
	line, err := p.GetLine()
	if err != nil {
		return nil
	}

	if line == "quit" {
		return nil
	}

	return (*player).statePlaying
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

// save stores the player's data to disk and returns an
// error if the save fails.
func (p *player) save() error {
	filename := path.Join("players", p.login+".dat")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)

	if err := enc.Encode(p.login); err != nil {
		return err
	}
	if err := enc.Encode(p.properties); err != nil {
		return err
	}

	return nil
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

	if err := dec.Decode(&p.properties); err != nil {
		return err
	}

	return nil
}
