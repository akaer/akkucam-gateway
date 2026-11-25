package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"akkucam-gateway/internal/apierr"
)

const webRoot = "../../web"

var (
	placeholderPattern = regexp.MustCompile(`\{(\w+)\}`)
	// data-i18n, data-i18n-title, data-i18n-aria-label and
	// data-i18n-document-title attributes in HTML and in the JavaScript templates.
	i18nAttributePattern = regexp.MustCompile(`data-i18n(?:-title|-aria-label|-document-title)?="([^"]+)"`)
	// Single- and double-quoted string literals in app.js. Used to find catalog
	// keys passed to helper functions and metadataItem.
	quotedStringPattern = regexp.MustCompile(`'([^'\\\n]*)'|"([^"\\\n]*)"`)
	// Single-quoted literals that look like catalog keys. The double-quoted
	// alternative in quotedStringPattern swallows keys inside template-literal
	// attributes such as data-i18n="${cond ? 'a.b' : 'c.d'}".
	singleQuotedKeyPattern = regexp.MustCompile(`'([a-z][a-zA-Z0-9]*\.[a-zA-Z0-9.]+)'`)
)

func readCatalog(t *testing.T, lang string) map[string]string {
	t.Helper()
	path := filepath.Join(webRoot, "locales", lang+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var catalog map[string]string
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return catalog
}

func sortedKeys(catalog map[string]string) []string {
	keys := make([]string, 0, len(catalog))
	for key := range catalog {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func placeholders(value string) []string {
	matches := placeholderPattern.FindAllStringSubmatch(value, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match[1])
	}
	sort.Strings(names)
	return names
}

func hasKey(catalog map[string]string, key string) bool {
	if _, ok := catalog[key]; ok {
		return true
	}
	// Plural keys are stored as <key>.one and <key>.other.
	if _, ok := catalog[key+".one"]; ok {
		return true
	}
	_, ok := catalog[key+".other"]
	return ok
}

func TestLocalesAreValidJSON(t *testing.T) {
	// Arrange
	_ = webRoot

	// Act
	english := readCatalog(t, "en")
	german := readCatalog(t, "de")

	// Assert
	if len(english) == 0 {
		t.Fatal("English catalog is empty")
	}
	if len(german) == 0 {
		t.Fatal("German catalog is empty")
	}
}

func TestLocalesHaveSameKeys(t *testing.T) {
	// Arrange
	english := readCatalog(t, "en")
	german := readCatalog(t, "de")

	// Act
	englishKeys := sortedKeys(english)
	germanKeys := sortedKeys(german)

	// Assert
	if strings.Join(englishKeys, "\n") != strings.Join(germanKeys, "\n") {
		t.Fatalf("catalog keys differ\nen=%v\nde=%v", englishKeys, germanKeys)
	}
}

func TestLocalePlaceholdersMatch(t *testing.T) {
	// Arrange
	english := readCatalog(t, "en")
	german := readCatalog(t, "de")

	// Act and Assert
	for key, englishValue := range english {
		germanValue, ok := german[key]
		if !ok {
			t.Fatalf("German catalog is missing %q", key)
		}
		englishPlaceholders := placeholders(englishValue)
		germanPlaceholders := placeholders(germanValue)
		if strings.Join(englishPlaceholders, ",") != strings.Join(germanPlaceholders, ",") {
			t.Fatalf("placeholders for %q differ: en=%v de=%v", key, englishPlaceholders, germanPlaceholders)
		}
	}
}

func TestLocaleValuesAreNonEmpty(t *testing.T) {
	// Arrange
	catalogs := map[string]map[string]string{
		"en": readCatalog(t, "en"),
		"de": readCatalog(t, "de"),
	}

	// Act and Assert
	for lang, catalog := range catalogs {
		for key, value := range catalog {
			if strings.TrimSpace(value) == "" {
				t.Fatalf("%s.%s is empty", lang, key)
			}
		}
	}
}

func TestLocaleValuesAreWholePhrases(t *testing.T) {
	// Arrange: sentences are assembled from one complete catalog value, never
	// from fragments glued together in app.js (ADR-0028).
	catalogs := map[string]map[string]string{
		"en": readCatalog(t, "en"),
		"de": readCatalog(t, "de"),
	}

	// Act and Assert
	for lang, catalog := range catalogs {
		for key, value := range catalog {
			if strings.TrimSpace(value) != value {
				t.Errorf("%s.%s starts or ends with whitespace: %q", lang, key, value)
			}
			if strings.HasPrefix(value, "·") || strings.HasSuffix(value, "·") {
				t.Errorf("%s.%s is a fragment with a separator: %q", lang, key, value)
			}
		}
	}
}

func TestIndexHTMLKeysExistInLocales(t *testing.T) {
	// Arrange
	raw, err := os.ReadFile(filepath.Join(webRoot, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	catalogs := map[string]map[string]string{
		"en": readCatalog(t, "en"),
		"de": readCatalog(t, "de"),
	}

	// Act
	keys := make([]string, 0)
	for _, match := range i18nAttributePattern.FindAllStringSubmatch(string(raw), -1) {
		key := match[1]
		if strings.Contains(key, "${") {
			continue
		}
		keys = append(keys, key)
	}

	// Assert
	if len(keys) == 0 {
		t.Fatal("index.html has no data-i18n attributes")
	}
	for lang, catalog := range catalogs {
		for _, key := range keys {
			if !hasKey(catalog, key) {
				t.Fatalf("%s catalog is missing key %q used in index.html", lang, key)
			}
		}
	}
}

func TestAppJSKeysExistInLocales(t *testing.T) {
	// Arrange
	raw, err := os.ReadFile(filepath.Join(webRoot, "app.js"))
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	english := readCatalog(t, "en")
	german := readCatalog(t, "de")
	prefixes := make(map[string]bool)
	for key := range english {
		if dot := strings.IndexByte(key, '.'); dot > 0 {
			prefixes[key[:dot]] = true
		}
	}

	// Act
	keys := make(map[string]bool)
	for _, match := range i18nAttributePattern.FindAllStringSubmatch(string(raw), -1) {
		key := match[1]
		if strings.Contains(key, "${") {
			continue
		}
		keys[key] = true
	}
	for _, match := range quotedStringPattern.FindAllStringSubmatch(string(raw), -1) {
		value := match[1]
		if value == "" {
			value = match[2]
		}
		if !strings.Contains(value, ".") || strings.Contains(value, "${") {
			continue
		}
		if prefixes[value[:strings.IndexByte(value, '.')]] {
			keys[value] = true
		}
	}
	// Keys inside template-literal attributes are hidden from the double-quoted
	// alternative above, so collect them with a dedicated pattern.
	for _, match := range singleQuotedKeyPattern.FindAllStringSubmatch(string(raw), -1) {
		key := match[1]
		if prefixes[key[:strings.IndexByte(key, '.')]] {
			keys[key] = true
		}
	}

	// Assert
	if len(keys) == 0 {
		t.Fatal("app.js has no translatable keys")
	}
	for key := range keys {
		if !hasKey(english, key) {
			t.Fatalf("English catalog is missing key %q used in app.js", key)
		}
		if !hasKey(german, key) {
			t.Fatalf("German catalog is missing key %q used in app.js", key)
		}
	}
}

func TestAPICodesHaveTranslations(t *testing.T) {
	// Arrange
	catalogs := map[string]map[string]string{
		"en": readCatalog(t, "en"),
		"de": readCatalog(t, "de"),
	}

	// Act and Assert
	for _, code := range apierr.AllCodes() {
		key := "error.code." + string(code)
		for lang, catalog := range catalogs {
			if _, ok := catalog[key]; !ok {
				t.Fatalf("%s catalog is missing %q", lang, key)
			}
		}
	}
}
