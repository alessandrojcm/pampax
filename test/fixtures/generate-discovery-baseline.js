#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import fg from "fast-glob";

const SUPPORTED_LANGUAGE_EXTENSIONS = [
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
];

const DEFAULT_IGNORE_PATTERNS = [
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
];

function normalizeRelativePath(relativePath) {
  return relativePath.replaceAll(path.sep, "/").replace(/^\.\//, "");
}

function parseArgs(argv) {
  const args = argv.slice(2);
  if (args.length === 0 || args.includes("-h") || args.includes("--help")) {
    return { help: true };
  }

  const repoPath = args[0];
  let outputPath = null;

  for (let i = 1; i < args.length; i += 1) {
    const token = args[i];
    if (token === "--out") {
      outputPath = args[i + 1] ?? null;
      i += 1;
    }
  }

  return {
    help: false,
    repoPath,
    outputPath,
  };
}

function printUsage() {
  console.log("Usage: node test/fixtures/generate-discovery-baseline.js <repo-path> [--out <file>]");
  console.log("");
  console.log("Examples:");
  console.log("  node test/fixtures/generate-discovery-baseline.js ./test/fixtures/small");
  console.log("  node test/fixtures/generate-discovery-baseline.js ./test/fixtures/small --out instructions/stage2_artifacts/03_fixture_paths_small.txt");
}

async function buildBaselineList(repoPath) {
  const languagePatterns = SUPPORTED_LANGUAGE_EXTENSIONS.map((ext) => `**/*${ext}`);

  const files = await fg(languagePatterns, {
    cwd: repoPath,
    absolute: false,
    followSymbolicLinks: false,
    ignore: DEFAULT_IGNORE_PATTERNS,
    onlyFiles: true,
    dot: false,
  });

  const unique = new Set(files.map(normalizeRelativePath));
  return Array.from(unique).sort((a, b) => a.localeCompare(b, "en"));
}

async function main() {
  const parsed = parseArgs(process.argv);

  if (parsed.help) {
    printUsage();
    process.exit(0);
  }

  if (!parsed.repoPath) {
    printUsage();
    process.exit(1);
  }

  const repoPath = path.resolve(parsed.repoPath);

  try {
    await fs.access(repoPath);
  } catch {
    console.error(`Repository path does not exist: ${repoPath}`);
    process.exit(1);
  }

  const baseline = await buildBaselineList(repoPath);
  const outputText = `${baseline.join("\n")}\n`;

  if (parsed.outputPath) {
    const resolvedOutputPath = path.resolve(parsed.outputPath);
    const outputDir = path.dirname(resolvedOutputPath);
    await fs.mkdir(outputDir, { recursive: true });
    await fs.writeFile(resolvedOutputPath, outputText, "utf8");
    console.log(`Wrote ${baseline.length} paths to ${resolvedOutputPath}`);
    return;
  }

  process.stdout.write(outputText);
}

main().catch((error) => {
  console.error(`Discovery baseline generation failed: ${error.message}`);
  process.exit(1);
});
