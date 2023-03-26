package main

import (
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setUp(t *testing.T, args ...string) *assert.Assertions {
	logger.Reset()
	mockOs.Setenv("AUTOSERV", "ACE")
	mockOs.Setenv("AUTO_JOB_NAME", "TEST_JOB")
	mockOs.SetArgs(args...)

	return assert.New(t)
}

func Test_MainShowsUsageAndExits_WhenNoArgs(t *testing.T) {
	assert := setUp(t, "k8srun")

	mock.ExitsWith(t, 1, main)

	assert.Contains(mockOs.StderrBuffer().String(),
		"Error: requires at least 1 arg(s), only received 0")
	assert.Contains(mockOs.StdoutBuffer().String(), "Usage:")
	assert.Empty(logger.Entries)
}

func Test_MainLogsErrorAndExits_WhenNoAutoserv(t *testing.T) {
	assert := setUp(t, "k8srun", "template")
	mockOs.Setenv("AUTOSERV", "")

	mock.ExitsWith(t, 1, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.FatalLevel, logger.LastEntry().Level)
	assert.Contains(logger.LastEntry().Message, "AUTOSERV and AUTO_JOB_NAME")
}

func Test_MainLogsErrorAndExits_WhenNoAutoJobName(t *testing.T) {
	assert := setUp(t, "k8srun", "template")
	mockOs.Setenv("AUTO_JOB_NAME", "")

	mock.ExitsWith(t, 1, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.FatalLevel, logger.LastEntry().Level)
	assert.Contains(logger.LastEntry().Message, "AUTOSERV and AUTO_JOB_NAME")
}
