#!/usr/bin/env node

/**
 * Fixture Generator for PAMPAX Stage 0.2
 * 
 * Generates golden fixtures for Go port compatibility testing:
 * - Indexes a test repository using the local embedding provider
 * - Captures all artifacts (.pampa/, pampa.codemap.json)
 * - Generates search output snapshots
 * - Creates manifest with metadata for reproducibility
 */

import fs from 'node:fs/promises';
import path from 'node:path';
import { promisify } from 'node:util';
import { exec as execCallback } from 'node:child_process';
import os from 'node:os';
import sqlite3 from 'sqlite3';
import { searchCode } from '../../src/service.js';

const exec = promisify(execCallback);

// Benchmark queries from test/benchmarks/fixtures/queries.js
const BENCHMARK_QUERIES = [
    'create stripe checkout session',
    'synchronize cart totals',
    'send password reset email',
    'render invoice pdf'
];

/**
 * Gather system metadata for reproducibility
 */
async function gatherSystemMetadata() {
    const nodeVersion = process.version;
    const packageJson = JSON.parse(await fs.readFile('package.json', 'utf-8'));
    
    const cpuInfo = os.cpus()[0];
    const cpuModel = cpuInfo ? cpuInfo.model : 'Unknown';
    const cpuCores = os.cpus().length;
    
    const totalMemGB = (os.totalmem() / (1024 ** 3)).toFixed(2);
    
    let gitCommit = 'unknown';
    try {
        const { stdout } = await exec('git rev-parse HEAD');
        gitCommit = stdout.trim();
    } catch (error) {
        console.warn('Could not determine git commit:', error.message);
    }
    
    return {
        timestamp: new Date().toISOString(),
        pampax_version: packageJson.version,
        node_version: nodeVersion,
        platform: {
            os: `${os.type()} ${os.release()}`,
            arch: os.arch(),
            cpu_model: cpuModel,
            cpu_cores: cpuCores,
            total_memory_gb: totalMemGB
        },
        embedding: {
            provider: 'transformers',
            model: 'Xenova/all-MiniLM-L6-v2',
            dimensions: 384
        },
        git_commit: gitCommit
    };
}

/**
 * Index a repository with PAMPAX
 */
async function indexRepository(repoPath, outputDir) {
    console.log(`Indexing repository: ${repoPath}`);
    
    const indexCommand = `node src/cli.js index "${repoPath}" --provider transformers`;
    console.log(`Running: ${indexCommand}`);
    
    const { stdout, stderr } = await exec(indexCommand);
    
    if (stderr) {
        console.warn('Index stderr:', stderr);
    }
    
    console.log('Index stdout:', stdout);
    
    // Verify artifacts exist
    const pampaDir = path.join(repoPath, '.pampa');
    const codemapFile = path.join(repoPath, 'pampa.codemap.json');
    
    const [pampaExists, codemapExists] = await Promise.all([
        fs.access(pampaDir).then(() => true).catch(() => false),
        fs.access(codemapFile).then(() => true).catch(() => false)
    ]);
    
    if (!pampaExists || !codemapExists) {
        throw new Error('Indexing failed: artifacts not found');
    }
    
    console.log('✓ Indexing complete');
}

/**
 * Get database statistics
 */
async function getDatabaseStats(dbPath) {
    return new Promise((resolve, reject) => {
        const db = new sqlite3.Database(dbPath, sqlite3.OPEN_READONLY);
        
        db.get('SELECT COUNT(*) as count FROM code_chunks', (err, row) => {
            if (err) {
                db.close();
                return reject(err);
            }
            
            const chunkCount = row.count;
            
            db.all('SELECT lang, COUNT(*) as count FROM code_chunks GROUP BY lang', (err, rows) => {
                db.close();
                
                if (err) {
                    return reject(err);
                }
                
                const langCounts = {};
                rows.forEach(row => {
                    langCounts[row.lang] = row.count;
                });
                
                resolve({
                    total_chunks: chunkCount,
                    chunks_by_language: langCounts
                });
            });
        });
    });
}

/**
 * Generate search output snapshots
 */
async function generateSearchSnapshots(repoPath, outputDir) {
    console.log('Generating search output snapshots...');
    
    const searchOutputDir = path.join(outputDir, 'search_outputs');
    await fs.mkdir(searchOutputDir, { recursive: true });
    
    for (let i = 0; i < BENCHMARK_QUERIES.length; i++) {
        const query = BENCHMARK_QUERIES[i];
        console.log(`  Query ${i + 1}/${BENCHMARK_QUERIES.length}: "${query}"`);
        
        try {
            const results = await searchCode(query, 10, 'auto', repoPath, {
                hybrid: true,
                bm25: true,
                reranker: 'off'
            });
            
            const snapshotFile = path.join(searchOutputDir, `query_${String(i + 1).padStart(3, '0')}.json`);
            await fs.writeFile(snapshotFile, JSON.stringify({
                query,
                ...results
            }, null, 2));
            
            console.log(`    ✓ ${results.results.length} results saved`);
        } catch (error) {
            console.error(`    ✗ Search failed: ${error.message}`);
        }
    }
    
    console.log('✓ Search snapshots complete');
}

/**
 * Copy artifacts to fixture directory
 */
async function copyArtifacts(repoPath, outputDir) {
    console.log('Copying artifacts to fixture directory...');
    
    const pampaDir = path.join(repoPath, '.pampa');
    const codemapFile = path.join(repoPath, 'pampa.codemap.json');
    
    const targetPampaDir = path.join(outputDir, '.pampa');
    const targetCodemapFile = path.join(outputDir, 'pampa.codemap.json');
    
    // Copy .pampa directory
    await fs.cp(pampaDir, targetPampaDir, { recursive: true });
    
    // Copy codemap
    await fs.copyFile(codemapFile, targetCodemapFile);
    
    console.log('✓ Artifacts copied');
}

/**
 * Create manifest file
 */
async function createManifest(repoPath, outputDir) {
    console.log('Creating manifest...');
    
    const metadata = await gatherSystemMetadata();
    
    const dbPath = path.join(outputDir, '.pampa', 'pampa.db');
    const dbStats = await getDatabaseStats(dbPath);
    
    const manifest = {
        ...metadata,
        repository: {
            path: repoPath,
            ...dbStats
        },
        index_command: `pampax index "${repoPath}" --provider transformers`,
        artifacts: {
            database: '.pampa/pampa.db',
            chunks_directory: '.pampa/chunks/',
            codemap: 'pampa.codemap.json',
            search_outputs: 'search_outputs/'
        }
    };
    
    const manifestFile = path.join(outputDir, 'manifest.json');
    await fs.writeFile(manifestFile, JSON.stringify(manifest, null, 2));
    
    console.log('✓ Manifest created');
    console.log(`\n📊 Fixture Statistics:`);
    console.log(`   Total chunks: ${dbStats.total_chunks}`);
    console.log(`   Languages: ${Object.keys(dbStats.chunks_by_language).join(', ')}`);
}

/**
 * Generate fixture for a repository
 */
async function generateFixture(repoPath, size) {
    console.log(`\n${'='.repeat(70)}`);
    console.log(`Generating ${size.toUpperCase()} fixture`);
    console.log('='.repeat(70));
    
    const outputDir = path.join('test', 'fixtures', size);
    
    // Create output directory
    await fs.mkdir(outputDir, { recursive: true });
    
    try {
        // Step 1: Index repository
        await indexRepository(repoPath, outputDir);
        
        // Step 2: Copy artifacts
        await copyArtifacts(repoPath, outputDir);
        
        // Step 3: Generate search snapshots
        await generateSearchSnapshots(repoPath, outputDir);
        
        // Step 4: Create manifest
        await createManifest(repoPath, outputDir);
        
        console.log(`\n✅ ${size.toUpperCase()} fixture generation complete!`);
        console.log(`   Output: ${outputDir}/`);
        
    } catch (error) {
        console.error(`\n❌ Fixture generation failed:`, error);
        throw error;
    }
}

/**
 * Main execution
 */
async function main() {
    const args = process.argv.slice(2);
    
    if (args.length !== 2) {
        console.error('Usage: node generate-fixtures.js <repo-path> <size>');
        console.error('');
        console.error('Arguments:');
        console.error('  repo-path  Path to repository to index');
        console.error('  size       Fixture size: small, medium, or large');
        console.error('');
        console.error('Example:');
        console.error('  node generate-fixtures.js /path/to/repo small');
        process.exit(1);
    }
    
    const [repoPath, size] = args;
    
    if (!['small', 'medium', 'large'].includes(size)) {
        console.error(`Invalid size: ${size}. Must be small, medium, or large.`);
        process.exit(1);
    }
    
    const resolvedPath = path.resolve(repoPath);
    
    try {
        await fs.access(resolvedPath);
    } catch (error) {
        console.error(`Repository path does not exist: ${resolvedPath}`);
        process.exit(1);
    }
    
    await generateFixture(resolvedPath, size);
}

main().catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
});
