<script lang="ts">
  import { onMount } from 'svelte';
  import { useAuth } from './lib/stores/auth.svelte';
  import { useRouter } from './lib/router/router.svelte';
  import AppShell from './lib/components/layout/AppShell.svelte';

  import SetupView from './views/SetupView.svelte';
  import LoginView from './views/LoginView.svelte';
  import DashboardView from './views/DashboardView.svelte';
  import AuditView from './views/AuditView.svelte';
  import SitesView from './views/SitesView.svelte';
  import ConfigView from './views/ConfigView.svelte';
  import ConfigHistoryView from './views/ConfigHistoryView.svelte';
  import RuntimeView from './views/RuntimeView.svelte';
  import PlaceholderView from './views/PlaceholderView.svelte';
  import { Server } from 'lucide-svelte';

  const auth = useAuth();
  const router = useRouter();

  onMount(async () => {
    await auth.init();

    if (auth.setupRequired) {
      router.navigate('/setup');
    } else if (!auth.authenticated && router.route !== '/login') {
      router.navigate('/login');
    } else if (auth.authenticated && (router.route === '/login' || router.route === '/setup')) {
      router.navigate('/dashboard');
    }
  });

  $effect(() => {
    if (!auth.loading) {
      if (auth.setupRequired && router.route !== '/setup') {
        router.navigate('/setup');
      } else if (!auth.authenticated && !auth.setupRequired && router.route !== '/login') {
        router.navigate('/login');
      }
    }
  });
</script>

{#if auth.loading}
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 flex flex-col items-center justify-center p-4">
    <div class="w-12 h-12 rounded-xl bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 flex items-center justify-center animate-bounce mb-3 shadow-sm">
      <Server class="w-6 h-6" />
    </div>
    <div class="text-xs font-mono text-slate-500 dark:text-slate-400">Loading Caddio...</div>
  </div>
{:else}
  <AppShell>
    {#if auth.setupRequired}
      <SetupView />
    {:else if !auth.authenticated}
      <LoginView />
    {:else}
      {#if router.route === '/dashboard'}
        <DashboardView />
      {:else if router.route === '/audit'}
        <AuditView />
      {:else if router.route === '/sites'}
        <SitesView />
      {:else if router.route === '/config'}
        <ConfigView />
      {:else if router.route === '/config/history' || router.route === '/history'}
        <ConfigHistoryView />
      {:else if router.route === '/runtime' || router.route === '/logs'}
        <RuntimeView />
      {:else if router.route === '/settings'}
        <PlaceholderView
          title="Application & Caddy Settings"
          description="Global preferences, Admin API URL configuration, and maintenance"
          plannedPhase="Phase 2"
        />
      {:else}
        <DashboardView />
      {/if}
    {/if}
  </AppShell>
{/if}
