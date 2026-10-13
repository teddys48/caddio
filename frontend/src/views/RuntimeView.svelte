<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api/client';
  import type { RuntimeStatusResponse, RuntimeLogEntry, RuntimeLogsResponse, RuntimeLogFilters } from '../lib/types';
  import RuntimeStatusCard from '../lib/components/runtime/RuntimeStatusCard.svelte';
  import RuntimeLogFiltersComponent from '../lib/components/runtime/RuntimeLogFilters.svelte';
  import RuntimeLogsTable from '../lib/components/runtime/RuntimeLogsTable.svelte';
  import RuntimeLogDetail from '../lib/components/runtime/RuntimeLogDetail.svelte';
  import RuntimeRefreshControl from '../lib/components/runtime/RuntimeRefreshControl.svelte';

  let status: RuntimeStatusResponse | null = null;
  let statusLoading: boolean = false;

  let logs: RuntimeLogEntry[] = [];
  let logsLoading: boolean = false;
  let logsAvailable: boolean = true;
  let statusReason: string | undefined = undefined;

  let filters: RuntimeLogFilters = {
    level: 'all',
    search: '',
    limit: 100
  };

  let selectedLog: RuntimeLogEntry | null = null;
  let errorMessage: string | null = null;
  let successMessage: string | null = null;

  async function fetchStatus() {
    statusLoading = true;
    try {
      status = await api.getRuntimeStatus();
    } catch (err: any) {
      errorMessage = err.message || 'Failed to fetch runtime status';
    } finally {
      statusLoading = false;
    }
  }

  async function fetchLogs() {
    logsLoading = true;
    try {
      const res: RuntimeLogsResponse = await api.getRuntimeLogs(filters);
      logs = res.entries || [];
      logsAvailable = res.available;
      statusReason = res.reason;
    } catch (err: any) {
      errorMessage = err.message || 'Failed to fetch runtime logs';
      logsAvailable = false;
    } finally {
      logsLoading = false;
    }
  }

  async function refreshAll() {
    errorMessage = null;
    await Promise.all([fetchStatus(), fetchLogs()]);
  }

  function handleFilterChange(e: CustomEvent<RuntimeLogFilters>) {
    filters = e.detail;
    fetchLogs();
  }

  function handleLogSelect(e: CustomEvent<RuntimeLogEntry>) {
    selectedLog = e.detail;
  }

  async function handleEnableLogging() {
    if (!status?.active_config_hash) {
      errorMessage = 'Active configuration hash unavailable. Please refresh.';
      return;
    }
    errorMessage = null;
    successMessage = null;

    try {
      const res = await api.enableRuntimeLogging(status.active_config_hash);
      if (res.changed) {
        successMessage = res.message || 'Caddy logging enabled successfully!';
        await refreshAll();
      } else {
        errorMessage = res.message || 'Caddy logging configuration already exists.';
      }
    } catch (err: any) {
      errorMessage = err.message || 'Failed to enable Caddy logging.';
    }
  }

  onMount(() => {
    refreshAll();
  });
</script>

<div class="space-y-6 max-w-7xl mx-auto px-4 sm:px-6 py-6">
  <!-- Page Header -->
  <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">Runtime & Logs</h1>
      <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
        Monitor live Caddy operational state and inspect real-time application logs
      </p>
    </div>

    <div>
      <RuntimeRefreshControl loading={statusLoading || logsLoading} on:refresh={refreshAll} />
    </div>
  </div>

  <!-- Global Error Alert -->
  {#if errorMessage}
    <div class="p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-700 dark:text-rose-400 text-sm flex items-start space-x-3">
      <svg class="w-5 h-5 shrink-0 text-rose-500 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <div class="flex-1">
        <span class="font-semibold">Error: </span>{errorMessage}
      </div>
      <button
        type="button"
        on:click={() => (errorMessage = null)}
        class="text-rose-500 hover:text-rose-700 dark:hover:text-rose-300 font-bold"
      >
        ✕
      </button>
    </div>
  {/if}

  <!-- Success Notification -->
  {#if successMessage}
    <div class="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-700 dark:text-emerald-400 text-sm flex items-start space-x-3">
      <svg class="w-5 h-5 shrink-0 text-emerald-500 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
      </svg>
      <div class="flex-1">
        <span class="font-semibold">Success: </span>{successMessage}
      </div>
      <button
        type="button"
        on:click={() => (successMessage = null)}
        class="text-emerald-500 hover:text-emerald-700 dark:hover:text-emerald-300 font-bold"
      >
        ✕
      </button>
    </div>
  {/if}

  <!-- Status Card -->
  <RuntimeStatusCard {status} loading={statusLoading} />

  <!-- Log Filters & Table -->
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">Runtime Logs</h2>
      {#if logs.length > 0}
        <span class="text-xs text-slate-500 dark:text-slate-400 font-mono">
          Showing {logs.length} entries
        </span>
      {/if}
    </div>

    <RuntimeLogFiltersComponent {filters} on:change={handleFilterChange} />

    <RuntimeLogsTable
      {logs}
      loading={logsLoading}
      {logsAvailable}
      {statusReason}
      on:select={handleLogSelect}
      on:enableLogging={handleEnableLogging}
    />
  </div>

  <!-- Log Detail Modal / Slide-over -->
  <RuntimeLogDetail log={selectedLog} on:close={() => (selectedLog = null)} />
</div>
