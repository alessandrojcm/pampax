package discovery

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type compiledRule struct {
	source     RuleSource
	pattern    string
	ignoreFile string
	matcher    gitignore.Pattern
}

// IgnoreMatcher implements discovery ignore semantics with fixed precedence:
// default rules < .gitignore < .pampignore.
type IgnoreMatcher struct {
	root         string
	defaultRules []compiledRule
	gitRules     []compiledRule
	pampRules    []compiledRule
}

func NewIgnoreMatcher(root string) (*IgnoreMatcher, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve matcher root: %w", err)
	}

	matcher := &IgnoreMatcher{root: absRoot}

	for _, pattern := range DefaultIgnorePatterns() {
		rule, ok, buildErr := compileRule(pattern, RuleSourceDefault, "", "<default>")
		if buildErr != nil {
			return nil, fmt.Errorf("compile default ignore pattern %q: %w", pattern, buildErr)
		}
		if !ok {
			continue
		}

		matcher.defaultRules = append(matcher.defaultRules, rule)
	}

	gitFiles, pampaFiles, err := collectIgnoreFiles(absRoot)
	if err != nil {
		return nil, err
	}

	for _, ignoreFile := range gitFiles {
		rules, parseErr := parseIgnoreFile(absRoot, ignoreFile, RuleSourceGitIgnore)
		if parseErr != nil {
			return nil, parseErr
		}
		matcher.gitRules = append(matcher.gitRules, rules...)
	}

	for _, ignoreFile := range pampaFiles {
		rules, parseErr := parseIgnoreFile(absRoot, ignoreFile, RuleSourcePampIgnore)
		if parseErr != nil {
			return nil, parseErr
		}
		matcher.pampRules = append(matcher.pampRules, rules...)
	}

	return matcher, nil
}

func (m *IgnoreMatcher) ShouldSkipDir(relativePath string) bool {
	return m.DecisionFor(relativePath, true).Excluded
}

func (m *IgnoreMatcher) ShouldSkipFile(relativePath string) bool {
	return m.DecisionFor(relativePath, false).Excluded
}

func (m *IgnoreMatcher) DecisionFor(relativePath string, isDir bool) IgnoreDecision {
	normalized := normalizeFromRelative(relativePath)
	decision := IgnoreDecision{Path: normalized, IsDir: isDir, Source: RuleSourceNone}

	best, negated := m.lastMatch(normalized, isDir)
	if best == nil {
		return decision
	}

	decision.Matched = true
	decision.Source = best.source
	decision.Pattern = best.pattern
	decision.IgnoreFile = best.ignoreFile
	decision.Negated = negated
	decision.Excluded = !negated
	return decision
}

func (m *IgnoreMatcher) lastMatch(relativePath string, isDir bool) (*compiledRule, bool) {
	defaultMatch, defaultNegated := lastMatchingRule(m.defaultRules, relativePath, isDir)
	gitMatch, gitNegated := lastMatchingRule(m.gitRules, relativePath, isDir)
	pampMatch, pampNegated := lastMatchingRule(m.pampRules, relativePath, isDir)

	if pampMatch != nil {
		return pampMatch, pampNegated
	}

	if gitMatch != nil {
		return gitMatch, gitNegated
	}

	return defaultMatch, defaultNegated
}

func collectIgnoreFiles(root string) ([]string, []string, error) {
	var gitFiles []string
	var pampFiles []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		switch d.Name() {
		case ".gitignore":
			gitFiles = append(gitFiles, path)
		case ".pampignore":
			pampFiles = append(pampFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk ignore files: %w", err)
	}

	normalizeSort := func(paths []string) {
		slices.SortFunc(paths, func(a, b string) int {
			aRel, _ := filepath.Rel(root, a)
			bRel, _ := filepath.Rel(root, b)
			return strings.Compare(normalizeFromRelative(aRel), normalizeFromRelative(bRel))
		})
	}

	normalizeSort(gitFiles)
	normalizeSort(pampFiles)

	return gitFiles, pampFiles, nil
}

func parseIgnoreFile(root string, ignoreFile string, source RuleSource) ([]compiledRule, error) {
	file, err := os.Open(ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", ignoreFile, err)
	}
	defer file.Close()

	baseDirAbs := filepath.Dir(ignoreFile)
	baseDirRel, err := filepath.Rel(root, baseDirAbs)
	if err != nil {
		return nil, fmt.Errorf("resolve ignore base dir for %s: %w", ignoreFile, err)
	}
	baseDir := normalizeFromRelative(baseDirRel)
	if baseDir == "." {
		baseDir = ""
	}

	ignoreFileRel, err := filepath.Rel(root, ignoreFile)
	if err != nil {
		return nil, fmt.Errorf("resolve ignore file path %s: %w", ignoreFile, err)
	}

	var rules []compiledRule
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		rule, ok, compileErr := compileRule(line, source, baseDir, normalizeFromRelative(ignoreFileRel))
		if compileErr != nil {
			return nil, fmt.Errorf("compile %s:%d: %w", ignoreFile, lineNo, compileErr)
		}
		if !ok {
			continue
		}
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", ignoreFile, err)
	}

	return rules, nil
}

func compileRule(line string, source RuleSource, baseDir string, ignoreFile string) (compiledRule, bool, error) {
	raw := strings.TrimSpace(line)
	if raw == "" {
		return compiledRule{}, false, nil
	}

	if strings.HasPrefix(raw, "#") && !strings.HasPrefix(raw, `\#`) {
		return compiledRule{}, false, nil
	}

	raw = strings.ReplaceAll(raw, "\\\\", "/")
	domain := splitPathParts(baseDir)

	rule := compiledRule{
		source:     source,
		pattern:    line,
		ignoreFile: ignoreFile,
		matcher:    gitignore.ParsePattern(raw, domain),
	}

	return rule, true, nil
}

func lastMatchingRule(rules []compiledRule, relativePath string, isDir bool) (*compiledRule, bool) {
	var match *compiledRule
	negated := false
	for i := range rules {
		rule := &rules[i]
		ruleMatched, ruleNegated := ruleMatches(rule, relativePath, isDir)
		if !ruleMatched {
			continue
		}
		match = rule
		negated = ruleNegated
	}

	return match, negated
}

func ruleMatches(rule *compiledRule, relativePath string, isDir bool) (bool, bool) {
	parts := splitPathParts(relativePath)
	if len(parts) == 0 {
		return false, false
	}

	result := rule.matcher.Match(parts, isDir)
	switch result {
	case gitignore.Exclude:
		return true, false
	case gitignore.Include:
		return true, true
	default:
		return false, false
	}
}

func splitPathParts(relativePath string) []string {
	normalized := normalizeFromRelative(relativePath)
	if normalized == "" {
		return nil
	}

	parts := strings.Split(normalized, "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		filtered = append(filtered, part)
	}

	return filtered
}
