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

func DefaultSupportedExtensions() map[string]struct{} {
	exts := make(map[string]struct{}, len(defaultLanguageExtensions))
	for _, ext := range defaultLanguageExtensions {
		exts[strings.ToLower(ext)] = struct{}{}
	}

	return exts
}
