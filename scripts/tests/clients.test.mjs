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
