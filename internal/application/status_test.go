package application

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"debafr/internal/domain"
)

func TestActivePorts(t *testing.T) {
	t.Parallel()

	locPorts := []LocationPort{
		{Location: "/api", BluePort: "3001", GreenPort: "3011"},
		{Location: "/", BluePort: "3002", GreenPort: "3012"},
	}

	assert.Equal(t, []activePort{
		{Location: "/api", Port: "3001"},
		{Location: "/", Port: "3002"},
	}, activePorts(locPorts, domain.StrategyBlue))

	assert.Equal(t, []activePort{
		{Location: "/api", Port: "3011"},
		{Location: "/", Port: "3012"},
	}, activePorts(locPorts, domain.StrategyGreen))
}
