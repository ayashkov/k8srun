package mock

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockOsServices struct {
	args   []string
	env    map[string]string
	stderr *bytes.Buffer
	stdin  *bytes.Buffer
	stdout *bytes.Buffer
}

func NewMockOsServices() *MockOsServices {
	return &MockOsServices{
		args:   []string{"a.out"},
		env:    map[string]string{},
		stderr: new(bytes.Buffer),
		stdin:  new(bytes.Buffer),
		stdout: new(bytes.Buffer),
	}
}

func ExitsWith(t *testing.T, code int, runnable func()) {
	defer func() {
		panic := recover()

		if panic == nil {
			return
		}

		rc := panic.(int)

		assert.Equal(t, code, rc, "expected to exit with %v", code)
	}()

	runnable()
}

func (mock *MockOsServices) Args() []string {
	return mock.args
}

func (mock *MockOsServices) SetArgs(args ...string) {
	mock.args = args
}

func (*MockOsServices) Exit(code int) {
	panic(code)
}

func (mock *MockOsServices) Getenv(key string) string {
	return mock.env[key]
}

func (mock *MockOsServices) Setenv(key string, value string) {
	mock.env[key] = value
}

func (mock *MockOsServices) Stderr() io.Writer {
	return mock.stderr
}

func (mock *MockOsServices) StderrBuffer() *bytes.Buffer {
	return mock.stderr
}

func (mock *MockOsServices) Stdin() io.Reader {
	return mock.stdin
}

func (mock *MockOsServices) StdinBuffer() *bytes.Buffer {
	return mock.stdin
}

func (mock *MockOsServices) Stdout() io.Writer {
	return mock.stdout
}

func (mock *MockOsServices) StdoutBuffer() *bytes.Buffer {
	return mock.stdout
}
