package main

import (
	"fmt"
	"os"
)

type workload struct {
	image   string
	command string
}

func main() {
	workload, err := configure()

	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}

	fmt.Println(workload.image)
	fmt.Println(workload.command)
}

func configure() (*workload, error) {
	argc := len(os.Args)

	if argc < 2 {
		return nil, fmt.Errorf("usage: %v image [command]", os.Args[0])
	}

	workload := workload{
		image:   os.Args[1],
		command: "",
	}

	if argc > 2 {
		workload.command = os.Args[2]
	}

	return &workload, nil
}
