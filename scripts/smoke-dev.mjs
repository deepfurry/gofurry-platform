import { join } from 'node:path';
import { localEnv } from './local-env.mjs';
import { root, run } from './process.mjs';

for (const service of ['api', 'admin', 'migrator', 'worker']) {
  run('go', ['run', './cmd/smoke', '-service', service], { cwd: join(root, 'server'), env: localEnv(service) });
}
