package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type integrationArtifactsGetMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newIntegrationArtifactsGetTestsUtils() integrationArtifactsGetMockUtils {
	utils := integrationArtifactsGetMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunIntegrationArtifactsGet(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := integrationArtifactsGetOptions{}

		utils := newIntegrationArtifactsGetTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runIntegrationArtifactsGet(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := integrationArtifactsGetOptions{}

		utils := newIntegrationArtifactsGetTestsUtils()

		// test
		err := runIntegrationArtifactsGet(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
