import { join } from 'node:path';
import { localEnv } from './local-env.mjs';
import { root, run } from './process.mjs';

run('go', ['run', './cmd/oauth-smoke'], { cwd: join(root, 'server'), env: localEnv('api') });
