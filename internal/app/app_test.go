package app

import (
	"github.com/MateuszW99/GoBalancer/internal/config"
	"github.com/MateuszW99/GoBalancer/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func Test_NewApp_BuildsAppWithExpectedServers(t *testing.T) {
	cfg := &config.AppConfig{
		ServerPools: []*server.ServerPool{
			{
				Servers:  make([]*server.Server, 0),
				Strategy: "rb",
			},
		},
		LbPort:    1,
		AdminPort: 2,
	}

	app, err := NewApp(
		WithConfig(cfg),
		WithLogger(zap.NewNop().Sugar()),
	)
	require.NoError(t, err)

	assert.Equal(t, cfg, app.Cfg)
	assert.Equal(t, ":1", app.LbProxy.Addr)
	assert.Equal(t, ":2", app.AdminServer.Addr)
}
