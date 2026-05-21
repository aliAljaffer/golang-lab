package multicmd

import "github.com/spf13/cobra"

// NewRootCmd builds the cobra command tree, wired against the given Store.
// Passing the Store in (rather than newing one inside) is what makes the cmd
// testable from a unit test if you want to go that far.
//
// Shape to build:
//
//   root: "tool"
//     get:
//       pods                 -> list, print one Name per line
//       pod NAME             -> print "NAME IMAGE"
//     create:
//       pod NAME --image=X   -> Store.CreatePod
//     delete:
//       pod NAME             -> Store.DeletePod
func NewRootCmd(s *Store) *cobra.Command {
	root := &cobra.Command{
		Use:   "tool",
		Short: "kubectl-shaped demo",
	}

	// TODO: build `get` parent + `pods` and `pod NAME` children.
	// TODO: build `create` parent + `pod NAME --image=X` child.
	// TODO: build `delete` parent + `pod NAME` child.
	// Tip: cobra.ExactArgs(1) for the NAME positional.

	_ = s
	return root
}
