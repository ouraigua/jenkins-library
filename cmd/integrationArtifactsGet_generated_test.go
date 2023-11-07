//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationArtifactsGetCommand(t *testing.T) {
	t.Parallel()

	testCmd := IntegrationArtifactsGetCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "integrationArtifactsGet", testCmd.Use, "command name incorrect")

}
