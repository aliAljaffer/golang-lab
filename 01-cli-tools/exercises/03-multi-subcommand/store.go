// 03-multi-subcommand — a kubectl-shaped CLI in miniature.
//
// You'll build the *core* in this package (Store + actions), and the cobra
// wiring in cmd.go. Tests target the core, since cobra plumbing is dull to test.
//
// Final shape (after you wire cmd.go):
//   tool get pods                        # list all pods
//   tool get pod NAME                    # get one
//   tool create pod NAME --image=...     # create
//   tool delete pod NAME                 # delete
package multicmd

import "errors"

// ErrNotFound is returned when a named resource doesn't exist.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned by Create when the name is taken.
var ErrAlreadyExists = errors.New("already exists")

// Pod is the only resource type for this exercise. (Real kubectl has many.)
type Pod struct {
	Name  string
	Image string
}

// Store holds resources in memory. It is the seam between business logic and CLI plumbing.
// In a real tool, this would be backed by an API client.
type Store struct {
	// TODO: pick a field layout. The tests do per-name operations (Create
	//   must reject duplicates, Get/Delete address one pod by name) — pick
	//   something whose lookup-by-name shape matches that.
}

// NewStore returns an empty, ready-to-use Store.
func NewStore() *Store {
	// TODO: initialize and return.
	return &Store{}
}

// CreatePod adds a new pod. Returns ErrAlreadyExists if Name is taken.
func (s *Store) CreatePod(p Pod) error {
	// TODO
	return errors.New("CreatePod: not implemented")
}

// GetPod returns the pod with the given name, or ErrNotFound.
func (s *Store) GetPod(name string) (Pod, error) {
	// TODO
	return Pod{}, errors.New("GetPod: not implemented")
}

// ListPods returns all pods, sorted by Name ascending (so output is deterministic).
func (s *Store) ListPods() []Pod {
	// TODO
	return nil
}

// DeletePod removes the named pod. Returns ErrNotFound if it doesn't exist.
func (s *Store) DeletePod(name string) error {
	// TODO
	return errors.New("DeletePod: not implemented")
}
