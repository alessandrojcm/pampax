package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

func Walk(options WalkOptions) (WalkResult, error) {
	rootPath, err := filepath.Abs(options.Root)
	if err != nil {
		return WalkResult{}, fmt.Errorf("resolve absolute root path: %w", err)
	}

	rootInfo, err := os.Stat(rootPath)
	if err != nil {
		return WalkResult{}, fmt.Errorf("stat root path: %w", err)
	}

	if !rootInfo.IsDir() {
		return WalkResult{}, fmt.Errorf("root path is not a directory: %s", rootPath)
	}

	workerCount := options.workerCount()
	if len(options.SupportedExts) == 0 {
		options.SupportedExts = DefaultSupportedExtensions()
	}

	var (
		pathsMu    sync.Mutex
		warningsMu sync.Mutex
		paths      []string
		warnings   []Warning
	)

	appendPath := func(path string) {
		pathsMu.Lock()
		paths = append(paths, path)
		pathsMu.Unlock()
	}

	appendWarning := func(w Warning) {
		warningsMu.Lock()
		warnings = append(warnings, w)
		warningsMu.Unlock()
	}

	dirs := make(chan string, workerCount)
	var dirQueue sync.WaitGroup
	var workers sync.WaitGroup

	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()

			for dirPath := range dirs {
				entries, readErr := os.ReadDir(dirPath)
				if readErr != nil {
					relativePath, relErr := normalizeRelativePath(rootPath, dirPath)
					if relErr != nil {
						relativePath = normalizeFromRelative(dirPath)
					}

					appendWarning(classifyReadDirError(relativePath, readErr))
					dirQueue.Done()
					continue
				}

				slices.SortFunc(entries, func(a os.DirEntry, b os.DirEntry) int {
					return strings.Compare(a.Name(), b.Name())
				})

				for _, entry := range entries {
					fullPath := filepath.Join(dirPath, entry.Name())
					relativePath, relErr := normalizeRelativePath(rootPath, fullPath)
					if relErr != nil {
						appendWarning(Warning{
							Code:    WarningStatFailed,
							Path:    normalizeFromRelative(fullPath),
							Message: fmt.Sprintf("failed to normalize path: %v", relErr),
						})
						continue
					}

					entryType := entry.Type()
					if entryType&os.ModeSymlink != 0 {
						if _, statErr := os.Stat(fullPath); statErr != nil {
							appendWarning(classifyStatError(relativePath, statErr))
						}
						continue
					}

					if entry.IsDir() {
						if options.Matcher != nil && options.Matcher.ShouldSkipDir(relativePath) {
							continue
						}

						dirQueue.Add(1)
						dirs <- fullPath
						continue
					}

					if options.Matcher != nil && options.Matcher.ShouldSkipFile(relativePath) {
						continue
					}

					ext := strings.ToLower(filepath.Ext(entry.Name()))
					if _, ok := options.SupportedExts[ext]; !ok {
						continue
					}

					appendPath(relativePath)
				}

				dirQueue.Done()
			}
		}()
	}

	dirQueue.Add(1)
	dirs <- rootPath

	go func() {
		dirQueue.Wait()
		close(dirs)
	}()

	workers.Wait()

	slices.Sort(paths)
	slices.SortFunc(warnings, func(a Warning, b Warning) int {
		if pathCmp := strings.Compare(a.Path, b.Path); pathCmp != 0 {
			return pathCmp
		}

		if codeCmp := strings.Compare(string(a.Code), string(b.Code)); codeCmp != 0 {
			return codeCmp
		}

		return strings.Compare(a.Message, b.Message)
	})

	return WalkResult{Paths: paths, Warnings: warnings}, nil
}
