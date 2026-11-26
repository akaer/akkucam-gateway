package gateway

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	// href="/..." and src="/..." attributes in the dashboard markup.
	assetReferencePattern = regexp.MustCompile(`(?:href|src)="(/[^"]*)"`)
	// The brand mark must contain an inline SVG, not a text glyph.
	brandMarkPattern = regexp.MustCompile(`(?s)class="brand-mark"[^>]*>\s*<svg`)
	// The favicon link must point at a local SVG.
	faviconPattern = regexp.MustCompile(`<link[^>]+rel="icon"[^>]+href="(/[^"]+)"`)
)

func readIndex(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(webRoot, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	return string(raw)
}

func TestDashboardReferencesExistingStaticAssets(t *testing.T) {
	// Arrange
	index := readIndex(t)
	references := assetReferencePattern.FindAllStringSubmatch(index, -1)

	// Act and Assert
	if len(references) == 0 {
		t.Fatal("index.html has no static asset references")
	}
	for _, match := range references {
		path := match[1]
		// The brand link points at the dashboard route, and /api/ entries are
		// HTTP routes rather than files.
		if path == "/" || strings.HasPrefix(path, "/api/") {
			continue
		}
		local := filepath.Join(webRoot, strings.TrimPrefix(path, "/"))
		if _, err := os.Stat(local); err != nil {
			t.Errorf("referenced asset %q does not exist at %s", path, local)
		}
	}
}

func TestBrandMarkRendersCameraIcon(t *testing.T) {
	// Arrange
	index := readIndex(t)

	// Act
	matches := brandMarkPattern.MatchString(index)

	// Assert
	if !matches {
		t.Fatal("brand mark does not contain an inline SVG")
	}
}

func TestDashboardUsesOneIconSet(t *testing.T) {
	// Arrange: text glyphs stood in for icons before ADR-0029 and render
	// differently on every platform.
	index := readIndex(t)
	raw, err := os.ReadFile(filepath.Join(webRoot, "app.js"))
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	glyphs := []string{"◐", "⌄", "⋯", "▶", "■", "⛶", "▣"}

	// Act and Assert
	for _, glyph := range glyphs {
		if strings.Contains(index, glyph) {
			t.Errorf("index.html still uses the text glyph %q as an icon", glyph)
		}
		if strings.Contains(string(raw), glyph) {
			t.Errorf("app.js still uses the text glyph %q as an icon", glyph)
		}
	}
	if !strings.Contains(index, `id="icon-camera"`) {
		t.Error("index.html does not define the shared icon sprite")
	}
}

func TestStylesheetUsesOnlyShippedFontWeights(t *testing.T) {
	// Arrange: web/fonts/ ships Inter Tight 400 and 700. Other weights are
	// synthesised by the browser (ADR-0023).
	raw, err := os.ReadFile(filepath.Join(webRoot, "style.css"))
	if err != nil {
		t.Fatalf("read style.css: %v", err)
	}
	weightPattern := regexp.MustCompile(`font-weight:\s*([0-9]{3})`)
	allowed := map[string]bool{"400": true, "700": true}

	// Act and Assert
	for _, match := range weightPattern.FindAllStringSubmatch(string(raw), -1) {
		if !allowed[match[1]] {
			t.Errorf("style.css uses font-weight %s, but only 400 and 700 are shipped", match[1])
		}
	}
}

func TestStylesheetRespectsSafeAreaInsets(t *testing.T) {
	// Arrange: the viewport uses viewport-fit=cover, so the phone layout must
	// keep content clear of the notch and the home indicator (ADR-0030).
	index := readIndex(t)
	raw, err := os.ReadFile(filepath.Join(webRoot, "style.css"))
	if err != nil {
		t.Fatalf("read style.css: %v", err)
	}
	css := string(raw)

	// Act and Assert
	if !strings.Contains(index, "viewport-fit=cover") {
		t.Fatal("index.html does not opt into the safe area")
	}
	for _, inset := range []string{"env(safe-area-inset-top", "env(safe-area-inset-bottom"} {
		if !strings.Contains(css, inset) {
			t.Errorf("style.css does not use %s", inset)
		}
	}
}

func TestDashboardHasNoInlineEventHandlers(t *testing.T) {
	// Arrange: inline handlers block a strict Content Security Policy (ADR-0028).
	index := readIndex(t)
	inlineHandlerPattern := regexp.MustCompile(`(?i)\son[a-z]+="`)

	// Act and Assert
	if matches := inlineHandlerPattern.FindAllString(index, -1); len(matches) > 0 {
		t.Fatalf("index.html uses inline event handlers: %v", matches)
	}
}

func TestStylesheetScalesWithRootFontSize(t *testing.T) {
	// Arrange: OS text scaling only reaches relative sizes (ADR-0030).
	raw, err := os.ReadFile(filepath.Join(webRoot, "style.css"))
	if err != nil {
		t.Fatalf("read style.css: %v", err)
	}
	css := string(raw)
	index := readIndex(t)
	pixelTypePattern := regexp.MustCompile(`(?m)^\s*font(?:-size)?:\s*[0-9.]+px`)

	// Act and Assert
	if !strings.Contains(index, `name="text-scale"`) {
		t.Error("index.html does not opt into the operating system text scale")
	}
	if matches := pixelTypePattern.FindAllString(css, -1); len(matches) > 0 {
		t.Fatalf("style.css sizes text in pixels: %v", matches)
	}
}

func TestFaviconUsesLocalThemeIcons(t *testing.T) {
	// Arrange
	index := readIndex(t)
	matches := faviconPattern.FindAllStringSubmatch(index, -1)

	// Act and Assert
	if len(matches) != 1 {
		t.Fatalf("found %d favicon links, want 1", len(matches))
	}
	href := matches[0][1]
	if !strings.HasPrefix(href, "/icon-") {
		t.Fatalf("favicon href=%q, want a local /icon-*.svg", href)
	}
	for _, name := range []string{"icon-light.svg", "icon-dark.svg"} {
		if _, err := os.Stat(filepath.Join(webRoot, name)); err != nil {
			t.Fatalf("theme icon %s is missing: %v", name, err)
		}
	}
}

func staticHandler(t *testing.T) (http.Handler, string) {
	t.Helper()
	webRoot := t.TempDir()
	streams := &fakeStreams{root: t.TempDir()}
	return New(webRoot, streams).Routes(), webRoot
}

func TestGzipCompressesTextAssets(t *testing.T) {
	// Arrange
	handler, webRoot := staticHandler(t)
	content := strings.Repeat("const answer = 42;\n", 200)
	if err := os.WriteFile(filepath.Join(webRoot, "app.js"), []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/app.js", nil)
	request.Header.Set("Accept-Encoding", "gzip")

	// Act
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	// Assert
	if got := recorder.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	if got := recorder.Header().Get("Vary"); got != "Accept-Encoding" {
		t.Errorf("Vary = %q, want Accept-Encoding", got)
	}
	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("open gzip body: %v", err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip body: %v", err)
	}
	if string(body) != content {
		t.Error("decompressed body differs from the file")
	}
}

func TestGzipLeavesNonTextAndRangeRequestsAlone(t *testing.T) {
	// Arrange
	handler, webRoot := staticHandler(t)
	for name, content := range map[string]string{"font.woff2": "font", "app.js": "const a = 1;\n"} {
		if err := os.WriteFile(filepath.Join(webRoot, name), []byte(content), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	tests := []struct {
		name    string
		path    string
		method  string
		headers map[string]string
	}{
		{name: "binary asset", path: "/font.woff2", headers: map[string]string{"Accept-Encoding": "gzip"}},
		{name: "range request", path: "/app.js", headers: map[string]string{"Accept-Encoding": "gzip", "Range": "bytes=0-3"}},
		{name: "no gzip accepted", path: "/app.js"},
		{name: "head request", path: "/app.js", method: http.MethodHead, headers: map[string]string{"Accept-Encoding": "gzip"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			method := test.method
			if method == "" {
				method = http.MethodGet
			}
			request := httptest.NewRequestWithContext(context.Background(), method, test.path, nil)
			for key, value := range test.headers {
				request.Header.Set(key, value)
			}
			recorder := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(recorder, request)

			// Assert
			if got := recorder.Header().Get("Content-Encoding"); got != "" {
				t.Errorf("Content-Encoding = %q, want none", got)
			}
		})
	}
}
