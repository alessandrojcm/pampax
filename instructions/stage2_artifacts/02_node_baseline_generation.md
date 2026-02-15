# Stage 2 Node Baseline Generation

**Version:** 1.0  
**Date:** 2026-02-15  
**Purpose:** deterministic Node discovery baseline for Stage 2 parity tests

---

## 1. Baseline Harness

Use the Stage 2 baseline script:

- Script: `test/fixtures/generate-discovery-baseline.js`
- NPM alias: `npm run baseline:discovery -- <repo-path> [--out <file>]`

This harness emits canonical file lists matching current Node discovery behavior:

- language extension filtering mirrored from `LANG_RULES` in `src/service.js`
- frozen default ignore set
- symlinks not followed
- relative normalized paths with `/`
- stable sorted output

---

## 2. Usage

Print baseline list to stdout:

```bash
npm run baseline:discovery -- .
```

Write baseline list to artifact file:

```bash
npm run baseline:discovery -- . --out instructions/stage2_artifacts/03_fixture_paths_small.txt
```

Generate medium/large files:

```bash
npm run baseline:discovery -- ./test-repos/fastify --out instructions/stage2_artifacts/04_fixture_paths_medium.txt
npm run baseline:discovery -- ./test-repos/next.js --out instructions/stage2_artifacts/05_fixture_paths_large.txt
```

Important: run baselines against source repositories. Do not target `test/fixtures/*` artifact directories.

---

## 3. Output Contract

Generated files are UTF-8 text with:

1. one path per line
2. repo-relative paths only
3. forward slash separators
4. ascending lexical order
5. trailing newline

These files are the Node-side truth set for parity tests until updated by an explicit contract revision.

---

## 4. Determinism Verification

Recommended quick check:

```bash
npm run baseline:discovery -- . --out /tmp/stage2-small-run1.txt
npm run baseline:discovery -- . --out /tmp/stage2-small-run2.txt
cmp /tmp/stage2-small-run1.txt /tmp/stage2-small-run2.txt
```

`cmp` with no output means byte-identical deterministic output.
