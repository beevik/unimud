package unimud

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

// A room represents a location in the MUD. Each room contains
// zero or more active players.
type room struct {
	ID          int
	Name        string
	Description string
	Exits       []exit
	game        *Game
	players     []*player
}

type exit struct {
	Name string
	ID   int
}

// Load a room with the requested ID from disk. Associate it with
// the game `g`.
func roomLoad(g *Game, ID int) (*room, error) {
	filename := path.Join("rooms", fmt.Sprintf("%d.dat", ID))
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Use json to decode the room's data.
	dec := json.NewDecoder(f)

	r := &room{game: g}
	if err := dec.Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

// Display the room's description to the player `p`.
func (r *room) display(p *player) {
	p.Println(r.Name)
	p.Println(r.Description)

	var exits []string
	for _, e := range r.Exits {
		exits = append(exits, e.Name)
	}

	exitString := strings.Join(exits, ", ")
	p.Printf("Exits: %s\n", exitString)

	for _, op := range r.players {
		if p != op {
			p.Printf("%s is standing here.\n", op.login)
		}
	}
}

// Have the player enter the room.
func (r *room) playerEnter(p *player) {
	r.Println(p.login, "entered the room.")
	r.players = append(r.players, p)
	p.room = r
	p.properties["room"] = r.ID
}

// Have the player leave the room.
func (r *room) playerLeave(p *player) {
	for i, rp := range r.players {
		if rp == p {
			r.players = append(r.players[:i], r.players[i+1:]...)
			p.room = nil
			r.Println(p.login, "left the room.")
			break
		}
	}
}

// Print a message to all players in the room.
func (r *room) Print(args ...interface{}) {
	for _, p := range r.players {
		p.Print(args...)
	}
}

// Print a CR-terminated message to all players in the room.
func (r *room) Println(args ...interface{}) {
	for _, p := range r.players {
		p.Println(args...)
	}
}

// Print a formatted message to all players in the room.
func (r *room) Printf(format string, args ...interface{}) {
	for _, p := range r.players {
		p.Printf(format, args...)
	}
}
