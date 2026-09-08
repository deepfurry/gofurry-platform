import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

export const root = fileURLToPath(new URL('../', import.meta.url));

// All command names/arguments are repository-owned, never shell-interpolated input.
export function run(command, args = [], options = {}) {
  const isPnpm = command === 'pnpm';
  const executable = isPnpm ? process.execPath : command;
  const commandArgs = isPnpm ? [process.env.npm_execpath, ...args] : args;
  if (isPnpm && !process.env.npm_execpath) throw new Error('Run this command through pnpm');
  const result = spawnSync(executable, commandArgs, {
    cwd: root, stdio: 'inherit', windowsHide: true, ...options,
  });
  if (result.error || result.status !== 0) {
    throw new Error(`${command} failed (exit ${result.status ?? 'unavailable'})`);
  }
  return result;
}
