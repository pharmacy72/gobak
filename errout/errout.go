package errout

import "bytes"

//A ErrOut it wrapper over stdout/stderr for command.Command
type ErrOut struct {
	err     error
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	Report  bool
	Subject string
}

//New created *ErrOut
func New(e error, report bool, stdout, stderr *bytes.Buffer) *ErrOut {
	return &ErrOut{
		err:    e,
		stderr: stderr,
		stdout: stdout,
		Report: report,
	}
}

//Error it Errorer
func (e *ErrOut) Error() string {
	return e.err.Error()
}

//StdOutput returns a accumulated buffer from stdout
func (e *ErrOut) StdOutput() string {
	return e.stdout.String()
}

//StdErrOutput returns a accumulated buffer from stderr
func (e *ErrOut) StdErrOutput() string {
	return e.stderr.String()
}

//AddSubject change ErrOut Subject if e is *ErrOut
func AddSubject(e error, subject string) error {
	eo, ok := e.(*ErrOut)
	if ok {
		eo.Subject = subject
	}
	return e
}
