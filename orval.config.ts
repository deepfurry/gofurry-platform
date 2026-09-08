import { defineConfig } from 'orval';

export default defineConfig({
  public: {
    input: './contracts/openapi/public.yaml',
    output: { target: './packages/api-client/src/generated/public/client.ts',
      client: 'fetch', mode: 'single', clean: true, baseUrl: '/api', formatter: 'prettier' },
  },
  admin: {
    input: './contracts/openapi/admin.yaml',
    output: { target: './packages/api-client/src/generated/admin/client.ts',
      client: 'fetch', mode: 'single', clean: true, baseUrl: '/api', formatter: 'prettier' },
  },
});
