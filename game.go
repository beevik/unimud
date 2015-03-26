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
	DoneChan      chan bool          // used to signal that the game's Run goroutine has ended
	shutdownChan  chan bool          // used to signal that the game should shut down
	yieldChan     chan bool          // used to yield control to game Run goroutine
	resumeReqChan chan chan bool     // used to request resumption of control by another goroutine
	rooms         map[int]*room      // all loaded rooms
	players       []*player          // all connected players
	playerMap     map[string]*player // all players who have entered the game world
	listeners     []net.Listener     // tracks all known network listeners
	listenersLock sync.Mutex         // protects the listeners slice
}

// NewGame creates a new unimud game instance.
func NewGame() *Game {
	return &Game{
		DoneChan:      make(chan bool),
		shutdownChan:  make(chan bool),
		yieldChan:     make(chan bool),
		resumeReqChan: make(chan chan bool),
		rooms:         make(map[int]*room),
		playerMap:     make(map[string]*player),
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
	// Start listening on the requested TCP port.
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// Track the listener.
	g.listenerAdd(l)

	// Start an infinite loop that waits for new client
	// connections and creates players and associated goroutines
	// as connections arrive.
	for {
		// Block while waiting for a client connection.
		c, err := l.Accept()
		if err != nil {
			log.Printf("Listen on port %d ended.", port)
			break
		}

		// Create a new player on the accepted connection.
		p := newPlayer(g, newConnNet(c))
		go p.run()
	}

	g.listenerRemove(l)
}

// Run starts the game loop.
func (g *Game) Run() {
	// Create a channel that receives the current time once per second
	clock := time.Tick(1 * time.Second)

mainLoop:
	for {
		select {

		// Wait for shutdown signal
		case <-g.shutdownChan:
			g.onShutdown()
			break mainLoop

		// Another object is requesting resumption of control
		// from this goroutine. The channel on which resumption
		// should be communicated is passed as data through the
		// channel.
		case ch := <-g.resumeReqChan:
			ch <- true    // issue a resume signal through the passed channel.
			<-g.yieldChan // block until the resumed goroutine yields again.

		// The clock channel ticks sends the current time once per
		// second
		case t := <-clock:
			// TODO: Handle timed events here.
			_ = t
		}
	}

	// Signal on the Done channel that the game has ended.
	g.DoneChan <- true
}

// Shutdown sends a signal to the game to shut itself down.
func (g *Game) Shutdown() {
	g.shutdownChan <- true
}

// onShutdown is called when the game's Run goroutine processes
// the shutdown request.
func (g *Game) onShutdown() {
	g.removeAllListeners()
	for _, p := range g.players {
		p.leaveGame()
	}
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

// Have the player "enter" the game world.
func (g *Game) playerEnter(p *player) {
	g.playerMap[p.login] = p
	g.broadcast(fmt.Sprintf("%s entered the game.\n", p.login))
	p.entered = true
}

// Have the player "leave" the game world.
func (g *Game) playerLeave(p *player) {
	p.entered = false
	delete(g.playerMap, p.login)
	g.broadcast(fmt.Sprintf("%s left the game.\n", p.login))
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

// Close all active network listeners.
func (g *Game) removeAllListeners() {
	g.listenersLock.Lock()
	defer g.listenersLock.Unlock()
	for _, l := range g.listeners {
		l.Close()
	}
	g.listeners = nil
}

// Broadcast a message to all players who have entered the game.
func (g *Game) broadcast(s string) {
	for _, p := range g.players {
		if p.entered {
			p.Print(s)
		}
	}
}

// Look up the room in the game's room map. If it's not there,
// load it from disk and add it to the room map.
func (g *Game) roomGet(id int) (*room, error) {
	if r, ok := g.rooms[id]; ok {
		return r, nil
	}

	r, err := roomLoad(g, id)
	if r != nil {
		g.rooms[id] = r
	}
	return r, err
}
