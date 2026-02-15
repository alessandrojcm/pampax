# Stage 2 Discovery Contract

**Version:** 1.0  
**Date:** 2026-02-15  
**Scope:** Stage 2.1 contract freeze for file discovery parity

---

## 1. Locked Decisions (Inherited)

This contract inherits and freezes Stage 2 plan points 1-3:

1. Ignore precedence: `default rules < .gitignore < .pampignore`
2. Symlinks: do not follow during traversal
3. Hidden files: no extra hidden-file auto-skip beyond Node defaults and ignore rules
4. Parity comparison paths: repo-relative with `/` separators on all OSes
5. Permission errors: skip + warning + continue
6. Default ignore set: frozen from current Node implementation for this stage

Success criteria from the plan are also locked for this stage:

- Parity: file lists match Node output exactly on Stage 2 fixtures
- Determinism: repeated runs are byte-identical
- Cross-platform: macOS/Linux/Windows parity checks pass
- Performance: Go walker throughput is at least 1.5x Node walker

---

## 2. Discovery Contract (Behavior)

### Include semantics

- Candidate files are language-matched by the extension set currently defined in `LANG_RULES` (`src/service.js`) and mirrored by the Stage 2 baseline harness.
- Discovery output contains only files (not directories).
- Paths are repo-relative.

### Exclude semantics

- Default ignore patterns are applied during discovery.
- For parity output, the default ignore list is treated as frozen for Stage 2.
- Symlink targets are never traversed (`followSymbolicLinks: false`).

### Ignore precedence and negation

- Contract precedence is fixed to: defaults, then `.gitignore`, then `.pampignore`.
- Negation behavior (`!pattern`) and re-ignore chains are required semantics for Stage 2 fixtures.
- Rule type behavior is required for Stage 2 fixtures:
  - directory-only patterns (`foo/`)
  - anchored patterns (`/foo`)
  - recursive patterns (`**/foo` semantics)
  - escaped special characters

Note: the baseline harness in Stage 2.1 captures current Node discovery output for reproducible comparison. Full ignore-file parity behavior is implemented and verified in later Stage 2 tasks.

---

## 3. Canonical Output Format

All Stage 2 parity comparisons must use this exact format:

1. UTF-8 plain text
2. One normalized relative path per line
3. Forward slash (`/`) separators on all platforms
4. Sorted ascending
5. Stable and byte-identical across repeated runs
6. Trailing newline at end of file

Example:

```text
src/cli.js
src/indexer.js
src/service.js
```

---

## 4. Frozen Default Ignore Set (Node Baseline)

Source of truth for Stage 2 baseline generation is the Node discovery list used by `src/service.js` and mirrored in `test/fixtures/generate-discovery-baseline.js`:

- `**/vendor/**`
- `**/node_modules/**`
- `**/.git/**`
- `**/storage/**`
- `**/dist/**`
- `**/build/**`
- `**/tmp/**`
- `**/temp/**`
- `**/.npm/**`
- `**/.yarn/**`
- `**/Library/**`
- `**/System/**`
- `**/.Trash/**`
- `**/.pampa/**`
- `**/pampa.codemap.json`
- `**/pampa.codemap.json.backup-*`
- `**/package-lock.json`
- `**/yarn.lock`
- `**/pnpm-lock.yaml`
- `**/*.json`
- `**/*.sh`
- `**/examples/**`
- `**/assets/**`

Any change to this list during Stage 2 requires updating this contract and the Stage 2 plan docs.
