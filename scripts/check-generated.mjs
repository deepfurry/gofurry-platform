import { createHash } from 'node:crypto';
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { root, run } from './process.mjs';

const dirs = ['server/internal/transport/public/generated', 'server/internal/transport/admin/generated',
  'server/internal/database/sqlc', 'packages/api-client/src/generated'];
function snapshot() {
  const files = {};
  function visit(path) {
    if (!existsSync(join(root, path))) throw new Error(`Missing generated directory: ${path}`);
    for (const entry of readdirSync(join(root, path), { withFileTypes: true })) {
      const child = `${path}/${entry.name}`;
      if (entry.isDirectory()) visit(child);
      else files[child] = createHash('sha256').update(readFileSync(join(root, child))).digest('hex');
    }
  }
  dirs.forEach(visit);
  return JSON.stringify(Object.entries(files).sort(([a], [b]) => a.localeCompare(b)));
}
const before = snapshot();
run('pnpm', ['generate']);
if (snapshot() !== before) throw new Error('Generated-code drift: review regenerated output and rerun check');
console.log('Generated outputs are byte-identical; no added or removed files');
