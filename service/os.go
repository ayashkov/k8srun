package service

import (
	"os"
)

type OsServices interface {
	Exit(code int)
	Getenv(key string) string
	Stdout() *os.File
}

type defaultOsServices struct{}

var Os OsServices = defaultOsServices{}

func (defaultOsServices) Exit(code int) {
	os.Exit(code)
}

func (defaultOsServices) Getenv(key string) string {
	return os.Getenv(key)
}

func (defaultOsServices) Stdout() *os.File {
	return os.Stdout
}
