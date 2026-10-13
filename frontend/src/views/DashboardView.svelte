<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../lib/api/client';
  import type { DashboardData } from '../lib/types';
  import { useRouter } from '../lib/router/router.svelte';
  import CaddyStatusCard from '../lib/components/dashboard/CaddyStatusCard.svelte';
  import AppHealthCard from '../lib/components/dashboard/AppHealthCard.svelte';
  import AuditActivityList from '../lib/components/dashboard/AuditActivityList.svelte';
  import Button from '../lib/components/common/Button.svelte';
  import { RefreshCw, LayoutDashboard, Clock, Globe, ArrowRight, Layers, ShieldCheck, CheckCircle2, Activity } from 'lucide-svelte';

  const router = useRouter();
  let data = $state<DashboardData | undefined>(undefined);
  let loading = $state<boolean>(true);
  let refreshing = $state<boolean>(false);
  let error = $state<string | undefined>(undefined);
  let pollInterval = $state<'off' | '10' | '30'>('off');

  let timerId: any = null;

  async function fetchDashboard(isManual = false) {
    if (isManual) refreshing = true;
    error = undefined;

    try {
      const dashData = await api.getDashboard();
      
      // Fetch system Caddy version from host binary
      try {
        const sysVersion = await api.getSystemCaddyVersion();
        if (sysVersion && sysVersion.version) {
          dashData.caddy.version = sysVersion.version;
        }
      } catch {
        // System Caddy version call failed, retain existing version or empty
      }

      data = dashData;
    } catch (err: any) {
      error = err.message || 'Failed to fetch dashboard data';
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  function setPolling(interval: 'off' | '10' | '30') {
    pollInterval = interval;
    if (timerId) {
      clearInterval(timerId);
      timerId = null;
    }

    if (interval !== 'off') {
      const ms = parseInt(interval, 10) * 1000;
      timerId = setInterval(() => {
        fetchDashboard(false);
      }, ms);
    }
  }

  onMount(() => {
    fetchDashboard(false);
  });

  onDestroy(() => {
    if (timerId) {
      clearInterval(timerId);
    }
  });
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
    <div>
      <div class="flex items-center gap-2">
        <LayoutDashboard class="w-5 h-5 text-slate-700 dark:text-slate-300" />
        <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
          Operational Dashboard
        </h1>
      </div>
      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
        Caddy server connectivity & system health overview
      </p>
    </div>

    <div class="flex items-center gap-3">
      <div class="flex items-center gap-1.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-md p-1 text-xs">
        <Clock class="w-3.5 h-3.5 text-slate-400 ml-1.5" />
        <span class="text-slate-500 dark:text-slate-400 text-[11px] font-medium mr-1">Poll:</span>
        <button
          type="button"
          onclick={() => setPolling('off')}
          class="px-2 py-0.5 rounded-sm transition-colors cursor-pointer {pollInterval === 'off' ? 'bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900 font-semibold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
        >
          Off
        </button>
        <button
          type="button"
          onclick={() => setPolling('10')}
          class="px-2 py-0.5 rounded-sm transition-colors cursor-pointer {pollInterval === '10' ? 'bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900 font-semibold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
        >
          10s
        </button>
        <button
          type="button"
          onclick={() => setPolling('30')}
          class="px-2 py-0.5 rounded-sm transition-colors cursor-pointer {pollInterval === '30' ? 'bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900 font-semibold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
        >
          30s
        </button>
      </div>

      <Button
        variant="outline"
        size="sm"
        onclick={() => fetchDashboard(true)}
        loading={refreshing}
      >
        <RefreshCw class="w-3.5 h-3.5" />
        <span>Refresh</span>
      </Button>
    </div>
  </div>

  {#if error}
    <div class="p-4 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-600 dark:text-rose-400">
      <span class="font-semibold">Error loading dashboard:</span> {error}
    </div>
  {/if}

  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <CaddyStatusCard
      status={data?.caddy}
      loading={refreshing}
      onRetry={() => fetchDashboard(true)}
    />
    <AppHealthCard health={data?.health} />
  </div>

  <!-- Site Summary Operational Widget -->
  {#if data?.site_summary}
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-5 shadow-xs space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-md bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <Globe class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-sm font-semibold text-slate-900 dark:text-slate-100">
              Reverse Proxy Sites Overview
            </h2>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Configured reverse proxy domain routes in Caddy
            </p>
          </div>
        </div>

        <button
          onclick={() => router.navigate('/sites')}
          class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 hover:underline flex items-center gap-1 cursor-pointer"
        >
          <span>Manage Sites</span>
          <ArrowRight class="w-3.5 h-3.5" />
        </button>
      </div>

      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div class="p-3 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md">
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-medium">Total Sites</div>
          <div class="text-xl font-bold font-mono text-slate-900 dark:text-slate-100 mt-0.5">
            {data.site_summary.total_sites}
          </div>
        </div>

        <div class="p-3 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md">
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-medium">Caddio Managed</div>
          <div class="text-xl font-bold font-mono text-emerald-600 dark:text-emerald-400 mt-0.5">
            {data.site_summary.managed_sites}
          </div>
        </div>

        <div class="p-3 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md">
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-medium font-medium">Active / Enabled</div>
          <div class="text-xl font-bold font-mono text-slate-900 dark:text-slate-100 mt-0.5">
            {data.site_summary.enabled_sites}
          </div>
        </div>

        <div class="p-3 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md">
          <div class="text-[11px] text-slate-500 dark:text-slate-400 font-medium">Auto HTTPS</div>
          <div class="text-xl font-bold font-mono text-blue-600 dark:text-blue-400 mt-0.5">
            {data.site_summary.https_sites}
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Runtime Observability Quick Banner -->
  <div class="bg-gradient-to-r from-indigo-900/10 via-purple-900/10 to-indigo-900/10 dark:from-indigo-950/40 dark:via-purple-950/40 dark:to-indigo-950/40 border border-indigo-200 dark:border-indigo-800/60 rounded-lg p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div class="flex items-center space-x-3">
      <div class="w-10 h-10 rounded-lg bg-indigo-600 text-white flex items-center justify-center shrink-0 shadow-xs">
        <Activity class="w-5 h-5" />
      </div>
      <div>
        <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100">Runtime Monitoring & Logs</h3>
        <p class="text-xs text-slate-600 dark:text-slate-400">Inspect live server status, configuration hashes, and real-time Caddy log streams.</p>
      </div>
    </div>
    <button
      onclick={() => router.navigate('/runtime')}
      class="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white font-medium text-xs rounded-md shadow-xs transition-colors shrink-0 flex items-center justify-center space-x-1.5 cursor-pointer"
    >
      <span>Open Observability</span>
      <ArrowRight class="w-3.5 h-3.5" />
    </button>
  </div>

  <AuditActivityList logs={data?.recent_logs} loading={loading} />
</div>


