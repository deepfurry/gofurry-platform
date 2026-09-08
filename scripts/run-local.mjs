import { join } from 'node:path';
import { localEnv } from './local-env.mjs';
import { root, run } from './process.mjs';

const service = process.argv[2];
const env = localEnv(service);
run('go', ['run', `./cmd/${service === 'migrator' ? 'migrate' : service}`], { cwd: join(root, 'server'), env });
