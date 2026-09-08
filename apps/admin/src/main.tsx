import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { createRootRoute, createRoute, createRouter, Outlet, RouterProvider } from '@tanstack/react-router';
import { getReady } from '@gofurry/api-client/admin';
import '@gofurry/design/global.scss';
import './layout.css';
import styles from './Foundation.module.scss';

function Foundation() {
  const health = useQuery({
    queryKey: ['admin', 'health', 'ready'],
    queryFn: async ({ signal }) => (await getReady({ signal })).data,
    retry: false,
  });
  return <main className="mx-auto flex min-h-screen max-w-3xl flex-col justify-center gap-6 p-6">
    <section className={`${styles.panel} flex flex-col gap-6 p-6 sm:p-10`} aria-labelledby="title">
      <span className={styles.phase}>P0-0 · Engineering foundation</span>
      <h1 id="title">GoFurry Admin</h1>
      <p>The admin runtime is ready for the next implementation phase.</p>
      <div className="flex flex-wrap items-center gap-4">
        <button className={`${styles.button} px-4 py-2`} disabled={health.isFetching} onClick={() => { void health.refetch(); }}>Check API readiness</button>
        <span role="status">{health.isPending ? 'Checking…' : health.isError ? 'Admin API unavailable' : `Admin API: ${health.data.status}`}</span>
      </div>
    </section>
  </main>;
}
const rootRoute = createRootRoute({ component: Outlet });
const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: '/', component: Foundation });
const router = createRouter({ routeTree: rootRoute.addChildren([indexRoute]) });
const queryClient = new QueryClient();
declare module '@tanstack/react-router' { interface Register { router: typeof router } }
const element = document.getElementById('root');
if (!element) throw new Error('Missing admin root');
createRoot(element).render(<StrictMode><QueryClientProvider client={queryClient}><RouterProvider router={router} /></QueryClientProvider></StrictMode>);
