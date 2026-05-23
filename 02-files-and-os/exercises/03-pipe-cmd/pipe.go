// 03-pipe-cmd — implement `cmd1 | cmd2 | cmd3` using os/exec.
//
// You implement: Pipe.
// The tests in pipe_test.go drive the design.
package pipecmd

import "io"

// Pipe runs `cmds` as a pipeline: input -> cmds[0] -> cmds[1] -> ... -> returned stdout.
// Each cmds[i] is [binary, args...]. Returns the final command's stdout, or
// the first non-nil error from Start/Wait/io.
func Pipe(input io.Reader, cmds ...[]string) ([]byte, error) {
	// TODO: build N exec.Cmds and wire stdin↔stdout between consecutive
	//   ones (StdoutPipe is the right API — io.Pipe also works but is
	//   fiddlier). The non-obvious bits the test pins:
	//     - Start each in order (left → right) BEFORE any Wait, otherwise
	//       early commands fill their stdout buffer and deadlock.
	//     - Wait each in order, collecting the first error but still
	//       reaping the rest (avoid zombie processes).
	//     - the LAST command's stdout has to land somewhere you can read
	//       — a bytes.Buffer assigned to cmd[n-1].Stdout is the simplest.
	return nil, nil
}
