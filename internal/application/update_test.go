package application

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUpToDate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		current, latest string
		want            bool
	}{
		{"0.4.4", "v0.4.4", true},
		{"0.4.5", "v0.4.4", true},
		{"0.4.3", "v0.4.4", false},
		{"1.0.0", "v0.9.9", true},
		{"0.0.0", "v0.4.4", false},
	}

	for _, c := range cases {
		got, err := isUpToDate(c.current, c.latest)
		require.NoError(t, err)
		assert.Equal(t, c.want, got, "%s vs %s", c.current, c.latest)
	}

	_, err := isUpToDate("not-a-version", "v0.4.4")
	assert.Error(t, err)
}

func TestExtractBinary(t *testing.T) {
	t.Parallel()

	archive := func(entries map[string]string) *bytes.Buffer {
		buf := &bytes.Buffer{}
		gz := gzip.NewWriter(buf)
		tw := tar.NewWriter(gz)
		for name, body := range entries {
			require.NoError(t, tw.WriteHeader(&tar.Header{
				Name: name,
				Mode: 0o755,
				Size: int64(len(body)),
			}))
			_, err := tw.Write([]byte(body))
			require.NoError(t, err)
		}
		require.NoError(t, tw.Close())
		require.NoError(t, gz.Close())

		return buf
	}

	t.Run("extracts binary from archive", func(t *testing.T) {
		t.Parallel()

		dst := filepath.Join(t.TempDir(), binaryName)
		require.NoError(t, extractBinary(archive(map[string]string{
			"README.md": "docs",
			binaryName:  "payload",
		}), dst))

		body, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, "payload", string(body))
	})

	t.Run("fails when binary is absent", func(t *testing.T) {
		t.Parallel()

		dst := filepath.Join(t.TempDir(), binaryName)
		assert.Error(t, extractBinary(archive(map[string]string{"README.md": "docs"}), dst))
	})
}
