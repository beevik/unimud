package unimud

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// A Game is an instance of a uniMUD game.
type Game struct {
	DoneChan      chan bool
	yieldChan     chan bool
	resumeReqChan chan chan bool
	players       []*player      // all connected players
	listeners     []net.Listener // tracks all known network listeners
	listenersLock sync.Mutex     // protects the listeners slice
}

// NewGame creates a new unimud game instance.
func NewGame() *Game {
	return &Game{
		DoneChan:      make(chan bool),
		yieldChan:     make(chan bool),
		resumeReqChan: make(chan chan bool),
	}
}

// ListenConsole listens for player input on standard input
// and outputs text on standard output.
func (g *Game) ListenConsole() {
	for {
		p := newPlayer(g, newConnConsole())
		p.run()
	}
}

// ListenNet begins listening for new player connections on
// the specified port.
func (g *Game) ListenNet(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Track the listener
	g.listenerAdd(l)

	for {
		// Wait for a connection.
		c, err := l.Accept()
		if err != nil {
			break
		}

		// Create a new player on the accepted connection.
		p := newPlayer(g, newConnNet(c))
		go p.run()
	}

	g.listenerRemove(l)
}

// Add a connected player to the game's list of players.
func (g *Game) playerAdd(p *player) {
	g.players = append(g.players, p)
}

// Remove a connected player from the game's list of players.
func (g *Game) playerRemove(p *player) {
	for i, op := range g.players {
		if op == p {
			g.players = append(g.players[:i], g.players[i+1:]...)
			break
		}
	}
}

// Add a listener to the game's list of listeners.
func (g *Game) listenerAdd(l net.Listener) {
	g.listenersLock.Lock()
	defer g.listenersLock.Unlock()
	g.listeners = append(g.listeners, l)
}

// Remove a listener from the game's list of listeners.
func (g *Game) listenerRemove(l net.Listener) {
	g.listenersLock.Lock()
	defer g.listenersLock.Unlock()
	for i, ol := range g.listeners {
		if ol == l {
			g.listeners = append(g.listeners[:i], g.listeners[i+1:]...)
			break
		}
	}
}

// Run starts the game loop.
func (g *Game) Run() {
	clock := time.Tick(1 * time.Second)
	for {
		select {
		case ch := <-g.resumeReqChan:
			ch <- true
			<-g.yieldChan

		case t := <-clock:
			log.Println(t)
		}
	}
	g.DoneChan <- true
}
