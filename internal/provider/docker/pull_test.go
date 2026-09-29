package docker

import (
	"testing"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/stretchr/testify/assert"
)

func Test_pullStats_add(t *testing.T) {
	t.Parallel()

	t.Run("Should count the bytes of a layer only once", func(t *testing.T) {
		t.Parallel()

		stats := newPullStats()

		stats.add("a", &jsonmessage.JSONProgress{Current: 100, Total: 1000})
		got := stats.add("a", &jsonmessage.JSONProgress{Current: 300, Total: 1000})

		assert.Equal(t, PullProgress{Current: 300, Total: 1000}, got, "docker sends the state of the layer, not the delta")
	})

	t.Run("Should pick up the size of a layer announced without one", func(t *testing.T) {
		t.Parallel()

		stats := newPullStats()

		stats.add("a", &jsonmessage.JSONProgress{})
		got := stats.add("a", &jsonmessage.JSONProgress{Current: 200, Total: 1000})

		assert.Equal(t, PullProgress{Current: 200, Total: 1000}, got, "the first message of a layer has no size")
	})

	t.Run("Should keep the bytes of a layer docker has completed", func(t *testing.T) {
		t.Parallel()

		stats := newPullStats()

		stats.add("a", &jsonmessage.JSONProgress{Current: 900, Total: 1000})
		got := stats.add("a", &jsonmessage.JSONProgress{})

		assert.Equal(t, PullProgress{Current: 900, Total: 1000}, got, "a finished layer is reported with an empty progress")
	})

	t.Run("Should sum up the layers", func(t *testing.T) {
		t.Parallel()

		stats := newPullStats()

		stats.add("a", &jsonmessage.JSONProgress{Current: 100, Total: 1000})
		got := stats.add("b", &jsonmessage.JSONProgress{Current: 50, Total: 250})

		assert.Equal(t, PullProgress{Current: 150, Total: 1250}, got)
	})
}
