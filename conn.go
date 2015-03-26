package unimud

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

// A conn represents a connection from a player to the game.
type conn struct {
	closer io.Closer
	input  *bufio.Scanner
	output *bufio.Writer
}

type nopCloser int

func (c *nopCloser) Close() error {
	// Do nothing
	return nil
}

// newConnConsole creates a new connection using the standard I/O
// as game input and output.
func newConnConsole() *conn {
	return &conn{
		closer: new(nopCloser),
		input:  bufio.NewScanner(os.Stdin),
		output: bufio.NewWriter(os.Stdout),
	}
}

// newConnNet creates a new connection using the network connection
// `nc` for the input and output
func newConnNet(nc net.Conn) *conn {
	return &conn{
		closer: nc,
		input:  bufio.NewScanner(nc),
		output: bufio.NewWriter(nc),
	}
}

// Close the connection.
func (c *conn) Close() error {
	return c.closer.Close()
}

// Flush the output on the connection.
func (c *conn) Flush() {
	c.output.Flush()
}

// Print outputs arguments to the player's output writer
// without appending a trailing carriage return.
func (c *conn) Print(args ...interface{}) {
	fmt.Fprint(c.output, args...)
	c.Flush()
}

// Println outputs arguments to the player's output writer
// and appends a trailing carriage return.
func (c *conn) Println(args ...interface{}) {
	fmt.Fprintln(c.output, args...)
	c.Flush()
}

// Printf outputs a printf-formatted string to the player's
// output writer.
func (c *conn) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.output, format, args...)
	c.Flush()
}
