# Edge Cases and Platform Differences

**Version:** 1.0  
**Created:** 2026-01-28  
**Status:** Reference Specification for Go Port Compatibility

---

## Overview

This document catalogs all edge cases, platform-specific behaviors, and special handling requirements for the Go port to maintain full compatibility with the Node.js implementation.

---

## 1. Path Handling

### 1.1 Path Separators

**Problem:** Windows uses `\`, Unix/macOS uses `/`

#### Database Storage (code_chunks.file_path)

**Node.js Behavior:**
- Paths **normalized to forward slashes** before storage
- Example: `src\utils\logger.js` → `src/utils/logger.js`

**Go Requirements:**
```go
import (
    "path/filepath"
    "strings"
)

// Normalize path for storage
func normalizePathForStorage(p string) string {
    // Convert to forward slashes
    normalized := filepath.ToSlash(p)
    
    // Remove leading ./ if present
    normalized = strings.TrimPrefix(normalized, "./")
    
    return normalized
}
```

**Examples:**
```
Windows: src\utils\logger.js → src/utils/logger.js
Unix:    src/utils/logger.js → src/utils/logger.js
macOS:   src/utils/logger.js → src/utils/logger.js
```

#### Chunk Directory Access

**Node.js Behavior:**
- `.pampa/chunks` directory accessed with platform-native separators
- File operations use OS-native paths

**Go Requirements:**
```go
// Reading/writing chunk files
chunkDir := filepath.Join(projectRoot, ".pampa", "chunks")
chunkPath := filepath.Join(chunkDir, sha+".gz")
```

**DO NOT hardcode:**
```go
// ❌ BAD - platform-specific
chunkPath := projectRoot + "/.pampa/chunks/" + sha + ".gz"

// ✅ GOOD - cross-platform
chunkPath := filepath.Join(projectRoot, ".pampa", "chunks", sha+".gz")
```

### 1.2 Absolute vs Relative Paths

**Policy:** Store **relative paths only**

**Node.js Implementation:**
- `normalizeToProjectPath()` in `src/indexer/merkle.js:46-62`
- Converts absolute paths to project-relative

**Go Implementation:**
```go
func normalizeToProjectPath(projectRoot, absPath string) (string, error) {
    relPath, err := filepath.Rel(projectRoot, absPath)
    if err != nil {
        return "", err
    }
    
    // Normalize to forward slashes for storage
    return normalizePathForStorage(relPath), nil
}
```

**Examples:**
```
Project root: /home/user/myproject
File path:    /home/user/myproject/src/main.go
Stored as:    src/main.go
```

### 1.3 Symbolic Links

**Problem:** Symlinks may point outside project

**Node.js Behavior:**
- `fast-glob` option: `followSymbolicLinks: false`
- Symlinks are **not followed** during indexing
- Only real files are indexed

**Go Requirements:**
```go
// When globbing files
func shouldFollowSymlinks() bool {
    return false  // Match Node.js behavior
}

// When checking file existence
func isSymlink(path string) (bool, error) {
    info, err := os.Lstat(path)  // Lstat doesn't follow symlinks
    if err != nil {
        return false, err
    }
    return info.Mode()&os.ModeSymlink != 0, nil
}
```

**Resolution:**
- If symlink is resolved, store the **canonical path** (relative to project)
- Use `filepath.EvalSymlinks()` if symlinks MUST be followed

### 1.4 Case Sensitivity

**Problem:** Windows/macOS are case-insensitive, Linux is case-sensitive

**Node.js Behavior:**
- Stores paths **as-is** from file system
- No case normalization
- `Foo.js` and `foo.js` are **different** on Linux, **same** on Windows

**Go Requirements:**
- **Do NOT normalize case**
- Store exact path from `filepath.Walk()` or `filepath.Glob()`
- Let OS handle case-sensitivity

**Example:**
```go
// ✅ GOOD - preserve case
filePath := "src/MyComponent.tsx"  // Stored as-is

// ❌ BAD - case normalization
filePath := strings.ToLower("src/MyComponent.tsx")  // ❌ Don't do this
```

**Edge Case: Case-only renames**

```
# On Linux
git mv Foo.js foo.js  # Creates two different files

# On Windows/macOS
git mv Foo.js foo.js  # Renames the same file
```

**Handling:**
- Trust file system reports
- On Windows/macOS: case-only rename triggers update
- On Linux: both files may exist simultaneously

---

## 2. Line Endings

### 2.1 File Reading

**Problem:** Windows uses `\r\n`, Unix/macOS uses `\n`

**Node.js Behavior:**
- `fs.readFileSync(path, 'utf8')` returns **exact bytes**
- No line ending conversion
- Chunk content preserves original line endings

**Go Requirements:**
```go
// ✅ GOOD - preserves line endings
content, err := os.ReadFile(path)
text := string(content)  // Preserves \r\n or \n

// ❌ BAD - converts line endings
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()  // Strips \r\n → \n
}
```

**Critical:**
- Use `os.ReadFile()` or `io.ReadAll()` (preserves bytes)
- Do NOT use `bufio.Scanner` (strips line endings)
- Do NOT use `strings.ReplaceAll("\r\n", "\n")`

### 2.2 SHA Calculation

**Problem:** SHA must be computed on **original bytes**

**Node.js Reference:**
```javascript
// src/service.js:2178
const sha = crypto.createHash("sha1").update(code).digest("hex");
// `code` contains original \r\n if present
```

**Go Requirements:**
```go
func computeChunkSHA(code string) string {
    hasher := sha1.New()
    hasher.Write([]byte(code))  // Preserve \r\n
    return hex.EncodeToString(hasher.Sum(nil))
}
```

**Test:**
```go
code := "function foo() {\r\n  return 42;\r\n}"
sha := computeChunkSHA(code)
// SHA must match Node.js SHA for same \r\n content
```

### 2.3 JSON Serialization Line Endings

**Requirement:**
- JSON artifacts must use Unix line endings (`\n`) on all platforms.

---

## 3. Character Encoding

### 3.1 UTF-8 Requirement

**All text files MUST be UTF-8:**
- Source code files
- Chunk files (before gzip)
- Codemap JSON
- SQLite text columns

**Node.js Behavior:**
- `fs.readFileSync(path, 'utf8')` assumes UTF-8
- Fails gracefully if invalid UTF-8 (replacement characters)

**Go Requirements:**
```go
// Reading with UTF-8 validation
import "unicode/utf8"

func readFileUTF8(path string) (string, error) {
    bytes, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    
    // Validate UTF-8
    if !utf8.Valid(bytes) {
        return "", fmt.Errorf("file is not valid UTF-8: %s", path)
    }
    
    return string(bytes), nil
}
```

**Edge Case: BOM (Byte Order Mark)**

**Problem:** UTF-8 BOM (`EF BB BF`) at start of file

**Node.js Behavior:**
- BOM is **preserved** in chunk content
- SHA includes BOM

**Go Requirements:**
- **Do NOT strip BOM**
- Preserve exact bytes
- SHA must include BOM

### 3.2 Non-ASCII Filenames

**Problem:** Unicode characters in file paths

**Examples:**
```
src/测试/test.js
src/café/utils.js
src/файл.py
```

**Node.js Behavior:**
- Fully supported via UTF-8 encoding
- Stored in database and codemap as UTF-8

**Go Requirements:**
```go
// Paths are UTF-8 strings (Go's default)
filePath := "src/测试/test.js"  // ✅ Works

// SQLite stores as UTF-8 TEXT
db.Exec("INSERT INTO code_chunks (file_path, ...) VALUES (?, ...)", filePath)
```

**Platform Notes:**
- **Linux/macOS:** UTF-8 native (no issues)
- **Windows:** UTF-16 native, but Go's `os` package handles conversion

### 3.3 Surrogate Pairs and Emoji

**Problem:** Multi-byte UTF-8 characters (emoji, etc.)

**Examples:**
```javascript
const icon = "🚀";  // 4-byte UTF-8 character
const text = "Hello 👋";
```

**Node.js Behavior:**
- Full support (JavaScript strings are UTF-16)
- Correctly counts grapheme clusters

**Go Requirements:**
```go
// Go strings are UTF-8 byte slices
text := "Hello 👋"
len(text)  // Returns 10 (bytes), not 7 (characters)

// Count runes (Unicode code points)
runeCount := utf8.RuneCountInString(text)  // Returns 7
```

**Critical:**
- Use `utf8.RuneCountInString()` for character counting
- Use `range` over strings (iterates runes, not bytes)

---

## 4. Timestamps

### 4.1 SQLite CURRENT_TIMESTAMP

**Format:** ISO 8601 without timezone suffix

**Node.js Output:**
```sql
created_at = '2026-01-28 14:30:00'
```

**Go Requirements:**
```go
// SQLite CURRENT_TIMESTAMP is UTC
// No 'Z' suffix, no timezone offset

// Reading timestamp
var timestamp string
row.Scan(&timestamp)  // "2026-01-28 14:30:00"

// Parsing
t, err := time.Parse("2006-01-02 15:04:05", timestamp)
// Assume UTC
t = t.UTC()
```

### 4.2 Codemap last_used_at

**Format:** ISO 8601 with `Z` suffix (explicit UTC)

**Node.js Output:**
```json
"last_used_at": "2026-01-28T14:30:00.000Z"
```

**Go Requirements:**
```go
// Generating timestamp for codemap
now := time.Now().UTC()
timestamp := now.Format("2006-01-02T15:04:05.000Z07:00")
// Result: "2026-01-28T14:30:00.000Z"
```

**Critical:**
- Database timestamps: **no Z suffix**
- Codemap timestamps: **with Z suffix**
- Always UTC (never local timezone)

---

## 5. SQLite Specifics

### 5.1 Pragmas

**Required Settings:**
```sql
PRAGMA encoding = 'UTF-8';
PRAGMA journal_mode = delete;
```

---

## 6. File System Operations

### 8.1 Atomic Writes

**Problem:** Partial writes on crash/power loss

**Node.js:**
- `fs.writeFileSync()` is **not atomic** on all platforms

**Go Solution:**
```go
func atomicWrite(path string, data []byte) error {
    // Write to temp file
    tmpPath := path + ".tmp"
    if err := os.WriteFile(tmpPath, data, 0644); err != nil {
        return err
    }
    
    // Atomic rename
    return os.Rename(tmpPath, path)  // Atomic on POSIX
}
```

**Note:**
- Atomic on POSIX (Linux, macOS)
- **Not guaranteed atomic on Windows** (but usually is)

### 8.2 Directory Creation

**Node.js:**
```javascript
fs.mkdirSync(dir, { recursive: true })
```

**Go:**
```go
os.MkdirAll(dir, 0755)
```

### 8.3 File Deletion

**Node.js:**
```javascript
fs.rmSync(path, { force: true })  // No error if not exists
```

**Go:**
```go
os.Remove(path)  // Error if not exists

// Force (ignore errors):
_ = os.Remove(path)

// Or check existence first:
if _, err := os.Stat(path); err == nil {
    os.Remove(path)
}
```

---

## 7. Error Handling

### 10.1 Missing Chunk Files

**Node.js:**
```javascript
if (!fs.existsSync(chunkPath)) {
    return null;  // Or throw error
}
```

**Go:**
```go
if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
    return nil, fmt.Errorf("chunk not found: %s", sha)
}
```

### 10.2 Encrypted Chunk Without Key

**Node.js:**
```javascript
throw new Error(`Chunk ${sha} is encrypted and no PAMPAX_ENCRYPTION_KEY is configured.`)
```

**Go:**
```go
if isEncrypted && masterKey == nil {
    return nil, fmt.Errorf("chunk %s is encrypted but no PAMPAX_ENCRYPTION_KEY set", sha)
}
```

---

## 8. Testing Edge Cases

### Test Matrix

| Case | Node.js | Go | Notes |
|------|---------|-----|-------|
| Windows path `src\file.js` | → `src/file.js` | Same | Normalize to `/` |
| Unix path `src/file.js` | → `src/file.js` | Same | Already `/` |
| Line endings `\r\n` | Preserved | Must preserve | SHA must match |
| Line endings `\n` | Preserved | Must preserve | SHA must match |
| UTF-8 filename `测试.js` | Supported | Must support | UTF-8 encoding |
| Empty symbol | `symbol=""` | Same | Empty string, not NULL |
| NULL tags | `pampa_tags=NULL` | Same | NULL, not empty string |
| Gzip default compression | Level 6 | Must match | Use `DefaultCompression` |
| JSON key order | Alphabetical | Must match | Sort keys |
| Timestamp (DB) | `YYYY-MM-DD HH:MM:SS` | Must match | No `Z` |
| Timestamp (codemap) | `YYYY-MM-DDTHH:MM:SS.SSSZ` | Must match | With `Z` |

---

## 9. Compatibility Validation

### Cross-Implementation Tests

**Test 1: Path Normalization**
```go
assert.Equal(
    normalizePathForStorage("src\\utils\\logger.js"),
    "src/utils/logger.js",
)
```

**Test 2: Line Ending Preservation**
```go
code := "function foo() {\r\n  return 42;\r\n}"
sha := computeChunkSHA(code)
// SHA must match Node.js SHA for same content
```

**Test 3: JSON Key Ordering**
```go
codemap := Codemap{
    "b": ChunkMetadata{},
    "a": ChunkMetadata{},
}
json, _ := json.Marshal(codemap)
// Must start with "a" entry (alphabetical)
```

**Test 4: UTF-8 Filenames**
```go
path := "src/测试/test.js"
normalized := normalizePathForStorage(path)
assert.Equal(normalized, "src/测试/test.js")
```

---

## Summary Checklist

### Platform Compatibility
- [ ] Path separators normalized to `/` for storage
- [ ] Native separators used for file system operations
- [ ] Symlinks not followed during indexing
- [ ] Case preservation (no normalization)

### Line Endings
- [ ] Original line endings preserved in chunks
- [ ] SHA computed on original bytes (with `\r\n`)
- [ ] Codemap JSON uses Unix `\n` line endings

### Character Encoding
- [ ] UTF-8 validation for all text files
- [ ] BOM preserved if present
- [ ] Non-ASCII filenames supported
- [ ] Emoji and multi-byte characters handled

### Timestamps
- [ ] Database: `YYYY-MM-DD HH:MM:SS` (no Z)
- [ ] Codemap: `YYYY-MM-DDTHH:MM:SS.SSSZ` (with Z)
- [ ] Always UTC (never local timezone)

### Error Handling
- [ ] Missing chunks return error
- [ ] Encrypted chunks without key return error
- [ ] Invalid UTF-8 returns error
- [ ] Graceful handling of filesystem errors

---

**End of Edge Cases Documentation**
