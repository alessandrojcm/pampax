package discovery

import "runtime"

type WarningCode string

const (
	WarningPermissionDenied WarningCode = "permission_denied"
	WarningReadDirFailed    WarningCode = "read_dir_failed"
	WarningStatFailed       WarningCode = "stat_failed"
	WarningBrokenSymlink    WarningCode = "broken_symlink"
)

type Warning struct {
	Code    WarningCode
	Path    string
	Message string
}

type Matcher interface {
	ShouldSkipDir(relativePath string) bool
	ShouldSkipFile(relativePath string) bool
}

type RuleSource string

const (
	RuleSourceNone       RuleSource = "none"
	RuleSourceDefault    RuleSource = "default"
	RuleSourceGitIgnore  RuleSource = "gitignore"
	RuleSourcePampIgnore RuleSource = "pampignore"
)

type IgnoreDecision struct {
	Path       string
	IsDir      bool
	Excluded   bool
	Matched    bool
	Source     RuleSource
	Pattern    string
	IgnoreFile string
	Negated    bool
}

type WalkOptions struct {
	Root          string
	Workers       int
	SupportedExts map[string]struct{}
	Matcher       Matcher
}

func (o WalkOptions) workerCount() int {
	if o.Workers > 0 {
		return o.Workers
	}

	count := runtime.NumCPU()
	if count < 1 {
		return 1
	}

	return count
}

type WalkResult struct {
	Paths    []string
	Warnings []Warning
}
