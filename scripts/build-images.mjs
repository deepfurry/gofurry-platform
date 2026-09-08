import { run } from './process.mjs';

run('docker', ['version', '--format', '{{.Server.Version}}']);
for (const [name, dockerfile] of [
  ['server', 'deploy/docker/server.Dockerfile'], ['web', 'deploy/docker/web.Dockerfile'],
  ['admin', 'deploy/docker/admin.Dockerfile'], ['postgres', 'deploy/postgres/Dockerfile'],
]) run('docker', ['build', '--file', dockerfile, '--tag', `gofurry-${name}:p0-0-local`, '.']);
