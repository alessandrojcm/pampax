# Chunk File Format Specification

**Version:** 1.0  
**Created:** 2026-01-28  
**Status:** Reference Specification for Go Port Compatibility

---

## Overview

This document specifies the exact format for chunk files stored in `.pampa/chunks/`. The Go implementation MUST produce byte-identical chunk files for the same source code content.

---

## Format Summary

| Property | Value |
|----------|-------|
| **Location** | `.pampa/chunks/` directory |
| **Naming** | `{SHA1_HASH}.gz` or `{SHA1_HASH}.gz.enc` |
| **Hash Algorithm** | SHA-1 (NOT SHA-256) |
| **Hash Input** | Raw code chunk content (UTF-8) |
| **Encoding** | UTF-8 |
| **Compression** | gzip (RFC 1952) |
| **Encryption** | Optional AES-256-GCM |
| **Line Endings** | Preserved from source |

---

## File Naming

### Plain (Unencrypted) Chunks

**Format:** `{sha1_hash}.gz`

**Example:**
```
5ea95a5a78779486d1fccdab927a7d64f5cf1599.gz
```

### Encrypted Chunks

**Format:** `{sha1_hash}.gz.enc`

**Example:**
```
5ea95a5a78779486d1fccdab927a7d64f5cf1599.gz.enc
```

### SHA-1 Hash Computation

**Node.js Reference Implementation:**

**Location:** `src/service.js:2178`

```javascript
const sha = crypto.createHash("sha1").update(code).digest("hex");
```

**Go Implementation:**

```go
import (
    "crypto/sha1"
    "encoding/hex"
)

func computeChunkSHA(code string) string {
    hasher := sha1.New()
    hasher.Write([]byte(code))
    return hex.EncodeToString(hasher.Sum(nil))
}
```

**Characteristics:**
- **Algorithm:** SHA-1 (160 bits = 40 hex characters)
- **Input:** Raw UTF-8 encoded code string
- **Output:** Lowercase hexadecimal string
- **Length:** Always 40 characters

**Example:**
```go
code := "export const foo = 'bar';"
sha := computeChunkSHA(code)
// sha = "5ea95a5a78779486d1fccdab927a7d64f5cf1599"
```

---

## Directory Structure

### Layout

```
.pampa/
└── chunks/
    ├── 5ea95a5a78779486d1fccdab927a7d64f5cf1599.gz
    ├── 640d912a6d1cce651dd0e1906414e07445a4fc91.gz
    ├── a681484f123456789abcdef0123456789abcdef0.gz
    └── ... (flat directory, no subdirectories)
```

### Characteristics

- **Flat structure:** All chunks in single directory (no subdirectories)
- **No collision handling:** SHA-1 provides sufficient uniqueness
- **File permissions:** Default file permissions (typically 0644)
- **Directory permissions:** 0755 (read/execute for all, write for owner)

---

## Plain Chunk Format

### Structure

```
┌─────────────────────────────────────┐
│  Gzip Header (10+ bytes)            │
├─────────────────────────────────────┤
│  Compressed Chunk Content           │
│  (gzip-compressed UTF-8 text)       │
├─────────────────────────────────────┤
│  Gzip Footer (8 bytes)              │
└─────────────────────────────────────┘
```

### Gzip Compression

**Standard:** RFC 1952 (gzip file format)

**Node.js Implementation:**

**Location:** `src/storage/encryptedChunks.js:190`

```javascript
const buffer = Buffer.isBuffer(code) ? code : Buffer.from(code, 'utf8');
const compressed = zlib.gzipSync(buffer);
fs.writeFileSync(plainPath, compressed);
```

**Go Implementation:**

```go
import (
    "compress/gzip"
    "io"
    "os"
)

func writeChunk(path, code string) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    gzWriter := gzip.NewWriter(file)
    defer gzWriter.Close()
    
    _, err = io.WriteString(gzWriter, code)
    return err
}
```

### Gzip Header Breakdown

**Example hex dump:**
```
00000000: 1f8b 0800 0000 0000 0013 8dd0 318b c240
          ^^^ ^^^ ^^^^ ^^^^ ^^^^ ^^^
          |   |   |    |    |    └─ Extra flags / OS
          |   |   |    |    └─────── Modification time
          |   |   |    └──────────── Flags
          |   |   └───────────────── Compression method (08 = deflate)
          |   └───────────────────── Magic number (8b1f)
          └───────────────────────── Magic number (1f8b)
```

**Key bytes:**
- `1f 8b`: Gzip magic number (always present)
- `08`: Deflate compression method
- `00 00 00 00`: Modification time (typically zero in PAMPAX)
- `00`: Extra flags
- `03`: OS (Unix)

**Compatibility:**
- Go's `compress/gzip` produces compatible output
- No custom gzip options needed
- Default compression level is sufficient

### Decompression

**Node.js:**
```javascript
const compressed = fs.readFileSync(plainPath);
const code = zlib.gunzipSync(compressed).toString('utf8');
```

**Go:**
```go
func readChunk(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    gzReader, err := gzip.NewReader(file)
    if err != nil {
        return "", err
    }
    defer gzReader.Close()
    
    content, err := io.ReadAll(gzReader)
    if err != nil {
        return "", err
    }
    
    return string(content), nil
}
```

---

## Encrypted Chunk Format

### Overview

When encryption is enabled (`PAMPAX_ENCRYPTION_KEY` set), chunks are encrypted using AES-256-GCM before storage.

### File Extension

- **Encrypted:** `.gz.enc` (NOT `.gz`)
- **Unencrypted:** `.gz`

**Mutual Exclusion:**
- Only ONE file exists per SHA: either `.gz` OR `.gz.enc`
- When encryption is enabled, `.gz` is deleted and replaced with `.gz.enc`
- When encryption is disabled, `.gz.enc` is deleted and replaced with `.gz`

### Encryption Structure

```
┌─────────────────────────────────────┐
│  Magic Header (7 bytes)             │  "PAMPAE1"
├─────────────────────────────────────┤
│  Salt (16 bytes)                    │  Random bytes for key derivation
├─────────────────────────────────────┤
│  IV (12 bytes)                      │  Initialization vector (random)
├─────────────────────────────────────┤
│  Ciphertext (variable length)       │  Encrypted gzipped content
├─────────────────────────────────────┤
│  Auth Tag (16 bytes)                │  GCM authentication tag
└─────────────────────────────────────┘
```

**Total overhead:** 7 + 16 + 12 + 16 = 51 bytes

### Encryption Specification

**Reference Implementation:** `src/storage/encryptedChunks.js`

**Algorithm:** AES-256-GCM

**Key Derivation:**
- **Algorithm:** HKDF-SHA256
- **Master Key:** 32-byte key from `PAMPAX_ENCRYPTION_KEY` (base64 or hex)
- **Salt:** 16 random bytes (per chunk)
- **Info:** `"pampa-chunk-v1"` (UTF-8 encoded)
- **Output:** 32-byte derived key

**Node.js Reference:**

```javascript
// Location: src/storage/encryptedChunks.js:125-127
function deriveChunkKey(masterKey, salt) {
    return crypto.hkdfSync('sha256', masterKey, salt, HKDF_INFO, 32);
}
```

**Go Implementation:**

```go
import (
    "crypto/sha256"
    "golang.org/x/crypto/hkdf"
    "io"
)

func deriveChunkKey(masterKey, salt []byte) ([]byte, error) {
    info := []byte("pampa-chunk-v1")
    reader := hkdf.New(sha256.New, masterKey, salt, info)
    
    derivedKey := make([]byte, 32)
    if _, err := io.ReadFull(reader, derivedKey); err != nil {
        return nil, err
    }
    
    return derivedKey, nil
}
```

### Encryption Process

**Node.js Reference:** `src/storage/encryptedChunks.js:129-140`

```javascript
function encryptBuffer(plaintext, masterKey) {
    const salt = crypto.randomBytes(SALT_LENGTH);        // 16 bytes
    const iv = crypto.randomBytes(IV_LENGTH);            // 12 bytes
    const derivedKey = deriveChunkKey(masterKey, salt);  // 32 bytes

    const cipher = crypto.createCipheriv('aes-256-gcm', derivedKey, iv);
    const encrypted = Buffer.concat([cipher.update(plaintext), cipher.final()]);
    const tag = cipher.getAuthTag();                     // 16 bytes

    const payload = Buffer.concat([MAGIC_HEADER, salt, iv, encrypted, tag]);
    return { payload, salt, iv, tag };
}
```

**Input:** Gzipped chunk content  
**Output:** Encrypted payload with header

**Go Implementation:**

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
)

func encryptChunk(gzippedContent, masterKey []byte) ([]byte, error) {
    // Generate random salt and IV
    salt := make([]byte, 16)
    iv := make([]byte, 12)
    if _, err := rand.Read(salt); err != nil {
        return nil, err
    }
    if _, err := rand.Read(iv); err != nil {
        return nil, err
    }
    
    // Derive key
    derivedKey, err := deriveChunkKey(masterKey, salt)
    if err != nil {
        return nil, err
    }
    
    // Create AES-GCM cipher
    block, err := aes.NewCipher(derivedKey)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    // Encrypt
    ciphertext := gcm.Seal(nil, iv, gzippedContent, nil)
    
    // Build payload: MAGIC_HEADER + salt + iv + ciphertext + tag
    // Note: gcm.Seal() appends the tag automatically
    magicHeader := []byte("PAMPAE1")
    payload := make([]byte, 0, len(magicHeader)+16+12+len(ciphertext))
    payload = append(payload, magicHeader...)
    payload = append(payload, salt...)
    payload = append(payload, iv...)
    payload = append(payload, ciphertext...)  // Includes tag
    
    return payload, nil
}
```

### Decryption Process

**Node.js Reference:** `src/storage/encryptedChunks.js:142-179`

```javascript
function decryptBuffer(payload, masterKey) {
    const minimumLength = MAGIC_HEADER.length + SALT_LENGTH + IV_LENGTH + TAG_LENGTH + 1;
    if (!payload || payload.length < minimumLength) {
        throw new Error('Encrypted chunk payload is truncated.');
    }

    const header = payload.subarray(0, MAGIC_HEADER.length);
    if (!header.equals(MAGIC_HEADER)) {
        throw new Error('Encrypted chunk payload has an unknown header.');
    }

    const saltStart = MAGIC_HEADER.length;
    const ivStart = saltStart + SALT_LENGTH;
    const cipherStart = ivStart + IV_LENGTH;
    const cipherEnd = payload.length - TAG_LENGTH;

    const salt = payload.subarray(saltStart, saltStart + SALT_LENGTH);
    const iv = payload.subarray(ivStart, ivStart + IV_LENGTH);
    const ciphertext = payload.subarray(cipherStart, cipherEnd);
    const tag = payload.subarray(cipherEnd);

    const derivedKey = deriveChunkKey(masterKey, salt);
    const decipher = crypto.createDecipheriv('aes-256-gcm', derivedKey, iv);
    decipher.setAuthTag(tag);

    try {
        return Buffer.concat([decipher.update(ciphertext), decipher.final()]);
    } catch (error) {
        throw new Error('authentication failed');
    }
}
```

**Go Implementation:**

```go
func decryptChunk(payload, masterKey []byte) ([]byte, error) {
    magicHeader := []byte("PAMPAE1")
    minimumLength := len(magicHeader) + 16 + 12 + 16 + 1
    
    if len(payload) < minimumLength {
        return nil, errors.New("payload truncated")
    }
    
    // Validate magic header
    if !bytes.Equal(payload[:7], magicHeader) {
        return nil, errors.New("invalid magic header")
    }
    
    // Extract components
    salt := payload[7:23]
    iv := payload[23:35]
    ciphertext := payload[35:]  // Includes auth tag (last 16 bytes)
    
    // Derive key
    derivedKey, err := deriveChunkKey(masterKey, salt)
    if err != nil {
        return nil, err
    }
    
    // Create AES-GCM cipher
    block, err := aes.NewCipher(derivedKey)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    // Decrypt and verify
    plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
    if err != nil {
        return nil, fmt.Errorf("authentication failed: %w", err)
    }
    
    return plaintext, nil
}
```

### Master Key Format

**Environment Variable:** `PAMPAX_ENCRYPTION_KEY`

**Accepted Formats:**
1. **Base64:** 32-byte key encoded as base64 (44 characters)
2. **Hex:** 32-byte key encoded as hexadecimal (64 characters)

**Example:**
```bash
# Base64 (recommended)
export PAMPAX_ENCRYPTION_KEY="rL8/vZ3x9Kq2mP4nW7sT1eR6uY0hG5jA8bC3dE9fF1g="

# Hex
export PAMPAX_ENCRYPTION_KEY="acbf3fbd9df1f4aab698fe2757bb13d5e47ab98d2107b8648f1b0b775f07d75a"
```

**Generation:**
```bash
# Generate 32 random bytes and encode as base64
openssl rand -base64 32
```

---

## Content Encoding

### UTF-8 Requirement

**All chunk files MUST use UTF-8 encoding:**
- Source code is read as UTF-8
- Gzip compresses UTF-8 bytes
- Decompression yields UTF-8 bytes
- Final string is UTF-8 decoded

### Line Ending Preservation

**Node.js behavior:**
- Line endings are **preserved exactly as-is** from source
- No normalization (Windows `\r\n` stays `\r\n`, Unix `\n` stays `\n`)
- Gzip compresses bytes verbatim

**Go compatibility:**
- Must preserve exact bytes
- Do NOT normalize line endings
- Use `os.ReadFile()` / `os.WriteFile()` without conversion

**Example:**
```javascript
// Source file (Windows)
"function foo() {\r\n  return 42;\r\n}"

// Stored chunk (exact same)
"function foo() {\r\n  return 42;\r\n}"
```

### Special Characters

**Non-ASCII characters:**
- Fully supported via UTF-8
- Example: `const 日本語 = '値';`
- Gzip handles multi-byte UTF-8 correctly

**BOM (Byte Order Mark):**
- If source has BOM, it's preserved
- If source has no BOM, none is added
- Typical behavior: no BOM in modern editors

---

## File Lifecycle

### Creation

1. **Extract chunk code** from source file
2. **Compute SHA-1 hash** of chunk code (UTF-8 bytes)
3. **Gzip compress** chunk code
4. **Optional: Encrypt** gzipped content (if `PAMPAX_ENCRYPTION_KEY` set)
5. **Write to disk:** `.pampa/chunks/{sha}.gz` or `.pampa/chunks/{sha}.gz.enc`

### Retrieval

1. **Determine file type:** Check for `.gz.enc` first, then `.gz`
2. **Read file** from disk
3. **Optional: Decrypt** if `.gz.enc`
4. **Gunzip decompress** to get original code
5. **Return UTF-8 string**

### Deletion

**When chunk is removed from index:**
1. Delete from `code_chunks` table
2. Remove from `pampa.codemap.json`
3. **Delete chunk file:** `.gz` and/or `.gz.enc`

**Node.js Reference:** `src/storage/encryptedChunks.js:260-268`

```javascript
export function removeChunkArtifacts(chunkDir, sha) {
    const { plainPath, encryptedPath } = getChunkPaths(chunkDir, sha);
    if (fs.existsSync(plainPath)) {
        fs.rmSync(plainPath, { force: true });
    }
    if (fs.existsSync(encryptedPath)) {
        fs.rmSync(encryptedPath, { force: true });
    }
}
```

### Encryption Toggle

**When toggling encryption:**
1. Read existing chunk (`.gz` or `.gz.enc`)
2. Decompress (and decrypt if needed)
3. Recompress
4. Encrypt (if new mode is encrypted)
5. Write new file
6. **Delete old file**

---

## Edge Cases

### 1. Duplicate SHAs

**Problem:** Two different chunks with same SHA

**Solution:**
- SHA-1 collisions are extremely rare in practice
- Node.js implementation **overwrites** existing file
- Go should match: overwrite silently

### 2. Orphaned Chunks

**Problem:** Chunk file exists but not in database

**Solution:**
- Can occur after failed indexing
- Safe to delete (garbage collection)
- Future: implement `pampax gc` command

### 3. Missing Chunks

**Problem:** Database references SHA but file doesn't exist

**Solution:**
- Treat as error during retrieval
- Re-index file to regenerate chunk
- Node.js throws error: `"Failed to read chunk {sha}"`

### 4. Partial Writes

**Problem:** Power loss during write

**Solution:**
- Atomic write using temp file + rename:
  ```go
  tmpFile := path + ".tmp"
  os.WriteFile(tmpFile, data, 0644)
  os.Rename(tmpFile, path)  // Atomic on POSIX
  ```

### 5. Encryption Key Rotation

**Problem:** Need to change encryption key

**Solution:**
1. Index with new key → creates `.gz.enc` with new key
2. Old chunks fail authentication
3. **Full re-index required** (no migration path)

### 6. Platform Path Separators

**Problem:** Directory name uses `/` or `\`

**Solution:**
- `.pampa/chunks` uses forward slash on **all platforms**
- Go: use `filepath.Join(".pampa", "chunks")` (cross-platform)

---

## Testing and Validation

### Roundtrip Test

**Test:** Write → Read → Compare

```go
func TestChunkRoundtrip(t *testing.T) {
    original := "export const foo = 'bar';"
    
    // Compute SHA
    sha := computeChunkSHA(original)
    
    // Write
    path := filepath.Join(chunkDir, sha+".gz")
    err := writeChunk(path, original)
    assert.NoError(t, err)
    
    // Read
    retrieved, err := readChunk(path)
    assert.NoError(t, err)
    
    // Compare
    assert.Equal(t, original, retrieved)
}
```

### Encryption Roundtrip Test

```go
func TestEncryptedChunkRoundtrip(t *testing.T) {
    original := "export const secret = 'password';"
    masterKey := generateRandomKey(32)
    
    // Compute SHA
    sha := computeChunkSHA(original)
    
    // Write encrypted
    path := filepath.Join(chunkDir, sha+".gz.enc")
    err := writeEncryptedChunk(path, original, masterKey)
    assert.NoError(t, err)
    
    // Read encrypted
    retrieved, err := readEncryptedChunk(path, masterKey)
    assert.NoError(t, err)
    
    // Compare
    assert.Equal(t, original, retrieved)
}
```

### Cross-Implementation Test

**Test:** Node.js writes → Go reads

```bash
# Node.js writes chunk
node -e "require('./src/service.js').indexProject({repoPath: './test-repo'})"

# Go reads chunk
go test -run TestReadNodeChunk
```

---

## Compatibility Checklist

- [ ] SHA-1 hash (NOT SHA-256)
- [ ] Gzip compression (RFC 1952)
- [ ] UTF-8 encoding
- [ ] Line endings preserved exactly
- [ ] File naming: `{sha}.gz` or `{sha}.gz.enc`
- [ ] Flat directory structure (no subdirectories)
- [ ] Encrypted format: PAMPAE1 magic header
- [ ] AES-256-GCM encryption
- [ ] HKDF-SHA256 key derivation
- [ ] Master key: base64 or hex (32 bytes)
- [ ] Atomic writes (temp file + rename)
- [ ] Proper error handling for missing chunks
- [ ] Deletion removes both `.gz` and `.gz.enc`

---

## Reference Implementation Locations

**Node.js:**
- Chunk writing: `src/storage/encryptedChunks.js:187-206`
- Chunk reading: `src/storage/encryptedChunks.js:208-258`
- Encryption: `src/storage/encryptedChunks.js:129-140`
- Decryption: `src/storage/encryptedChunks.js:142-179`
- SHA computation: `src/service.js:2178`

**Related:**
- Encryption key handling: `src/storage/encryptedChunks.js:54-79`
- Key derivation: `src/storage/encryptedChunks.js:125-127`

---

## Summary

Chunk files are:
1. ✅ SHA-1 named (40 hex chars)
2. ✅ Gzip compressed UTF-8 text
3. ✅ Optionally AES-256-GCM encrypted
4. ✅ Stored in flat `.pampa/chunks/` directory
5. ✅ Line-ending preserving
6. ✅ Cross-platform compatible

**Critical for Compatibility:**
- Use SHA-1 (NOT SHA-256)
- Preserve exact line endings
- Encrypt with AES-256-GCM + HKDF
- Magic header: `PAMPAE1`
- Atomic writes for reliability
