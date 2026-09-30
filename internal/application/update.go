package application

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
)

const (
	latestReleaseAPIURL = "https://api.github.com/repos/dimns/debafr/releases/latest"
	downloadURLTmpl     = "https://github.com/dimns/debafr/releases/download/%s/debafr_%s_%s_%s.tar.gz"
	binaryName          = "debafr"

	maxBinarySize = 256 << 20 // 256 MiB
	binFileMode   = 0o755
)

// Update checks the latest released version and replaces the running binary when a newer one exists.
func Update(currentVersion string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultUpdateTimeout)
	defer cancel()

	latestVersion, err := latestVersion(ctx)
	if err != nil {
		return err
	}

	upToDate, err := isUpToDate(currentVersion, latestVersion)
	if err != nil {
		return err
	}
	if upToDate {
		fmt.Printf("debafr %s is up to date\n", latestVersion)

		return nil
	}

	fmt.Printf("Updating debafr %s -> %s\n", currentVersion, latestVersion)

	url := fmt.Sprintf(downloadURLTmpl, latestVersion, strings.TrimPrefix(latestVersion, "v"), runtime.GOOS, runtime.GOARCH)
	if err := download(ctx, url); err != nil {
		return err
	}

	fmt.Printf("debafr %s installed\n", latestVersion)

	return nil
}

func latestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseAPIURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get latest release: unexpected status %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"` //nolint:tagliatelle // GitHub API naming
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode latest release: %w", err)
	}

	tag := release.TagName
	if _, err := version.NewVersion(tag); err != nil {
		return "", fmt.Errorf("unexpected latest release tag %q: %w", tag, err)
	}

	return tag, nil
}

func isUpToDate(currentVersion, latestVersion string) (bool, error) {
	current, err := version.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("parse current version %q: %w", currentVersion, err)
	}

	latest, err := version.NewVersion(latestVersion)
	if err != nil {
		return false, fmt.Errorf("parse latest version %q: %w", latestVersion, err)
	}

	return current.GreaterThanOrEqual(latest), nil
}

func download(ctx context.Context, url string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}

	tmp, err := os.MkdirTemp(filepath.Dir(exePath), binaryName+"-update-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	newBinary := filepath.Join(tmp, binaryName)
	if err := extractBinary(resp.Body, newBinary); err != nil {
		return err
	}

	if err := os.Rename(newBinary, exePath); err != nil {
		return fmt.Errorf("replace %s: %w", exePath, err)
	}

	return nil
}

func extractBinary(r io.Reader, dst string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("archive does not contain %s binary", binaryName)
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}

		if header.Name != binaryName {
			continue
		}

		return writeBinary(dst, tr)
	}
}

func writeBinary(dst string, r io.Reader) error {
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, binFileMode)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}

	// ponytail: bounded copy instead of io.Copy to cap decompression size; raise maxBinarySize if releases grow.
	if _, err := io.CopyN(f, r, maxBinarySize); err != nil && !errors.Is(err, io.EOF) {
		f.Close()

		return fmt.Errorf("extract %s: %w", dst, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dst, err)
	}

	if err := os.Chmod(dst, binFileMode); err != nil {
		return fmt.Errorf("chmod %s: %w", dst, err)
	}

	return nil
}
