package main

import (
	"fmt"
	"os"

	"github.com/beevik/unimud"
	"github.com/spf13/cobra"
)

var (
	console bool
	port    int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Unimud server",
	Long:  "Unimud server launcher",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch a server game instance.",
	Long:  `Launch a server game instance, and listen for clients on a TCP port and/or the console.`,
	Run:   run,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().IntVarP(&port, "port", "p", 0, "Listen on TCP port")
	runCmd.Flags().BoolVarP(&console, "console", "c", false, "Listen on console")
}

func run(cmd *cobra.Command, args []string) {
	if !console && port == 0 {
		fmt.Println("You must specify at least one flag (--console or --port).\n")
		cmd.Usage()
		return
	}
	if port < 0 {
		fmt.Printf("Invalid listening port: %d.\n\n", port)
		cmd.Usage()
		return
	}

	game := unimud.NewGame()
	if console {
		go game.ListenConsole()
	}
	if port > 0 {
		go game.ListenNet(port)
	}
	go game.Run()
	<-game.DoneChan
}
