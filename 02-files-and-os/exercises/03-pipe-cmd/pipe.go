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
	// TODO: validate cmds is non-empty.
	// TODO: construct exec.Cmd for each entry: exec.Command(c[0], c[1:]...).
	// TODO: cmd[0].Stdin = input
	// TODO: for i in 1..n-1: cmd[i].Stdin, _ = cmd[i-1].StdoutPipe()
	// TODO: var out bytes.Buffer; cmd[n-1].Stdout = &out
	// TODO: Start each cmd in order (cmd[0] first).
	// TODO: Wait each cmd in order. Collect the first error.
	// TODO: return out.Bytes(), firstErr.
	return nil, nil
}
