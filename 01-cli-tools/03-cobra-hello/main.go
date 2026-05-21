// 03-cobra-hello — smallest cobra app.
//
// Goal: build a CLI named "hello" with one subcommand "greet":
//   go run . greet --name Ali
//   -> "Hello, Ali!"
//
// Demonstrates: cobra.Command, PersistentFlags vs Flags, Execute(), Run vs RunE.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// TODO: build the root command.
	//   rootCmd := &cobra.Command{
	//     Use:   "hello",
	//     Short: "a tiny cobra demo",
	//   }

	// TODO: build a "greet" subcommand with a --name string flag.
	//   greetCmd := &cobra.Command{
	//     Use:   "greet",
	//     Short: "say hello to someone",
	//     RunE: func(cmd *cobra.Command, args []string) error { ... },
	//   }
	//   greetCmd.Flags().String("name", "world", "who to greet")

	// TODO: rootCmd.AddCommand(greetCmd)

	// TODO: if err := rootCmd.Execute(); err != nil { os.Exit(1) }

	_ = cobra.Command{}
	_ = fmt.Println
	_ = os.Exit
}
