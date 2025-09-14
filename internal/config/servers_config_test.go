package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
)

func resetFlags(t *testing.T, args ...string) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(os.Stdout)
	if len(args) == 0 {
		args = []string{os.Args[0]}
	}
	os.Args = args
}

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func Test_loadLBConfig_WhenConfigInJson_ReturnsExpectedParsedConfiguration(t *testing.T) {
	path := writeTempFile(t, "servers.json", `{
		"serverPools": [
			{
				"name": "serverPool",
				"strategy": "rb",
				"servers": [
					{"id":"1","name":"server1","protocol":"api","host":"localhost","port":8080,"healthcheckUrl":"/health"}
				]
			}
		]
	}`)
	cfg, err := loadLBConfig(path, zap.NewNop().Sugar())
	require.NoError(t, err)
	require.Len(t, cfg.ServerPools, 1)
	sp := cfg.ServerPools[0]
	assert.Equal(t, "serverPool", sp.Name)
	assert.Equal(t, "rb", sp.Strategy)
	require.Len(t, sp.Servers, 1)
	s := sp.Servers[0]
	assert.Equal(t, "1", s.ID)
	assert.Equal(t, "server1", s.Name)
	assert.Equal(t, "api", s.Protocol)
	assert.Equal(t, "localhost", s.Host)
	assert.Equal(t, 8080, s.Port)
	assert.Equal(t, "/health", s.HealthcheckUrl)
}

func Test_loadLBConfig_WhenConfigInYAML_ReturnsExpectedParsedConfiguration(t *testing.T) {
	path := writeTempFile(t, "servers.yaml", `
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
`)
	cfg, err := loadLBConfig(path, zap.NewNop().Sugar())
	require.NoError(t, err)
	require.Len(t, cfg.ServerPools, 1)
	sp := cfg.ServerPools[0]
	assert.Equal(t, "serverPool", sp.Name)
	assert.Equal(t, "rb", sp.Strategy)
	require.Len(t, sp.Servers, 1)
	s := sp.Servers[0]
	assert.Equal(t, "1", s.ID)
	assert.Equal(t, "server1", s.Name)
	assert.Equal(t, "api", s.Protocol)
	assert.Equal(t, "localhost", s.Host)
	assert.Equal(t, 8080, s.Port)
	assert.Equal(t, "/health", s.HealthcheckUrl)
}

func Test_loadLBConfig_WhenUnsupportedExtension_ReturnsError(t *testing.T) {
	path := writeTempFile(t, "testServers.txt", `{}`)
	_, err := loadLBConfig(path, zap.NewNop().Sugar())
	require.Error(t, err)
	assert.ErrorContains(t, err, "unsupported config format: .txt")
}

func Test_buildServerPools_MapsConfigToRuntimeServerConfiguration(t *testing.T) {
	cfg := &LoadBalancerConfig{
		ServerPools: []serverPoolConfig{
			{
				Name:     "p1",
				Strategy: "rb",
				Servers: []serverConfig{
					{ID: "1", Name: "s1", Protocol: "http", Host: "localhost", Port: 8080, HealthcheckUrl: "/h"},
				},
			},
		},
	}
	pools := buildServerPools(cfg)
	require.Len(t, pools, 1)
	p := pools[0]
	assert.Equal(t, "p1", p.Name)
	assert.Equal(t, "rb", p.Strategy)
	require.Len(t, p.Servers, 1)
	s := p.Servers[0]
	assert.Equal(t, "1", s.ID)
	assert.Equal(t, "s1", s.Name)
	assert.Equal(t, "http://localhost:8080", s.Url)
	assert.True(t, s.IsHealthy)
}

func Test_Load_WhenNoCLIArgs_AssumesDefaultValues(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	require.NoError(t, os.Chdir(dir))

	_ = os.WriteFile("servers.json", []byte(`{
		"serverPools": [
			{
				"name": "pool",
				"strategy": "rb",
				"servers": [{"id":"1","name":"s1","protocol":"http","host":"localhost","port":8080,"healthcheckUrl":"/health"}]
			}
		]
	}`), 0o644)

	resetFlags(t)

	appCfg, err := Load(zap.NewNop().Sugar())

	require.NoError(t, err)
	require.NotNil(t, appCfg)
	assert.Equal(t, 3000, appCfg.LbPort)
	assert.Equal(t, 3001, appCfg.AdminPort)
	require.Len(t, appCfg.ServerPools, 1)
	assert.Equal(t, "pool", appCfg.ServerPools[0].Name)
}

func Test_Load_WhenCLIArgs_AssumesValuesFromArgs(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	require.NoError(t, os.Chdir(dir))
	require.NoError(t, os.WriteFile(
		"servers1.json",
		[]byte(`{
		"serverPools": [
			{
				"name": "pool1",
				"strategy": "rb",
				"servers": [{"id":"1","name":"s1","protocol":"http","host":"localhost","port":8081,"healthcheckUrl":"/health"}]
			}
		]
	}`),
		0o644))

	args := []string{
		os.Args[0],
		"-lb-port=4000",
		"-admin-port=5000",
		"-server-config=servers1.json",
	}
	resetFlags(t, args...)

	appCfg, err := Load(zap.NewNop().Sugar())
	require.NoError(t, err)
	assert.Equal(t, 4000, appCfg.LbPort)
	assert.Equal(t, 5000, appCfg.AdminPort)
	require.Len(t, appCfg.ServerPools, 1)
	assert.Equal(t, "pool1", appCfg.ServerPools[0].Name)
}

func Test_Load_WhenNoPools_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(old) })
	require.NoError(t, os.Chdir(dir))

	require.NoError(t, os.WriteFile("servers.json", []byte(`{"serverPools": []}`), 0o644))

	resetFlags(t)

	_, err := Load(zap.NewNop().Sugar())
	require.Error(t, err)
	assert.ErrorContains(t, err, "no servers found in")
}
