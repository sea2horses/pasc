package toolchain

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"pascc/internal/i18n"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// Generated programs are compiled with the Go toolchain. When the machine has
// no suitable Go, pascc installs a private copy once under the user cache
// directory and reuses it from then on.

const (
	downloadBase  = "https://dl.google.com/go/"
	latestVersion = "https://go.dev/VERSION?m=text"
)

// goTool is the Go command to run and the extra environment it needs.
type goTool struct {
	path string
	env  []string
}

// findGo returns a Go toolchain at least as new as minimum (e.g. "1.25.0"),
// installing one if the system has none or an older one.
func findGo(minimum string) (goTool, error) {
	systemVersion := ""
	if path, err := exec.LookPath("go"); err == nil {
		systemVersion = goVersionOf(path)
		if systemVersion != "" && !versionLess(systemVersion, minimum) {
			return goTool{path: path}, nil
		}
	}

	version := wantedVersion()
	if version == "" || versionLess(version, minimum) {
		latest, err := fetchLatestVersion()
		if err != nil {
			return goTool{}, downloadError(version, err)
		}
		version = latest
	}

	root, err := managedRoot(version)
	if err != nil {
		return goTool{}, err
	}
	// GOTOOLCHAIN=local keeps the managed go from fetching yet another toolchain.
	tool := goTool{path: goBinary(root), env: []string{"GOTOOLCHAIN=local"}}
	if _, err := os.Stat(tool.path); err == nil {
		return tool, nil
	}

	if systemVersion == "" {
		notice(i18n.T(i18n.GoSetupMissing))
	} else {
		notice(i18n.T(i18n.GoSetupTooOld, systemVersion, minimum))
	}
	if err := installGo(version, root); err != nil {
		return goTool{}, err
	}
	return tool, nil
}

// wantedVersion is the Go release pascc itself was built with, which is known
// to compile the embedded runtime; development builds report none.
func wantedVersion() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	if !regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`).MatchString(v) {
		return ""
	}
	return v
}

func fetchLatestVersion() (string, error) {
	resp, err := httpGet(latestVersion)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return "", err
	}
	first, _, _ := strings.Cut(string(body), "\n")
	return strings.TrimPrefix(strings.TrimSpace(first), "go"), nil
}

// managedRoot is where the private toolchain for version lives.
func managedRoot(version string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "pasc", "go", "go"+version), nil
}

func goBinary(root string) string {
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(root, "bin", name)
}

func goVersionOf(path string) string {
	out, err := exec.Command(path, "env", "GOVERSION").Output()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "go")
}

// versionLess compares release numbers like "1.25.0"; pre-release suffixes
// (rc1, beta2) are ignored.
func versionLess(a, b string) bool {
	pa, pb := versionParts(a), versionParts(b)
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func versionParts(v string) [3]int {
	var parts [3]int
	for i, field := range strings.SplitN(v, ".", 3) {
		digits := field
		if j := strings.IndexFunc(field, func(r rune) bool { return r < '0' || r > '9' }); j >= 0 {
			digits = field[:j]
		}
		parts[i], _ = strconv.Atoi(digits)
	}
	return parts
}

// requiredGoVersion reads the go directive of the embedded runtime module.
func requiredGoVersion(runtimeMod []byte) string {
	if m := regexp.MustCompile(`(?m)^go\s+(\S+)`).FindSubmatch(runtimeMod); m != nil {
		return string(m[1])
	}
	return "1.21"
}

/* Installation */

func installGo(version, root string) error {
	archive := fmt.Sprintf("go%s.%s-%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		archive = strings.TrimSuffix(archive, ".tar.gz") + ".zip"
	}

	want, err := fetchChecksum(archive)
	if err != nil {
		return downloadError(version, err)
	}

	parent := filepath.Dir(root)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return extractError(err)
	}
	file, err := os.CreateTemp(parent, ".download-*")
	if err != nil {
		return extractError(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	got, err := download(downloadBase+archive, version, file)
	if err != nil {
		return downloadError(version, err)
	}

	step(i18n.T(i18n.GoSetupVerifying))
	if got != want {
		fmt.Fprintln(os.Stderr)
		return errors.New(i18n.T(i18n.GoChecksumMismatch))
	}
	done()

	step(i18n.T(i18n.GoSetupExtracting))
	// Extract next to the final location and rename at the end, so an
	// interrupted install never leaves a half-written toolchain behind.
	staging, err := os.MkdirTemp(parent, ".install-*")
	if err != nil {
		return extractError(err)
	}
	defer os.RemoveAll(staging)
	if strings.HasSuffix(archive, ".zip") {
		err = extractZip(file.Name(), staging)
	} else {
		err = extractTarGz(file.Name(), staging)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr)
		return extractError(err)
	}
	// Release archives hold a single top-level "go" directory.
	if err := os.Rename(filepath.Join(staging, "go"), root); err != nil {
		// Another pascc may have finished the same install first.
		if _, statErr := os.Stat(goBinary(root)); statErr != nil {
			fmt.Fprintln(os.Stderr)
			return extractError(err)
		}
	}
	done()

	success(i18n.T(i18n.GoSetupInstalled, version, root))
	detail(i18n.T(i18n.GoSetupFirstBuild))
	fmt.Fprintln(os.Stderr)
	return nil
}

func fetchChecksum(archive string) (string, error) {
	resp, err := httpGet(downloadBase + archive + ".sha256")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256))
	if err != nil {
		return "", err
	}
	sum := strings.Fields(string(body))
	if len(sum) == 0 {
		return "", errors.New("empty checksum")
	}
	return strings.ToLower(sum[0]), nil
}

// download streams url into w and returns the SHA-256 of what it wrote.
func download(url, version string, w io.Writer) (string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	detail(i18n.T(i18n.GoSetupOnce, version, formatSize(resp.ContentLength)))
	fmt.Fprintln(os.Stderr)

	hash := sha256.New()
	progress := newProgress(resp.ContentLength)
	_, err = io.Copy(io.MultiWriter(w, hash, progress), resp.Body)
	progress.finish(err == nil)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "pascc")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return resp, nil
}

func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeJoin(dest, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			err = writeFile(target, tr, header.FileInfo().Mode().Perm())
		case tar.TypeSymlink:
			if _, err = safeJoin(filepath.Dir(target), header.Linkname); err == nil {
				err = os.Symlink(header.Linkname, target)
			}
		}
		if err != nil {
			return err
		}
	}
}

func extractZip(archive, dest string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, entry := range zr.File {
		target, err := safeJoin(dest, entry.Name)
		if err != nil {
			return err
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		rc, err := entry.Open()
		if err != nil {
			return err
		}
		err = writeFile(target, rc, entry.Mode().Perm())
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFile(path string, r io.Reader, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm|0200)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// safeJoin rejects archive entries that would escape dest.
func safeJoin(dest, name string) (string, error) {
	target := filepath.Join(dest, name)
	if rel, err := filepath.Rel(dest, target); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe path in archive: %s", name)
	}
	return target, nil
}

func downloadError(version string, err error) error {
	return fmt.Errorf("%s\n%s", i18n.T(i18n.GoDownloadFailed, version, err), i18n.T(i18n.GoManualHint))
}

func extractError(err error) error {
	return errors.New(i18n.T(i18n.GoExtractFailed, err))
}

/* Console output (stderr, so program output stays clean) */

var (
	accent  = color.New(color.FgCyan, color.Bold)
	okColor = color.New(color.FgGreen, color.Bold)
)

func notice(msg string) {
	fmt.Fprintln(os.Stderr)
	accent.Fprint(os.Stderr, "⚙ ")
	fmt.Fprintln(os.Stderr, msg)
}

func detail(msg string) { fmt.Fprintln(os.Stderr, "  "+msg) }

func step(msg string) { fmt.Fprint(os.Stderr, "  "+msg+" ") }

func done() { okColor.Fprintln(os.Stderr, i18n.T(i18n.GoSetupDone)) }

func success(msg string) {
	okColor.Fprint(os.Stderr, "✔ ")
	fmt.Fprintln(os.Stderr, msg)
}

func formatSize(n int64) string {
	if n < 0 {
		return "?"
	}
	return fmt.Sprintf("%.1f MB", float64(n)/1e6)
}

// progress draws a bar on terminals and a line every 10% elsewhere (logs, CI).
type progress struct {
	total, written int64
	start, drawn   time.Time
	tty            bool
	lastDecile     int64
}

func newProgress(total int64) *progress {
	fd := os.Stderr.Fd()
	return &progress{total: total, start: time.Now(), tty: isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)}
}

func (p *progress) Write(b []byte) (int, error) {
	p.written += int64(len(b))
	if p.tty {
		if time.Since(p.drawn) >= 100*time.Millisecond {
			p.draw()
		}
	} else if p.total > 0 {
		if decile := p.written * 10 / p.total; decile > p.lastDecile {
			p.lastDecile = decile
			fmt.Fprintf(os.Stderr, "  %s... %d%%\n", i18n.T(i18n.GoSetupDownloading), decile*10)
		}
	}
	return len(b), nil
}

func (p *progress) draw() {
	p.drawn = time.Now()
	const width = 30
	speed := float64(p.written) / max(time.Since(p.start).Seconds(), 0.001) / 1e6
	if p.total <= 0 {
		fmt.Fprintf(os.Stderr, "\r  %s  %.1f MB  %.1f MB/s ", i18n.T(i18n.GoSetupDownloading), float64(p.written)/1e6, speed)
		return
	}
	fraction := float64(p.written) / float64(p.total)
	filled := int(fraction * width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Fprintf(os.Stderr, "\r  %s  %s %3.0f%%  %.1f/%.1f MB  %.1f MB/s ",
		i18n.T(i18n.GoSetupDownloading), accent.Sprint(bar), fraction*100,
		float64(p.written)/1e6, float64(p.total)/1e6, speed)
}

func (p *progress) finish(ok bool) {
	if p.tty {
		if ok {
			p.written = max(p.written, p.total)
		}
		p.draw()
		fmt.Fprintln(os.Stderr)
	}
}
