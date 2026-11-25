package tests

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	englishReadmePath = "../README.md"
	germanReadmePath  = "../README_de.md"
)

var markdownLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

// readmeLayout is the translation-independent skeleton of a README. Wording
// differs between the languages, so only structure, link targets and code
// block types are comparable.
type readmeLayout struct {
	headings []string
	links    []string
	blocks   []string
}

func parseReadme(t *testing.T, path string) readmeLayout {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var layout readmeLayout
	inBlock := false
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "```") {
			if inBlock {
				inBlock = false
				continue
			}
			inBlock = true
			layout.blocks = append(layout.blocks, strings.TrimPrefix(line, "```"))
			continue
		}
		if inBlock {
			continue
		}
		if level, ok := headingLevel(line); ok {
			layout.headings = append(layout.headings, level)
		}
		for _, match := range markdownLink.FindAllStringSubmatch(line, -1) {
			layout.links = append(layout.links, match[1])
		}
	}
	return layout
}

func headingLevel(line string) (string, bool) {
	hashes := len(line) - len(strings.TrimLeft(line, "#"))
	if hashes == 0 || !strings.HasPrefix(line[hashes:], " ") {
		return "", false
	}
	return strings.Repeat("#", hashes), true
}

func TestReadmesShareStructure(t *testing.T) {
	// Arrange
	english := parseReadme(t, englishReadmePath)
	german := parseReadme(t, germanReadmePath)

	// Act + Assert
	compareLayout(t, "headings", english.headings, german.headings)
}

func TestReadmesShareCodeBlocks(t *testing.T) {
	// Arrange
	english := parseReadme(t, englishReadmePath)
	german := parseReadme(t, germanReadmePath)

	// Act + Assert
	compareLayout(t, "code blocks", english.blocks, german.blocks)
}

func TestReadmesShareLinkTargets(t *testing.T) {
	// Arrange
	english := parseReadme(t, englishReadmePath)
	german := parseReadme(t, germanReadmePath)

	// Act
	englishLinks := sortedUnique(withoutLanguageLinks(english.links))
	germanLinks := sortedUnique(withoutLanguageLinks(german.links))

	// Assert
	compareLayout(t, "link targets", englishLinks, germanLinks)
}

func TestReadmesLinkToTheirTranslation(t *testing.T) {
	// Arrange
	english := parseReadme(t, englishReadmePath)
	german := parseReadme(t, germanReadmePath)

	// Assert
	if !slices.Contains(english.links, "README_de.md") {
		t.Fatal("README.md does not link to README_de.md")
	}
	if !slices.Contains(german.links, "README.md") {
		t.Fatal("README_de.md does not link to README.md")
	}
}

func withoutLanguageLinks(links []string) []string {
	kept := make([]string, 0, len(links))
	for _, link := range links {
		if link == "README.md" || link == "README_de.md" {
			continue
		}
		kept = append(kept, link)
	}
	return kept
}

func sortedUnique(values []string) []string {
	slices.Sort(values)
	return slices.Compact(values)
}

func compareLayout(t *testing.T, what string, english, german []string) {
	t.Helper()
	at := func(values []string, index int) string {
		if index >= len(values) {
			return "end of file"
		}
		return fmt.Sprintf("%q", values[index])
	}
	for index := 0; index < len(english) || index < len(german); index++ {
		if index < len(english) && index < len(german) && english[index] == german[index] {
			continue
		}
		t.Fatalf("%s differ at position %d: README.md has %s, README_de.md has %s", what, index, at(english, index), at(german, index))
	}
}
