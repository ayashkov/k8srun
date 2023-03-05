package main

import (
	"fmt"
	"os"
	"strings"
)

type workload struct {
	namespace string
	image     string
	command   string
}

func main() {
	workload, err := configure()

	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}

	fmt.Println(workload.namespace)
	fmt.Println(workload.image)
	fmt.Println(workload.command)
}

func configure() (*workload, error) {
	argc := len(os.Args)

	if argc < 2 {
		return nil, fmt.Errorf("usage: %v image [command]", os.Args[0])
	}

	workload := workload{
		namespace: normalize(os.Getenv("AUTO_JOB_NAME")),
		image:     os.Args[1],
		command:   "",
	}

	if argc > 2 {
		workload.command = os.Args[2]
	}

	return &workload, nil
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s)), "_", "-")
}
