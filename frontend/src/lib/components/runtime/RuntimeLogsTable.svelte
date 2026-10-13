<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { RuntimeLogEntry } from '../../types';

  export let logs: RuntimeLogEntry[] = [];
  export let loading: boolean = false;
  export let logsAvailable: boolean = true;
  export let statusReason: string | undefined = undefined;

  const dispatch = createEventDispatcher<{ select: RuntimeLogEntry; enableLogging: void }>();

  let showConfirmModal = false;

  function selectLog(log: RuntimeLogEntry) {
    dispatch('select', log);
  }

  function handleEnableLoggingClick() {
    showConfirmModal = true;
  }

  function confirmEnableLogging() {
    showConfirmModal = false;
    dispatch('enableLogging');
  }

  function getLevelBadgeClass(level: string): string {
    switch (level.toLowerCase()) {
      case 'error':
        return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20';
      case 'warn':
      case 'warning':
        return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20';
      case 'info':
        return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20';
      default:
        return 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20';
    }
  }

  function formatTimestamp(ts: string): string {
    if (!ts) return 'N/A';
    try {
      const d = new Date(ts);
      if (isNaN(d.getTime())) return ts;
      return d.toLocaleTimeString(undefined, {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3
      });
    } catch {
      return ts;
    }
  }

  function getSummaryText(log: RuntimeLogEntry): string {
    if (log.message) return log.message;
    if (log.fields) {
      const req = log.fields.request as Record<string, unknown> | undefined;
      if (req && req.host && req.uri) {
        return `${req.method || 'GET'} ${req.host}${req.uri} -> status ${log.fields.status || '?'}`;
      }
    }
    return JSON.stringify(log);
  }
</script>

<div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-xs">
  {#if !logsAvailable}
    <div class="p-8 text-center space-y-4">
      <div class="w-12 h-12 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 mx-auto flex items-center justify-center">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <div>
        <h4 class="text-base font-semibold text-slate-900 dark:text-slate-100">Logs Unavailable</h4>
        <p class="text-xs text-slate-500 dark:text-slate-400 max-w-md mx-auto mt-1">
          Caddio could not find a supported Caddy log source or active logging application configuration.
        </p>
      </div>

      <div class="pt-2">
        <button
          type="button"
          on:click={handleEnableLoggingClick}
          class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium text-xs rounded-lg shadow-xs transition-colors inline-flex items-center space-x-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          <span>Configure Caddy Logging</span>
        </button>
      </div>
    </div>
  {:else if loading && logs.length === 0}
    <div class="p-8 text-center space-y-3">
      <div class="w-8 h-8 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto"></div>
      <p class="text-xs text-slate-500 dark:text-slate-400 font-medium">Fetching runtime logs...</p>
    </div>
  {:else if logs.length === 0}
    <div class="p-8 text-center space-y-2">
      <svg class="w-10 h-10 text-slate-300 dark:text-slate-600 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300">No Logs Found</h4>
      <p class="text-xs text-slate-500 dark:text-slate-400">No log entries matched your level or search query.</p>
    </div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50 text-[11px] font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            <th class="py-3 px-4 w-20">Level</th>
            <th class="py-3 px-4 w-28">Time</th>
            <th class="py-3 px-4 w-32 hidden md:table-cell">Logger</th>
            <th class="py-3 px-4">Message</th>
            <th class="py-3 px-4 w-16 text-right">Details</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 text-xs">
          {#each logs as log}
            <tr
              on:click={() => selectLog(log)}
              class="hover:bg-slate-50/80 dark:hover:bg-slate-800/40 cursor-pointer transition-colors"
            >
              <td class="py-2.5 px-4 font-mono font-medium">
                <span class="inline-block px-2 py-0.5 rounded text-[10px] uppercase font-semibold border {getLevelBadgeClass(log.level)}">
                  {log.level}
                </span>
              </td>
              <td class="py-2.5 px-4 font-mono text-slate-500 dark:text-slate-400 text-[11px] whitespace-nowrap">
                {formatTimestamp(log.timestamp)}
              </td>
              <td class="py-2.5 px-4 font-mono text-slate-600 dark:text-slate-300 truncate max-w-[120px] hidden md:table-cell">
                {log.logger || 'default'}
              </td>
              <td class="py-2.5 px-4 font-mono text-slate-800 dark:text-slate-200 truncate max-w-md">
                {getSummaryText(log)}
              </td>
              <td class="py-2.5 px-4 text-right">
                <button
                  type="button"
                  on:click|stopPropagation={() => selectLog(log)}
                  aria-label="View log details"
                  class="p-1 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                  title="View detailed log entry"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<!-- Confirmation Modal for Safe Logging Provisioning -->
{#if showConfirmModal}
  <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 max-w-md w-full shadow-2xl space-y-4">
      <div class="flex items-center space-x-3 text-amber-600 dark:text-amber-400">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">Enable Caddy Logging?</h3>
      </div>

      <div class="text-xs text-slate-600 dark:text-slate-300 space-y-2 leading-relaxed">
        <p>Caddio will safely modify the active Caddy configuration to enable stdout JSON logging.</p>
        <ul class="list-disc list-inside space-y-1 text-slate-500 dark:text-slate-400 font-mono text-[11px]">
          <li>A pre-apply snapshot will be created automatically.</li>
          <li>Configuration will be validated before application.</li>
          <li>Optimistic concurrency checks will prevent race conditions.</li>
          <li>Existing sites and unmanaged routes will remain intact.</li>
        </ul>
      </div>

      <div class="flex items-center justify-end space-x-3 pt-2">
        <button
          type="button"
          on:click={() => (showConfirmModal = false)}
          class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-xs font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
        >
          Cancel
        </button>
        <button
          type="button"
          on:click={confirmEnableLogging}
          class="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium shadow-xs transition-colors"
        >
          Continue & Enable
        </button>
      </div>
    </div>
  </div>
{/if}
