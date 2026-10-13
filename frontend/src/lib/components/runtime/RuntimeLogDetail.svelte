<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { RuntimeLogEntry } from '../../types';

  export let log: RuntimeLogEntry | null = null;

  const dispatch = createEventDispatcher<{ close: void }>();

  let copied = false;

  function close() {
    dispatch('close');
  }

  async function copyJson() {
    if (!log) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(log, null, 2));
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch (err) {
      console.error('Failed to copy JSON', err);
    }
  }

  function formatValue(val: unknown): string {
    if (val === null || val === undefined) return 'null';
    if (typeof val === 'object') return JSON.stringify(val, null, 2);
    return String(val);
  }

  function getLevelBadge(level: string | undefined): { bg: string; text: string } {
    switch (level?.toLowerCase()) {
      case 'error':
        return { bg: 'bg-rose-500/10 border-rose-500/20 text-rose-600 dark:text-rose-400', text: 'ERROR' };
      case 'warn':
      case 'warning':
        return { bg: 'bg-amber-500/10 border-amber-500/20 text-amber-600 dark:text-amber-400', text: 'WARN' };
      case 'info':
        return { bg: 'bg-blue-500/10 border-blue-500/20 text-blue-600 dark:text-blue-400', text: 'INFO' };
      default:
        return { bg: 'bg-slate-500/10 border-slate-500/20 text-slate-600 dark:text-slate-400', text: level?.toUpperCase() || 'DEBUG' };
    }
  }
</script>

{#if log}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs z-40 transition-opacity"
    on:click={close}
    on:keydown={(e) => e.key === 'Escape' && close()}
    role="button"
    tabindex="0"
    aria-label="Close modal overlay"
  ></div>

  <!-- Modal / Drawer -->
  <div class="fixed inset-y-0 right-0 max-w-2xl w-full bg-white dark:bg-slate-900 border-l border-slate-200 dark:border-slate-800 shadow-2xl z-50 flex flex-col overflow-hidden">
    <!-- Header -->
    <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/50">
      <div class="flex items-center space-x-3">
        <span class="px-2.5 py-0.5 rounded-md text-xs font-mono font-semibold uppercase border {getLevelBadge(log.level).bg}">
          {getLevelBadge(log.level).text}
        </span>
        <div>
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">Log Entry Details</h3>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-mono">{log.timestamp}</p>
        </div>
      </div>

      <div class="flex items-center space-x-2">
        <button
          type="button"
          on:click={copyJson}
          class="px-3 py-1.5 text-xs font-medium rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center space-x-1.5"
        >
          {#if copied}
            <svg class="w-3.5 h-3.5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <span class="text-emerald-600 dark:text-emerald-400">Copied!</span>
          {:else}
            <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <span>Copy Raw JSON</span>
          {/if}
        </button>

        <button
          type="button"
          on:click={close}
          aria-label="Close modal"
          class="p-1.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-6 space-y-6">
      <div class="space-y-4">
        <div>
          <span class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider block mb-1">Message</span>
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 text-sm font-mono text-slate-900 dark:text-slate-100 break-words">
            {log.message || '(No message)'}
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <span class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider block mb-1">Logger</span>
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 text-xs font-mono text-slate-800 dark:text-slate-200">
              {log.logger || 'default'}
            </div>
          </div>
          <div>
            <span class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider block mb-1">Level</span>
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800 text-xs font-mono text-slate-800 dark:text-slate-200">
              {log.level}
            </div>
          </div>
        </div>
      </div>

      {#if log.fields && Object.keys(log.fields).length > 0}
        <div>
          <h4 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">Structured Fields</h4>
          <div class="border border-slate-200 dark:border-slate-800 rounded-lg overflow-hidden divide-y divide-slate-100 dark:divide-slate-800">
            {#each Object.entries(log.fields) as [key, value]}
              <div class="flex flex-col sm:flex-row sm:items-start p-3 bg-white dark:bg-slate-900 text-xs">
                <span class="font-mono text-indigo-600 dark:text-indigo-400 font-semibold w-44 shrink-0 break-all mb-1 sm:mb-0">
                  {key}
                </span>
                <span class="font-mono text-slate-800 dark:text-slate-200 break-words whitespace-pre-wrap flex-1">
                  {formatValue(value)}
                </span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <div>
        <h4 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">Normalized JSON Payload</h4>
        <pre class="p-4 rounded-lg bg-slate-950 text-slate-200 font-mono text-xs overflow-x-auto border border-slate-800 leading-relaxed">{JSON.stringify(log, null, 2)}</pre>
      </div>
    </div>
  </div>
{/if}
