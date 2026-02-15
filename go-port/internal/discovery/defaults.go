package discovery

import "strings"

var defaultLanguageExtensions = []string{
	".php",
	".py",
	".js",
	".jsx",
	".ts",
	".tsx",
	".go",
	".java",
	".cs",
	".rs",
	".rb",
	".cpp",
	".hpp",
	".cc",
	".c",
	".h",
	".scala",
	".swift",
	".sh",
	".bash",
	".kt",
	".lua",
	".html",
	".htm",
	".css",
	".json",
	".ml",
	".mli",
	".hs",
	".ex",
	".exs",
	".md",
	".markdown",
}

var defaultIgnorePatterns = []string{
	"**/vendor/**",
	"**/node_modules/**",
	"**/.git/**",
	"**/storage/**",
	"**/dist/**",
	"**/build/**",
	"**/tmp/**",
	"**/temp/**",
	"**/.npm/**",
	"**/.yarn/**",
	"**/Library/**",
	"**/System/**",
	"**/.Trash/**",
	"**/.pampa/**",
	"**/pampa.codemap.json",
	"**/pampa.codemap.json.backup-*",
	"**/package-lock.json",
	"**/yarn.lock",
	"**/pnpm-lock.yaml",
	"**/*.json",
	"**/*.sh",
	"**/examples/**",
	"**/assets/**",
}

func DefaultSupportedExtensions() map[string]struct{} {
	exts := make(map[string]struct{}, len(defaultLanguageExtensions))
	for _, ext := range defaultLanguageExtensions {
		exts[strings.ToLower(ext)] = struct{}{}
	}

	return exts
}

func DefaultIgnorePatterns() []string {
	patterns := make([]string, len(defaultIgnorePatterns))
	copy(patterns, defaultIgnorePatterns)
	return patterns
}
