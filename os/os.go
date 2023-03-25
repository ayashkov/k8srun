package os

import (
	"os"
)

type OsServices interface {
	Exit(code int)
	Getenv(key string) string
	Stdout() *os.File
}

type DefaultOsServices struct{}

var Os OsServices = DefaultOsServices{}

func (DefaultOsServices) Exit(code int) {
	os.Exit(code)
}

func (DefaultOsServices) Getenv(key string) string {
	return os.Getenv(key)
}

func (DefaultOsServices) Stdout() *os.File {
	return os.Stdout
}
