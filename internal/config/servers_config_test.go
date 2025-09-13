package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServersFromFile_GivenJsonFile_ReturnsServerPools(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "servers.json")

	content := `{
		"serverPools": [
			{
				"name": "serverPool",
				"strategy": "rb",
				"servers": [
					{
						"id": "1",
						"name": "server1",
						"protocol": "api",
						"host": "localhost",
						"port": 8080,
						"healthcheckUrl": "/health"
					}
				]
			}
		]
	}`

	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)

	serverPool, err := LoadServersFromFile(filePath, zap.NewNop().Sugar())
	assert.NoError(t, err)
	assert.Len(t, serverPool, 1)
	assert.Equal(t, "serverPool", serverPool[0].Name)
	assert.Equal(t, "rb", serverPool[0].Strategy)
	assert.Equal(t, "serverPool", serverPool[0].Name)
	assert.Len(t, serverPool[0].Servers, 1)
	assert.Equal(t, "1", serverPool[0].Servers[0].ID)
	assert.Equal(t, "server1", serverPool[0].Servers[0].Name)
	assert.Equal(t, "api", serverPool[0].Servers[0].Protocol)
	assert.Equal(t, "localhost", serverPool[0].Servers[0].Host)
	assert.Equal(t, 8080, serverPool[0].Servers[0].Port)
	assert.Equal(t, "/health", serverPool[0].Servers[0].HealthcheckUrl)
}

func TestLoadServersFromFile_GivenYamlFile_ReturnsServerPools(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "servers.yaml")

	content := `
serverPools:
  - name: serverPool
    strategy: rb
    servers:
      - id: "1"
        name: server1
        protocol: api
        host: localhost
        port: 8080
        healthcheckUrl: /health
`

	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)

	serverPool, err := LoadServersFromFile(filePath, zap.NewNop().Sugar())
	assert.NoError(t, err)
	assert.Len(t, serverPool, 1)
	assert.Equal(t, "serverPool", serverPool[0].Name)
	assert.Equal(t, "rb", serverPool[0].Strategy)
	assert.Equal(t, "serverPool", serverPool[0].Name)
	assert.Len(t, serverPool[0].Servers, 1)
	assert.Equal(t, "1", serverPool[0].Servers[0].ID)
	assert.Equal(t, "server1", serverPool[0].Servers[0].Name)
	assert.Equal(t, "api", serverPool[0].Servers[0].Protocol)
	assert.Equal(t, "localhost", serverPool[0].Servers[0].Host)
	assert.Equal(t, 8080, serverPool[0].Servers[0].Port)
	assert.Equal(t, "/health", serverPool[0].Servers[0].HealthcheckUrl)
}

func TestLoadServersFromFile_GivenUnsupportedFileExtension_ReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "servers.txt")
	err := os.WriteFile(filePath, []byte(`{}`), 0644)
	require.NoError(t, err)

	_, err = LoadServersFromFile(filePath, zap.NewNop().Sugar())
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unsupported config format: .txt")
}
