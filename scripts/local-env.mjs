import { readFileSync } from 'node:fs';
import { parseEnv } from 'node:util';
import { join } from 'node:path';
import { root } from './process.mjs';

export function localEnv(service) {
  if (process.env.CI) throw new Error('CI must never load private local configuration');
  if (!['api', 'admin', 'worker', 'migrator'].includes(service)) throw new Error('Unknown local service');
  let values;
  try { values = parseEnv(readFileSync(join(root, 'server/env', `${service}.local`), 'utf8')); }
  catch { throw new Error(`Unable to load private ${service} launch input (values withheld)`); }
  // Prepared credentials predate HTTP/River bootstrap fields. Supply only missing
  // non-secret launch defaults; explicit process/file settings still win.
  const defaults = service === 'api' ? { HTTP_ADDR: '127.0.0.1:8080' }
    : service === 'admin' ? { HTTP_ADDR: '127.0.0.1:8081' } : { RIVER_SCHEMA: 'river' };
  return { ...defaults, ...process.env, ...values };
}
