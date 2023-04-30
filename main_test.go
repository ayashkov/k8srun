package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/ayashkov/k8srun/runner"
	"github.com/ayashkov/k8srun/service"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var logger *test.Hook

var mockOs *mock.MockOsServices

var mockRunnerFactory *mock.MockRunnerFactory

var mockRunner *mock.MockRunner

func TestMain(m *testing.M) {
	mockOs = mock.NewMockOsServices()
	service.Os = mockOs
	service.Log, logger = test.NewNullLogger()
	service.Log.ExitFunc = mockOs.Exit
	m.Run()
}

func setUp(t *testing.T, args ...string) *assert.Assertions {
	prevRunnerFactory := runnerFactory

	t.Cleanup(func() {
		runnerFactory = prevRunnerFactory
		mockRunner = nil
		mockRunnerFactory = nil
		mockOs.StderrBuffer().Reset()
		mockOs.StdoutBuffer().Reset()
		logger.Reset()
	})

	mockOs.Setenv("AUTOSERV", "ACE")
	mockOs.Setenv("AUTO_JOB_NAME", "TEST_JOB")
	mockOs.SetArgs(args...)

	ctrl := gomock.NewController(t)

	mockRunner = mock.NewMockRunner(ctrl)
	mockRunnerFactory = mock.NewMockRunnerFactory(ctrl)
	runnerFactory = mockRunnerFactory

	return assert.New(t)
}

func Test_Main_ShowsUsage_WhenNoArgs(t *testing.T) {
	assert := setUp(t, "k8srun")

	mock.ExitsWith(t, 1, main)

	assert.Contains(mockOs.StderrBuffer().String(),
		"Error: requires at least 1 arg(s), only received 0")
	assert.Contains(mockOs.StdoutBuffer().String(), "Usage:")
	assert.Empty(logger.Entries)
}

func Test_Main_LogsError_WhenNoAutoserv(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockOs.Setenv("AUTOSERV", "")

	mock.ExitsWith(t, 1, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.FatalLevel, logger.LastEntry().Level)
	assert.Contains(logger.LastEntry().Message, "AUTOSERV and AUTO_JOB_NAME")
}

func Test_Main_LogsError_WhenNoAutoJobName(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockOs.Setenv("AUTO_JOB_NAME", "")

	mock.ExitsWith(t, 1, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.FatalLevel, logger.LastEntry().Level)
	assert.Contains(logger.LastEntry().Message, "AUTOSERV and AUTO_JOB_NAME")
}

func Test_Main_RunsJob_Normally(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockRunnerFactory.EXPECT().
		New("").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(context.Background(),
			&runner.Job{
				Instance:  "ACE",
				Name:      "TEST_JOB",
				Namespace: "",
				Template:  "template",
				Args:      []string{},
			}, service.Os.Stdout()).
		Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_Main_UsesKubeconfig_WhenKubeconfigFlag(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "--kubeconfig=k8s.conf")

	mockRunnerFactory.EXPECT().
		New("k8s.conf").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_Main_SuppliesNamespaceToPod_WhenNamespaceFlag(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "--namespace=build")

	mockRunnerFactory.EXPECT().
		New("").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(context.Background(),
			&runner.Job{
				Instance:  "ACE",
				Name:      "TEST_JOB",
				Namespace: "build",
				Template:  "template",
				Args:      []string{},
			}, service.Os.Stdout()).
		Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_Main_SuppliesNamespaceToPod_WhenNFlag(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "-n", "dev")

	mockRunnerFactory.EXPECT().
		New("").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(context.Background(),
			&runner.Job{
				Instance:  "ACE",
				Name:      "TEST_JOB",
				Namespace: "dev",
				Template:  "template",
				Args:      []string{},
			}, service.Os.Stdout()).
		Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_Main_SuppliesArgsToContainer_WhenProvided(t *testing.T) {
	assert := setUp(t, "k8srun", "template", "--", "ls", "-la", "/")

	mockRunnerFactory.EXPECT().
		New("").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(context.Background(),
			&runner.Job{
				Instance:  "ACE",
				Name:      "TEST_JOB",
				Namespace: "",
				Template:  "template",
				Args:      []string{"ls", "-la", "/"},
			}, service.Os.Stdout()).
		Return(0, nil)

	mock.ExitsWith(t, 0, main)

	assert.Empty(logger.Entries)
}

func Test_Main_UsesContainerExitCode_WhenProvided(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockRunnerFactory.EXPECT().
		New("").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(42, nil)

	mock.ExitsWith(t, 42, main)

	assert.Empty(logger.Entries)
}

func Test_Main_LogsError_WhenRunReportsError(t *testing.T) {
	assert := setUp(t, "k8srun", "template")

	mockRunnerFactory.EXPECT().
		New("").
		Return(mockRunner)
	mockRunner.EXPECT().
		Run(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(-1, fmt.Errorf("error running"))

	mock.ExitsWith(t, 128, main)

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.ErrorLevel, logger.LastEntry().Level)
	assert.Equal(logger.LastEntry().Message, "error running")
}
