// 04-cobra-nested — nested subcommands (kubectl-style).
//
// Goal: build a CLI shaped like:
//   tool get pods
//   tool get nodes
//   tool delete pod NAME
//
// Demonstrates: multi-level AddCommand, args validation (cobra.ExactArgs),
// PersistentFlags inherited by children (e.g. --namespace).
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tool",
		Short: "kubectl-shaped demo",
	}

	// TODO: add a persistent --namespace flag on rootCmd, default "default".
	//   rootCmd.PersistentFlags().String("namespace", "default", "k8s namespace")

	// TODO: build a "get" parent command with no Run — it only groups children.
	//   getCmd := &cobra.Command{ Use: "get", Short: "list resources" }
	// TODO: add "get pods" and "get nodes" subcommands; each prints what it would do.
	//   getCmd.AddCommand(...)
	//   rootCmd.AddCommand(getCmd)

	// TODO: build a "delete" parent and a "delete pod NAME" child using cobra.ExactArgs(1).
	//   deleteCmd := &cobra.Command{ Use: "delete", Short: "delete resources" }
	//   delPodCmd := &cobra.Command{
	//       Use: "pod NAME",
	//       Args: cobra.ExactArgs(1),
	//       RunE: ...,
	//   }

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
