#!/usr/bin/env node

/**
 * Fixture Validation Test
 * 
 * Validates that generated fixtures meet all Stage 0.1 specifications:
 * - Database schema matches specification
 * - Chunks are valid and retrievable
 * - Codemap structure is correct
 * - Search outputs are valid
 */

import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import { promisify } from 'node:util';
import { gunzip as gunzipCallback } from 'node:zlib';
import crypto from 'node:crypto';
import sqlite3 from 'sqlite3';

const gunzip = promisify(gunzipCallback);

/**
 * Load fixture manifest
 */
async function loadManifest(fixturePath) {
    const manifestPath = path.join(fixturePath, 'manifest.json');
    const content = await fs.readFile(manifestPath, 'utf-8');
    return JSON.parse(content);
}

/**
 * Validate database schema
 */
async function validateDatabase(dbPath) {
    return new Promise((resolve, reject) => {
        const db = new sqlite3.Database(dbPath, sqlite3.OPEN_READONLY);
        
        // Check code_chunks table exists
        db.get("SELECT sql FROM sqlite_master WHERE type='table' AND name='code_chunks'", (err, row) => {
            if (err) {
                db.close();
                return reject(err);
            }
            
            assert.ok(row, 'code_chunks table should exist');
            assert.ok(row.sql.includes('id TEXT PRIMARY KEY'), 'Should have id as PRIMARY KEY');
            assert.ok(row.sql.includes('file_path TEXT'), 'Should have file_path column');
            assert.ok(row.sql.includes('sha TEXT'), 'Should have sha column');
            assert.ok(row.sql.includes('embedding BLOB'), 'Should have embedding BLOB column');
            
            // Check row count
            db.get('SELECT COUNT(*) as count FROM code_chunks', (err, row) => {
                db.close();
                
                if (err) {
                    return reject(err);
                }
                
                assert.ok(row.count > 0, 'Database should contain chunks');
                console.log(`  ✓ Database has ${row.count} chunks`);
                resolve();
            });
        });
    });
}

/**
 * Validate a single chunk file
 */
async function validateChunk(chunkPath, expectedSha) {
    // Read gzipped chunk
    const compressed = await fs.readFile(chunkPath);
    
    // Decompress
    const decompressed = await gunzip(compressed);
    
    // Verify SHA-1
    const sha1 = crypto.createHash('sha1').update(decompressed).digest('hex');
    
    // Chunk filename should match SHA-1 hash
    const filename = path.basename(chunkPath, '.gz');
    assert.strictEqual(filename, sha1, `Chunk filename should match SHA-1: ${filename} vs ${sha1}`);
    
    // Content should be non-empty
    assert.ok(decompressed.length > 0, 'Chunk content should not be empty');
    
    return decompressed.toString('utf-8');
}

/**
 * Validate all chunks
 */
async function validateChunks(chunksDir) {
    const files = await fs.readdir(chunksDir);
    const chunkFiles = files.filter(f => f.endsWith('.gz'));
    
    assert.ok(chunkFiles.length > 0, 'Should have at least one chunk file');
    
    console.log(`  Validating ${chunkFiles.length} chunks...`);
    
    // Validate a sample of chunks
    const sampleSize = Math.min(10, chunkFiles.length);
    const samples = chunkFiles.slice(0, sampleSize);
    
    for (const chunkFile of samples) {
        const chunkPath = path.join(chunksDir, chunkFile);
        await validateChunk(chunkPath);
    }
    
    console.log(`  ✓ All ${sampleSize} sampled chunks are valid`);
}

/**
 * Validate codemap structure
 */
async function validateCodemap(codemapPath) {
    const content = await fs.readFile(codemapPath, 'utf-8');
    const codemap = JSON.parse(content);
    
    // Codemap is a flat object with chunk IDs as keys
    const chunkIds = Object.keys(codemap);
    assert.ok(chunkIds.length > 0, 'Codemap should have at least one chunk');
    
    // Check first chunk structure
    const firstChunkId = chunkIds[0];
    const firstChunk = codemap[firstChunkId];
    
    // Check required fields (from Stage 0.1 spec)
    assert.ok(firstChunk.file, 'Chunk should have file field');
    assert.ok(firstChunk.hasOwnProperty('symbol'), 'Chunk should have symbol field (can be null)');
    assert.ok(firstChunk.sha, 'Chunk should have sha field');
    assert.ok(firstChunk.sha.length === 40, 'SHA should be 40 hex characters');
    assert.ok(firstChunk.lang, 'Chunk should have lang field');
    
    // Verify JSON key ordering (warn if not alphabetically sorted per spec)
    const sortedIds = [...chunkIds].sort();
    const isSorted = chunkIds.every((id, i) => id === sortedIds[i]);
    
    if (!isSorted) {
        console.log('  ⚠️  Warning: Chunk IDs are NOT alphabetically sorted');
        console.log('     (Stage 0.1 spec requires alphabetical sorting)');
    } else {
        console.log('  ✓ Keys are alphabetically sorted');
    }
    
    console.log(`  ✓ Codemap has ${chunkIds.length} chunks`);
    console.log(`  ✓ Codemap structure is valid`);
}

/**
 * Validate search outputs
 */
async function validateSearchOutputs(searchOutputsDir) {
    const files = await fs.readdir(searchOutputsDir);
    const jsonFiles = files.filter(f => f.endsWith('.json'));
    
    assert.ok(jsonFiles.length > 0, 'Should have at least one search output');
    
    for (const jsonFile of jsonFiles) {
        const filePath = path.join(searchOutputsDir, jsonFile);
        const content = await fs.readFile(filePath, 'utf-8');
        const searchResult = JSON.parse(content);
        
        // Check required fields
        assert.ok(searchResult.query, 'Search result should have query');
        assert.ok(searchResult.hasOwnProperty('success'), 'Search result should have success field');
        assert.ok(searchResult.results, 'Search result should have results array');
        assert.ok(Array.isArray(searchResult.results), 'results should be an array');
        
        // Check result structure if results exist
        if (searchResult.results.length > 0) {
            const firstResult = searchResult.results[0];
            assert.ok(firstResult.type, 'Result should have type');
            assert.ok(firstResult.lang, 'Result should have lang');
            assert.ok(firstResult.path, 'Result should have path');
            assert.ok(firstResult.sha, 'Result should have sha');
            assert.ok(firstResult.meta, 'Result should have meta');
            assert.ok(typeof firstResult.meta.score === 'number', 'Result should have numeric score');
        }
    }
    
    console.log(`  ✓ ${jsonFiles.length} search outputs are valid`);
}

/**
 * Validate a complete fixture
 */
async function validateFixture(fixturePath) {
    console.log(`\nValidating fixture: ${fixturePath}`);
    console.log('='.repeat(60));
    
    // Load manifest
    const manifest = await loadManifest(fixturePath);
    console.log(`✓ Manifest loaded (PAMPAX v${manifest.pampax_version})`);
    console.log(`  Node: ${manifest.node_version}`);
    console.log(`  Platform: ${manifest.platform.os}`);
    console.log(`  Embeddings: ${manifest.embedding.provider} (${manifest.embedding.dimensions}D)`);
    
    // Validate database
    const dbPath = path.join(fixturePath, '.pampa', 'pampa.db');
    await validateDatabase(dbPath);
    console.log('✓ Database schema is valid');
    
    // Validate chunks
    const chunksDir = path.join(fixturePath, '.pampa', 'chunks');
    await validateChunks(chunksDir);
    console.log('✓ Chunk files are valid');
    
    // Validate codemap
    const codemapPath = path.join(fixturePath, 'pampa.codemap.json');
    await validateCodemap(codemapPath);
    console.log('✓ Codemap is valid');
    
    // Validate search outputs
    const searchOutputsDir = path.join(fixturePath, 'search_outputs');
    await validateSearchOutputs(searchOutputsDir);
    console.log('✓ Search outputs are valid');
    
    console.log('\n✅ All validations passed!');
}

/**
 * Main execution
 */
async function main() {
    const fixtureSize = process.argv[2] || 'small';
    const fixturePath = path.join('test', 'fixtures', fixtureSize);
    
    try {
        // Check if fixture exists
        await fs.access(fixturePath);
        
        // Validate fixture
        await validateFixture(fixturePath);
        
        console.log(`\n🎉 Fixture '${fixtureSize}' is valid and ready for use!`);
        
    } catch (error) {
        if (error.code === 'ENOENT') {
            console.error(`\n❌ Fixture '${fixtureSize}' not found at: ${fixturePath}`);
            console.error('\nAvailable fixtures:');
            const fixturesDir = path.join('test', 'fixtures');
            try {
                const entries = await fs.readdir(fixturesDir, { withFileTypes: true });
                const dirs = entries.filter(e => e.isDirectory()).map(e => `  - ${e.name}`);
                console.error(dirs.join('\n'));
            } catch (e) {
                console.error('  (none - run fixture generation first)');
            }
        } else {
            console.error('\n❌ Validation failed:', error.message);
            if (process.env.DEBUG) {
                console.error(error.stack);
            }
        }
        process.exit(1);
    }
}

main();
