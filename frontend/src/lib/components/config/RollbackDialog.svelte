<script lang="ts">
  import type { ConfigHistoryItem } from '../../types';
  import { RotateCcw, AlertTriangle, X, RefreshCw } from 'lucide-svelte';

  interface Props {
    open: boolean;
    snapshotItem: ConfigHistoryItem | null;
    activeHash: string;
    isRollingBack?: boolean;
    error?: string | null;
    onConfirm: () => void;
    onCancel: () => void;
  }

  let {
    open = false,
    snapshotItem,
    activeHash = '',
    isRollingBack = false,
    error = null,
    onConfirm,
    onCancel
  }: Props = $props();
</script>

{#if open && snapshotItem}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-xs">
    <div class="w-full max-w-lg bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/90">
        <div class="flex items-center space-x-2 text-slate-900 dark:text-slate-100 font-semibold text-base">
          <RotateCcw class="w-5 h-5 text-purple-600 dark:text-purple-400" />
          <span>Confirm Configuration Rollback</span>
        </div>
        <button
          type="button"
          onclick={onCancel}
          disabled={isRollingBack}
          class="p-1 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-4 text-sm">
        <!-- Target Snapshot Info -->
        <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 space-y-2 font-mono text-xs">
          <div class="flex justify-between text-slate-500 dark:text-slate-400">
            <span>Snapshot ID:</span>
            <span class="text-slate-900 dark:text-slate-200 font-bold">{snapshotItem.id}</span>
          </div>
          <div class="flex justify-between text-slate-500 dark:text-slate-400">
            <span>Target Hash:</span>
            <span class="text-sky-700 dark:text-sky-400">{snapshotItem.config_hash.substring(0, 16)}...</span>
          </div>
          <div class="flex justify-between text-slate-500 dark:text-slate-400">
            <span>Active Hash:</span>
            <span class="text-amber-700 dark:text-amber-400">{activeHash.substring(0, 16)}...</span>
          </div>
        </div>

        <!-- Warning Alert -->
        <div class="p-4 rounded-lg bg-purple-500/10 border border-purple-500/30 text-purple-800 dark:text-purple-200 space-y-2 text-xs">
          <div class="font-bold text-purple-700 dark:text-purple-300 flex items-center space-x-2">
            <AlertTriangle class="w-4 h-4 text-purple-600 dark:text-purple-400 shrink-0" />
            <span>Rollback Action</span>
          </div>
          <p class="leading-relaxed">
            Rolling back will replace Caddy's current active configuration with snapshot <strong>{snapshotItem.id}</strong>.
          </p>
          <p class="leading-relaxed">
            Caddio will automatically create a backup snapshot of your currently active configuration before applying the rollback.
          </p>
        </div>

        {#if error}
          <div class="p-3 rounded bg-rose-50 dark:bg-rose-950/60 border border-rose-200 dark:border-rose-900 text-rose-700 dark:text-rose-300 text-xs font-mono break-all">
            {error}
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end space-x-3 px-6 py-4 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800">
        <button
          type="button"
          onclick={onCancel}
          disabled={isRollingBack}
          class="px-4 py-2 rounded-md text-xs font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 transition-colors cursor-pointer disabled:opacity-50"
        >
          Cancel
        </button>
        <button
          type="button"
          onclick={onConfirm}
          disabled={isRollingBack}
          class="inline-flex items-center space-x-2 px-5 py-2 rounded-md text-xs font-semibold text-white bg-purple-600 hover:bg-purple-500 shadow-md transition-colors cursor-pointer disabled:opacity-50"
        >
          {#if isRollingBack}
            <RefreshCw class="w-4 h-4 animate-spin" />
            <span>Restoring Snapshot...</span>
          {:else}
            <RotateCcw class="w-4 h-4" />
            <span>Confirm & Rollback</span>
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
