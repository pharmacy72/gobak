package command

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

//A Command wrapper over os/exec.Cmd
type Command struct {
	verbose bool
	cmd     *exec.Cmd
	Error   error
	Stdout  *bufferWrap
	Stderr  *bufferWrap
}

type bufferWrap struct {
	Buffer  *bytes.Buffer
	verBuf  io.Writer
	verbose bool
}

func (b *bufferWrap) Write(p []byte) (n int, err error) {
	if b.verbose {
		n, err := b.verBuf.Write(p)
		if err != nil {
			return n, err
		}
	}
	return b.Buffer.Write(p)
}

func newBuffer(verbose bool, verBuf io.Writer) *bufferWrap {
	return &bufferWrap{
		Buffer:  &bytes.Buffer{},
		verbose: verbose,
		verBuf:  verBuf,
	}
}

// Exec starts the specified command and waits for it to complete.
// If the command fails to run or doesn't complete successfully, the
// error contains a *Command.Error
func Exec(verbose bool, name string, args ...string) *Command {
	w := &Command{
		verbose: verbose,
		cmd:     exec.Command(name, args[:]...),
	}
	w.Stdout = newBuffer(verbose, os.Stdout)
	w.Stderr = newBuffer(verbose, os.Stderr)
	w.cmd.Stdout = io.Writer(w.Stdout)
	w.cmd.Stderr = io.Writer(w.Stderr)
	return w.run()
}

func (c *Command) run() *Command {
	c.Error = c.cmd.Run()
	return c
}
