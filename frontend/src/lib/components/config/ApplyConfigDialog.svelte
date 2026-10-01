<script lang="ts">
  import type { ConfigDiffResponse } from '../../types';
  import { AlertTriangle, Send, X, PlusCircle, MinusCircle, RefreshCw } from 'lucide-svelte';

  interface Props {
    open: boolean;
    baseHash: string;
    currentHash: string;
    diff: ConfigDiffResponse | null;
    isApplying?: boolean;
    error?: string | null;
    onConfirm: () => void;
    onCancel: () => void;
  }

  let {
    open = false,
    baseHash = '',
    currentHash = '',
    diff,
    isApplying = false,
    error = null,
    onConfirm,
    onCancel
  }: Props = $props();

  let hasConflict = $derived(baseHash !== '' && baseHash !== currentHash);
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-xs">
    <div class="w-full max-w-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
      <!-- Modal Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/90">
        <div class="flex items-center space-x-2 text-slate-900 dark:text-slate-100 font-semibold text-base">
          <AlertTriangle class="w-5 h-5 text-amber-600 dark:text-amber-400" />
          <span>Confirm Configuration Apply</span>
        </div>
        <button
          type="button"
          onclick={onCancel}
          disabled={isApplying}
          class="p-1 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 space-y-4 text-sm">
        <!-- Whole Config Replacement Alert -->
        <div class="p-4 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-800 dark:text-amber-200 space-y-2">
          <div class="font-semibold text-amber-700 dark:text-amber-300 flex items-center space-x-2">
            <AlertTriangle class="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0" />
            <span>Whole-Config Replacement Warning</span>
          </div>
          <p class="text-xs leading-relaxed">
            Applying raw JSON will execute Caddy's <code>POST /load</code> endpoint, performing an intentional <strong>whole-configuration replacement</strong> of Caddy's active runtime state.
          </p>
          <p class="text-xs leading-relaxed">
            Caddio will automatically create a pre-apply backup snapshot in SQLite prior to applying.
          </p>
        </div>

        <!-- Concurrency Conflict Alert -->
        {#if hasConflict}
          <div class="p-4 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-800 dark:text-rose-300 space-y-2 font-mono text-xs">
            <div class="font-bold text-rose-700 dark:text-rose-400 flex items-center space-x-2">
              <AlertTriangle class="w-4 h-4 shrink-0" />
              <span>409 CONFIG_CONFLICT Detected</span>
            </div>
            <div>
              Active Caddy hash (<code>{currentHash.substring(0, 12)}...</code>) does not match base session hash (<code>{baseHash.substring(0, 12)}...</code>).
            </div>
            <div class="text-rose-700/90 dark:text-rose-200/90">
              Another administrator or site modification changed Caddy's configuration since you loaded this session. Reload before applying to prevent overwriting third-party updates.
            </div>
          </div>
        {/if}

        <!-- Diff Summary -->
        {#if diff}
          <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 space-y-2">
            <div class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">Change Summary</div>
            <div class="flex items-center space-x-4 font-mono text-xs">
              <span class="flex items-center space-x-1 text-emerald-700 dark:text-emerald-400">
                <PlusCircle class="w-4 h-4" />
                <span>+{diff.stats.additions} additions</span>
              </span>
              <span class="flex items-center space-x-1 text-rose-700 dark:text-rose-400">
                <MinusCircle class="w-4 h-4" />
                <span>-{diff.stats.deletions} deletions</span>
              </span>
              <span class="flex items-center space-x-1 text-amber-700 dark:text-amber-400">
                <RefreshCw class="w-4 h-4" />
                <span>~{diff.stats.modified} modified</span>
              </span>
            </div>
          </div>
        {/if}

        {#if error}
          <div class="p-3 rounded bg-rose-50 dark:bg-rose-950/60 border border-rose-200 dark:border-rose-900 text-rose-700 dark:text-rose-300 text-xs font-mono break-all">
            {error}
          </div>
        {/if}
      </div>

      <!-- Modal Footer -->
      <div class="flex items-center justify-end space-x-3 px-6 py-4 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800">
        <button
          type="button"
          onclick={onCancel}
          disabled={isApplying}
          class="px-4 py-2 rounded-md text-xs font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 transition-colors cursor-pointer disabled:opacity-50"
        >
          Cancel
        </button>
        <button
          type="button"
          onclick={onConfirm}
          disabled={isApplying || hasConflict}
          class="inline-flex items-center space-x-2 px-5 py-2 rounded-md text-xs font-semibold text-white bg-emerald-600 hover:bg-emerald-500 shadow-md transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {#if isApplying}
            <RefreshCw class="w-4 h-4 animate-spin" />
            <span>Applying & Verifying...</span>
          {:else}
            <Send class="w-4 h-4" />
            <span>Confirm & Apply Config</span>
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
