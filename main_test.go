package main

import (
	"fmt"
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/ayashkov/k8srun/runner"
	"github.com/ayashkov/k8srun/service"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var mockRunnerFactory *mock.MockRunnerFactory

var mockRunner *mock.MockRunner

func setUp(t *testing.T, args ...string) *assert.Assertions {
	prevRunnerFactory := runnerFactory

	t.Cleanup(logger.Reset)
	t.Cleanup(func() {
		runnerFactory = prevRunnerFactory
		mockRunner = nil
		mockRunnerFactory = nil
		mockOs.StderrBuffer().Reset()
		mockOs.StdoutBuffer().Reset()
	})

	ctrl := gomock.NewController(t)
	mockRunnerFactory = mock.NewMockRunnerFactory(ctrl)
	mockRunner = mock.NewMockRunner(ctrl)
	runnerFactory = mockRunnerFactory

	mockOs.Setenv("AUTOSERV", "ACE")
	mockOs.Setenv("AUTO_JOB_NAME", "TEST_JOB")
	mockOs.SetArgs(args...)

	return assert.New(t)
}

func Test_MainShowsUsage_WhenNoArgs(t *testing.T) {
	assert := setUp(t, "k8srun")

	mock.ExitsWith(t, 1, main)

	assert.Contains(mockOs.StderrBuffer().String(),
		"Error: requires at least 1 arg(s), only received 0")
	assert.Contains(mockOs.StdoutBuffer().String(), "Usage:")
	assert.Empty(logger.Entries)
}

func Test_MainLogsError_WhenNoAutoserv(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockOs.Setenv("AUTOSERV", "")

	mock.ExitsWith(t, 1, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.FatalLevel, logger.LastEntry().Level)
	assert.Contains(logger.LastEntry().Message, "AUTOSERV and AUTO_JOB_NAME")
}

func Test_MainLogsError_WhenNoAutoJobName(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockOs.Setenv("AUTO_JOB_NAME", "")

	mock.ExitsWith(t, 1, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.FatalLevel, logger.LastEntry().Level)
	assert.Contains(logger.LastEntry().Message, "AUTOSERV and AUTO_JOB_NAME")
}

func Test_MainRunsJob_Normally(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockRunnerFactory.EXPECT().New("").Return(mockRunner)
	mockRunner.EXPECT().Run(&runner.Job{
		Instance:  "ACE",
		Name:      "TEST_JOB",
		Namespace: "",
		Template:  "template",
		Args:      []string{},
	}, service.Os.Stdout()).Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_MainUsesKubeconfig_WhenKubeconfigFlag(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "--kubeconfig=k8s.conf")

	mockRunnerFactory.EXPECT().New("k8s.conf").Return(mockRunner)
	mockRunner.EXPECT().Run(gomock.Any(), gomock.Any()).Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_MainSuppliesNamespaceToPod_WhenNamespaceFlag(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "--namespace=build")

	mockRunnerFactory.EXPECT().New("").Return(mockRunner)
	mockRunner.EXPECT().Run(&runner.Job{
		Instance:  "ACE",
		Name:      "TEST_JOB",
		Namespace: "build",
		Template:  "template",
		Args:      []string{},
	}, service.Os.Stdout()).Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_MainSuppliesNamespaceToPod_WhenNFlag(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "-n", "dev")

	mockRunnerFactory.EXPECT().New("").Return(mockRunner)
	mockRunner.EXPECT().Run(&runner.Job{
		Instance:  "ACE",
		Name:      "TEST_JOB",
		Namespace: "dev",
		Template:  "template",
		Args:      []string{},
	}, service.Os.Stdout()).Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_MainSuppliesArgsToContainer_WhenProvided(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "--", "ls", "-la", "/")

	mockRunnerFactory.EXPECT().New("").Return(mockRunner)
	mockRunner.EXPECT().Run(&runner.Job{
		Instance:  "ACE",
		Name:      "TEST_JOB",
		Namespace: "",
		Template:  "template",
		Args:      []string{"ls", "-la", "/"},
	}, service.Os.Stdout()).Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_MainUsesContainerExitCode_WhenProvided(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockRunnerFactory.EXPECT().New("").Return(mockRunner)
	mockRunner.EXPECT().Run(gomock.Any(), gomock.Any()).Return(42, nil)

	mock.ExitsWith(t, 42, main)

	assert.Empty(logger.Entries)
}

func Test_MainLogsError_WhenRunReportsError(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockRunnerFactory.EXPECT().New("").Return(mockRunner)
	mockRunner.EXPECT().Run(gomock.Any(), gomock.Any()).
		Return(-1, fmt.Errorf("error running"))

	mock.ExitsWith(t, 128, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.ErrorLevel, logger.LastEntry().Level)
	assert.Equal(logger.LastEntry().Message, "error running")
}
