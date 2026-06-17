'use strict';
/*
 * DATA-HTML KNOCKOUT BINDING LINT
 * ================================
 * Scans web_frontend/src/**\/*.{html,js} for Knockout attr bindings that set
 * `data-html` to a value expression containing string concatenation (+),
 * array join (.join(...)), or a template literal (backtick). These patterns
 * can introduce XSS when untrusted input is interpolated into raw HTML.
 *
 * DETECTION SCOPE
 * ---------------
 * The linter operates line-by-line. A binding whose key and value span
 * multiple lines will only have the key-line portion analysed; continuation
 * lines are invisible. Known limitation — no such multi-line pattern exists
 * in the codebase today. The DEFERRED allowlist entry for list.html:112
 * keeps that case under observation.
 *
 * String concatenation (+) is flagged at ANY depth inside the value
 * expression (inside parens, ternaries, or function arguments), not only at
 * the top level. Paren-wrapping is NOT a valid bypass.
 *
 * ALLOWLIST CONTRACT
 * ------------------
 * Safe usages are recorded in scripts/lint-data-html.allowlist.json as:
 *
 *   {
 *     "web_frontend/src/path/to/file.html:lineNumber": {
 *       "justification": "why this specific binding is safe"
 *     }
 *   }
 *
 * Rules:
 * - Allowlist keys are "<file>:<line>" where <file> is relative to the
 *   repository root (e.g. "web_frontend/src/components/agents/agents.html:82").
 * - Every entry MUST have a non-empty "justification" string. Missing or
 *   empty justification causes a non-zero exit at load time, naming the key.
 * - Entries whose justification starts with "DEFERRED:" produce a one-line
 *   WARN on stdout so deferred items remain visible to reviewers. They do NOT
 *   cause a non-zero exit on their own.
 * - Any flagged binding NOT in the allowlist causes a non-zero exit with an
 *   error message naming the exact file, line, and offending value expression,
 *   ending with the literal instruction below.
 * - To add a new safe entry, add it to scripts/lint-data-html.allowlist.json
 *   with a justification.
 *
 * USAGE
 * -----
 *   node scripts/lint-data-html.js          # from web_frontend/
 *   npm run lint:data-html
 *   gulp lint:data-html                     # wired into the default build
 */

const fs = require('fs');
const path = require('path');

const SCRIPT_DIR = __dirname;                             // web_frontend/scripts/
const FRONTEND_DIR = path.resolve(SCRIPT_DIR, '..');     // web_frontend/
const REPO_ROOT = path.resolve(SCRIPT_DIR, '../..');     // repository root
const SRC_DIR = path.join(FRONTEND_DIR, 'src');
const ALLOWLIST_PATH = path.join(SCRIPT_DIR, 'lint-data-html.allowlist.json');

// Load allowlist and validate every entry has a non-empty justification.
let allowlist = {};
try {
    allowlist = JSON.parse(fs.readFileSync(ALLOWLIST_PATH, 'utf8'));
} catch (e) {
    if (e.code !== 'ENOENT') throw e;
    console.warn('WARN: lint-data-html.allowlist.json not found; no entries are allowlisted.');
}
{
    const badKeys = Object.keys(allowlist).filter(k => {
        const j = allowlist[k].justification;
        return typeof j !== 'string' || j.trim() === '';
    });
    if (badKeys.length > 0) {
        for (const k of badKeys) {
            console.error(`ERROR allowlist entry "${k}" is missing a non-empty "justification" field.\n  If this binding is safe, add it to scripts/lint-data-html.allowlist.json with a justification.`);
        }
        process.exit(1);
    }
}

// Recursively collect files with the given extensions
function collectFiles(dir, exts) {
    const results = [];
    const entries = fs.readdirSync(dir, {withFileTypes: true});
    for (const entry of entries) {
        const full = path.join(dir, entry.name);
        if (entry.isDirectory()) {
            results.push(...collectFiles(full, exts));
        } else if (exts.some(ext => entry.name.endsWith(ext))) {
            results.push(full);
        }
    }
    return results;
}

// Extract the Knockout binding value expression that starts at position `start`
// inside `line`. Scans forward respecting string literals and bracket depth
// until it hits an unbalanced `,` or closing `}`, `)`, or `]`.
// Returns the trimmed expression string.
function extractValueExpr(line, start) {
    let i = start;
    while (i < line.length && /\s/.test(line[i])) i++;
    const valueStart = i;
    let depth = 0;
    let inString = null;
    let escaped = false;

    while (i < line.length) {
        const ch = line[i];
        if (escaped) {
            escaped = false;
            i++;
            continue;
        }
        if (ch === '\\' && inString) {
            escaped = true;
            i++;
            continue;
        }
        if (inString) {
            if (ch === inString) inString = null;
            i++;
            continue;
        }
        if (ch === '"' || ch === "'" || ch === '`') {
            inString = ch;
            i++;
            continue;
        }
        if (ch === '(' || ch === '[' || ch === '{') {
            depth++;
            i++;
            continue;
        }
        if (ch === ')' || ch === ']' || ch === '}') {
            if (depth === 0) break;
            depth--;
            i++;
            continue;
        }
        if (ch === ',' && depth === 0) break;
        i++;
    }
    return line.slice(valueStart, i).trim();
}

// Returns a description of the first dangerous pattern found in `expr`, or
// null if the expression is safe.
function findDangerousPattern(expr) {
    // Template literal
    if (expr.includes('`')) return 'template literal (`...`)';

    // Array .join()
    if (/\.join\s*\(/.test(expr)) return 'array .join()';

    // String concatenation: look for `+` at ANY depth, outside strings.
    // Paren-wrapping (e.g. `(a + b)`) is NOT a bypass — depth tracking
    // is deliberately omitted here so nested expressions are caught.
    let inStr = null;
    let esc = false;
    for (let i = 0; i < expr.length; i++) {
        const ch = expr[i];
        if (esc) {
            esc = false;
            continue;
        }
        if (ch === '\\' && inStr) {
            esc = true;
            continue;
        }
        if (inStr) {
            if (ch === inStr) inStr = null;
            continue;
        }
        if (ch === '"' || ch === "'" || ch === '`') {
            inStr = ch;
            continue;
        }
        if (ch === '+') return 'string concatenation (+)';
    }
    return null;
}

// Regex: locate `'data-html':` or `"data-html":` within a line
const DATA_HTML_KEY_RE = /["']data-html["']\s*:/g;

let violations = 0;
let warns = 0;
const files = collectFiles(SRC_DIR, ['.html', '.js']);

for (const filePath of files) {
    const relPath = path.relative(REPO_ROOT, filePath).replace(/\\/g, '/'); // normalise Windows paths
    const content = fs.readFileSync(filePath, 'utf8');
    const lines = content.split('\n');

    for (let lineIdx = 0; lineIdx < lines.length; lineIdx++) {
        const line = lines[lineIdx];
        const lineNum = lineIdx + 1;

        DATA_HTML_KEY_RE.lastIndex = 0;
        let match;
        while ((match = DATA_HTML_KEY_RE.exec(line)) !== null) {
            const afterColon = match.index + match[0].length;
            const expr = extractValueExpr(line, afterColon);
            if (!expr) continue;

            const reason = findDangerousPattern(expr);
            if (!reason) continue;

            const key = `${relPath}:${lineNum}`;
            const entry = allowlist[key];

            if (entry) {
                const justification = (entry.justification || '').trim();
                if (justification.startsWith('DEFERRED:')) {
                    console.warn(`WARN [deferred] ${key}: ${justification}`);
                    warns++;
                }
                // Allowlisted — not an error
            } else {
                console.error(
                    `ERROR ${key}: unsafe data-html binding — ${reason}\n` +
                    `  Value: ${expr}\n` +
                    `  If this binding is safe, add it to scripts/lint-data-html.allowlist.json with a justification.`
                );
                violations++;
            }
        }
    }
}

if (violations > 0) {
    console.error(`\ndata-html lint FAILED: ${violations} violation(s). See above for details.`);
    process.exit(1);
} else if (warns > 0) {
    console.log(`data-html lint passed with ${warns} DEFERRED warning(s). Review DEFERRED allowlist entries before extending the pattern.`);
} else {
    console.log('data-html lint passed.');
}
