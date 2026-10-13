<script lang="ts">
  import type { RuntimeStatusResponse } from '../../types';

  export let status: RuntimeStatusResponse | null = null;
  export let loading: boolean = false;

  function getStatusBadgeClass(st: string | undefined): string {
    switch (st) {
      case 'healthy':
        return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20';
      case 'degraded':
        return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20';
      case 'unreachable':
        return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20';
      default:
        return 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20';
    }
  }

  function formatTime(isoStr: string | undefined): string {
    if (!isoStr) return 'N/A';
    try {
      return new Date(isoStr).toLocaleTimeString(undefined, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      });
    } catch {
      return isoStr;
    }
  }
</script>

<div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-xs">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 dark:border-slate-800/80 pb-4 mb-4">
    <div class="flex items-center space-x-3">
      <div class="w-10 h-10 rounded-lg bg-indigo-500/10 dark:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 flex items-center justify-center font-semibold text-lg">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      </div>
      <div>
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">Caddy Runtime Status</h3>
        <p class="text-xs text-slate-500 dark:text-slate-400">Live operational state & API metrics</p>
      </div>
    </div>

    <div class="flex items-center space-x-2">
      {#if loading && !status}
        <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 animate-pulse">
          Checking...
        </span>
      {:else if status}
        <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold uppercase tracking-wider border {getStatusBadgeClass(status.status)}">
          <span class="w-2 h-2 rounded-full mr-2 {status.status === 'healthy' ? 'bg-emerald-500' : status.status === 'degraded' ? 'bg-amber-500' : 'bg-rose-500'}"></span>
          {status.status}
        </span>

        {#if status.log_source}
          <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-mono font-medium bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
            source: {status.log_source}
          </span>
        {/if}

        {#if !status.logs_available}
          <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20" title="Caddy logs are currently unconfigured or unavailable">
            Logs Unavailable
          </span>
        {/if}
      {/if}
    </div>
  </div>

  {#if status}
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
        <span class="text-xs text-slate-500 dark:text-slate-400 font-medium block mb-1">Caddy Version</span>
        <span class="text-sm font-semibold text-slate-800 dark:text-slate-200 truncate block">
          {status.caddy_version || 'Unavailable'}
        </span>
      </div>

      <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
        <span class="text-xs text-slate-500 dark:text-slate-400 font-medium block mb-1">Active Config Hash</span>
        <span class="text-xs font-mono font-semibold text-slate-800 dark:text-slate-200 truncate block" title={status.active_config_hash}>
          {status.active_config_hash ? status.active_config_hash.slice(0, 12) + '...' : 'Unavailable'}
        </span>
      </div>

      <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
        <span class="text-xs text-slate-500 dark:text-slate-400 font-medium block mb-1">Active Servers</span>
        <span class="text-sm font-semibold text-slate-800 dark:text-slate-200 block">
          {status.server_count} {status.server_count === 1 ? 'server' : 'servers'}
        </span>
      </div>

      <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800">
        <span class="text-xs text-slate-500 dark:text-slate-400 font-medium block mb-1">Configured Sites</span>
        <span class="text-sm font-semibold text-slate-800 dark:text-slate-200 block">
          {status.site_count} {status.site_count === 1 ? 'site' : 'sites'}
        </span>
      </div>
    </div>

    <div class="mt-3 text-right">
      <span class="text-[11px] text-slate-400 dark:text-slate-500">
        Last checked: {formatTime(status.last_checked)}
      </span>
    </div>
  {:else if !loading}
    <div class="py-6 text-center text-slate-500 dark:text-slate-400 text-sm">
      Unable to retrieve runtime status.
    </div>
  {/if}
</div>
