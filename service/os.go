package service

import (
	"io"
	"os"
)

type OsServices interface {
	Args() []string
	Exit(code int)
	Getenv(key string) string
	Stderr() io.Writer
	Stdin() io.Reader
	Stdout() io.Writer
}

type defaultOsServices struct{}

var Os OsServices = defaultOsServices{}

func (defaultOsServices) Args() []string {
	return os.Args
}

func (defaultOsServices) Exit(code int) {
	os.Exit(code)
}

func (defaultOsServices) Getenv(key string) string {
	return os.Getenv(key)
}

func (defaultOsServices) Stderr() io.Writer {
	return os.Stderr
}

func (defaultOsServices) Stdin() io.Reader {
	return os.Stdin
}

func (defaultOsServices) Stdout() io.Writer {
	return os.Stdout
}
