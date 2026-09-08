import assert from 'node:assert/strict';
import { test } from 'node:test';
import * as publicClient from '../../packages/api-client/src/generated/public/client.ts';
import * as adminClient from '../../packages/api-client/src/generated/admin/client.ts';

for (const [name, client] of Object.entries({ public: publicClient, admin: adminClient })) {
  test(`${name}: generated fetch preserves degraded/503 responses and AbortSignal`, async t => {
    const controller = new AbortController();
    const calls = [];
    t.mock.method(globalThis, 'fetch', async (url, options) => {
      calls.push([url, options]);
      return new Response(JSON.stringify({ status: 'unavailable', postgres: 'down', redis: 'up' }), {
        status: 503, headers: { 'Content-Type': 'application/json' },
      });
    });
    const response = await client.getReady({ signal: controller.signal });
    assert.equal(response.status, 503);
    assert.equal(response.data.status, 'unavailable');
    assert.equal(calls[0][0], '/api/health/ready');
    assert.equal(calls[0][1].signal, controller.signal);
    assert.equal(calls[0][1].method, 'GET');
  });
  test(`${name}: generated liveness URL uses the same-origin API prefix`, () => {
    assert.equal(client.getGetLiveUrl(), '/api/health/live');
  });
}

test('public: auth clients preserve same-origin options, nullable PATCH fields and empty logout', async t => {
  const calls = [];
  t.mock.method(globalThis, 'fetch', async (url, options) => {
    calls.push([url, options]);
    if (url === '/api/auth/logout') return new Response(null, { status: 204 });
    return new Response(JSON.stringify({ code: 'AUTH_UNAUTHENTICATED', message: 'Please log in.' }), { status: 401 });
  });
  const controller = new AbortController();
  const options = { credentials: 'same-origin', signal: controller.signal };
  const login = await publicClient.login({ email: 'test@example.invalid', password: 'test password only' }, options);
  assert.equal(login.status, 401);
  assert.equal(login.data.code, 'AUTH_UNAUTHENTICATED');
  await publicClient.updateProfile({ bio: null }, options);
  const logout = await publicClient.logout(options);
  assert.equal(logout.status, 204);
  assert.equal(logout.data, undefined);
  assert.equal(calls[0][0], '/api/auth/login');
  assert.equal(calls[1][0], '/api/me/profile');
  assert.deepEqual(JSON.parse(calls[1][1].body), { bio: null });
  for (const [, request] of calls) {
    assert.equal(request.credentials, 'same-origin');
    assert.equal(request.signal, controller.signal);
  }
});
