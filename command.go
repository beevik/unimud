package unimud

import (
	"errors"
	"log"
	"strings"

	"github.com/beevik/prefixtree"
)

type handlerFunc func(p *player, arg string) error

type command struct {
	str     string
	handler handlerFunc
}

var commandList = []command{
	{"e", (*player).cmdEast},
	{"east", (*player).cmdEast},
	{"go", (*player).cmdGo},
	{"look", (*player).cmdLook},
	{"n", (*player).cmdNorth},
	{"north", (*player).cmdNorth},
	{"quit", (*player).cmdQuit},
	{"reply", (*player).cmdReply},
	{"s", (*player).cmdSouth},
	{"say", (*player).cmdSay},
	{"shutdown", (*player).cmdShutdown},
	{"south", (*player).cmdSouth},
	{"tell", (*player).cmdTell},
	{"w", (*player).cmdWest},
	{"west", (*player).cmdWest},
	{"whisper", (*player).cmdTell},
	{"who", (*player).cmdWho},
	{"yell", (*player).cmdYell},
}

var commands = prefixtree.New()

// Build the prefix tree containing all commands.
func init() {
	for _, c := range commandList {
		commands.Add(c.str, c.handler)
	}
}

func (p *player) cmdEast(arg string) error {
	return p.cmdGo("east")
}

func (p *player) cmdGo(arg string) error {
	if len(arg) > 0 {
		for _, exit := range p.room.Exits {
			if exit.Name == arg {
				newRoom, err := p.game.roomGet(exit.ID)
				if err != nil {
					log.Printf("Room %d failed to load: %v\n", exit.ID, err)
					break
				}
				p.room.playerLeave(p)
				newRoom.playerEnter(p)
				newRoom.display(p)
				return nil
			}
		}
	}
	p.Println("You can't go that direction.")
	return nil
}

func (p *player) cmdLook(arg string) error {
	p.room.display(p)
	return nil
}

func (p *player) cmdNorth(arg string) error {
	return p.cmdGo("north")
}

func (p *player) cmdQuit(arg string) error {
	p.Println("Quitting the game.")
	return errors.New("player: disconnecting")
}

func (p *player) cmdReply(arg string) error {
	name, ok := p.properties["replyto"].(string)
	if !ok || name == "" {
		p.Println("No one has whispered to you.")
		return nil
	}

	return p.cmdTell(name + " " + arg)
}

func (p *player) cmdSay(arg string) error {
	if arg == "" {
		p.Println("Syntax: say <message>")
		return nil
	}

	if len(p.room.players) == 1 {
		p.Println("No one hears you.")
		return nil
	}

	p.Printf("You say, '%s'.\n", arg)
	for _, op := range p.room.players {
		if p != op {
			op.Printf("%s says, '%s'.\n", p.login, arg)
		}
	}
	return nil
}

func (p *player) cmdShutdown(arg string) error {
	// Yield control back to game goroutine
	// just before shutting it down. Otherwise
	// it won't be able to process the channel
	// update.
	p.yield()
	p.game.Shutdown()
	p.resume()
	return nil
}

func (p *player) cmdSouth(arg string) error {
	return p.cmdGo("south")
}

func (p *player) cmdTell(arg string) error {
	split := strings.SplitN(arg, " ", 2)
	if len(split) < 2 {
		p.Println("Syntax: tell <name> <message>")
		return nil
	}

	op := p.game.playerMap[split[0]]
	switch {
	case op == nil:
		p.Println("Player", split[0], "not logged in.")
	case op == p:
		p.Println("See a psychiatrist.")
	default:
		p.Printf("Message sent to %s.\n", split[0])
		op.Printf("%s whispers, '%s'.\n", p.login, stripLeadingWhitespace(split[1]))
		op.properties["replyto"] = p.login
	}
	return nil
}

func (p *player) cmdWest(arg string) error {
	return p.cmdGo("west")
}

func (p *player) cmdWho(arg string) error {
	for login := range p.game.playerMap {
		p.Println(login)
	}
	return nil
}

func (p *player) cmdYell(arg string) error {
	for _, op := range p.game.playerMap {
		if p == op {
			op.Printf("You yelled, '%s'.\n", arg)
		} else {
			op.Printf("%s yelled, '%s'.\n", p.login, arg)
		}
	}
	return nil
}
