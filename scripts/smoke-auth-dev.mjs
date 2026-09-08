import { join } from 'node:path';
import { localEnv } from './local-env.mjs';
import { root, run } from './process.mjs';

// localEnv refuses CI and never prints the prepared credentials. The migrator
// connection is used only to remove this run's temporary test identity.
const env = { ...localEnv('api'), AUTH_SMOKE_CLEANUP_URL: localEnv('migrator').DATABASE_URL };
run('go', ['run', './cmd/auth-smoke'], { cwd: join(root, 'server'), env });
