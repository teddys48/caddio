<script lang="ts">
  import type { ConfigDiffResponse } from '../../types';
  import { PlusCircle, MinusCircle, RefreshCw, AlertTriangle, CheckCircle2 } from 'lucide-svelte';

  interface Props {
    diff: ConfigDiffResponse | null;
    loading?: boolean;
    error?: string | null;
  }

  let { diff, loading = false, error = null }: Props = $props();

  function parseUnifiedDiffLines(unifiedDiff: string) {
    if (!unifiedDiff) return [];
    return unifiedDiff.split('\n').map((line, idx) => {
      let type: 'header' | 'add' | 'delete' | 'context' = 'context';
      if (line.startsWith('@@')) {
        type = 'header';
      } else if (line.startsWith('+')) {
        type = 'add';
      } else if (line.startsWith('-')) {
        type = 'delete';
      }
      return { id: idx, type, text: line };
    });
  }

  let lines = $derived(diff ? parseUnifiedDiffLines(diff.unified_diff) : []);
</script>

<div class="flex flex-col h-full border border-slate-200 dark:border-slate-800 rounded-lg overflow-hidden bg-white dark:bg-[#090d16] shadow-xs">
  <!-- Header with Stats -->
  <div class="flex flex-wrap items-center justify-between px-4 py-3 bg-slate-50 dark:bg-slate-900/90 border-b border-slate-200 dark:border-slate-800 text-xs gap-2">
    <div class="flex items-center space-x-3 font-mono">
      <span class="text-slate-700 dark:text-slate-300 font-semibold">Diff Preview</span>
      {#if diff}
        <span class="inline-flex items-center space-x-1 text-emerald-700 dark:text-emerald-400 font-medium">
          <PlusCircle class="w-3.5 h-3.5" />
          <span>+{diff.stats.additions}</span>
        </span>
        <span class="inline-flex items-center space-x-1 text-rose-700 dark:text-rose-400 font-medium">
          <MinusCircle class="w-3.5 h-3.5" />
          <span>-{diff.stats.deletions}</span>
        </span>
        <span class="inline-flex items-center space-x-1 text-amber-700 dark:text-amber-400 font-medium">
          <RefreshCw class="w-3.5 h-3.5" />
          <span>~{diff.stats.modified}</span>
        </span>
      {/if}
    </div>

    {#if diff?.has_conflict}
      <div class="flex items-center space-x-1.5 px-2.5 py-1 rounded bg-rose-500/10 border border-rose-500/30 text-rose-700 dark:text-rose-300 font-mono text-xs">
        <AlertTriangle class="w-3.5 h-3.5 text-rose-600 dark:text-rose-400" />
        <span>Optimistic Concurrency Conflict (Base hash mismatch)</span>
      </div>
    {:else if diff && diff.stats.additions === 0 && diff.stats.deletions === 0 && diff.stats.modified === 0}
      <div class="flex items-center space-x-1.5 px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 font-mono text-xs">
        <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
        <span>No changes detected</span>
      </div>
    {/if}
  </div>

  <!-- Content Container -->
  <div class="flex-1 overflow-auto p-4 font-mono text-xs min-h-[350px]">
    {#if loading}
      <div class="flex items-center justify-center h-48 text-slate-500 dark:text-slate-400 space-x-2">
        <RefreshCw class="w-5 h-5 animate-spin text-sky-600 dark:text-sky-400" />
        <span>Computing structured line diff from active Caddy configuration...</span>
      </div>
    {:else if error}
      <div class="p-4 rounded-md bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900/60 text-rose-700 dark:text-rose-300 font-mono space-y-1">
        <div class="font-bold flex items-center space-x-2">
          <AlertTriangle class="w-4 h-4 text-rose-600 dark:text-rose-400" />
          <span>Diff Generation Error</span>
        </div>
        <div class="break-all">{error}</div>
      </div>
    {:else if !diff || lines.length === 0}
      <div class="flex items-center justify-center h-48 text-slate-400 dark:text-slate-500 font-sans text-sm">
        Click "Preview Diff" in the workspace toolbar to generate structured line changes.
      </div>
    {:else}
      <div class="space-y-0.5 rounded border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-[#070a10] overflow-hidden">
        {#each lines as line (line.id)}
          <div
            class="px-3 py-0.5 whitespace-pre font-mono leading-relaxed transition-colors flex items-start {line.type === 'add' ? 'bg-emerald-100/70 dark:bg-emerald-950/30 text-emerald-900 dark:text-emerald-300' : ''} {line.type === 'delete' ? 'bg-rose-100/70 dark:bg-rose-950/30 text-rose-900 dark:text-rose-300' : ''} {line.type === 'header' ? 'bg-sky-100/80 dark:bg-slate-800/50 text-sky-800 dark:text-sky-400' : ''} {line.type === 'context' ? 'text-slate-700 dark:text-slate-400' : ''}"
          >
            <span class="select-none text-slate-400 dark:text-slate-600 w-6 shrink-0 text-right pr-2">
              {#if line.type === 'add'}+{:else if line.type === 'delete'}-{:else if line.type === 'header'}@{/if}
            </span>
            <span class="break-all">{line.text}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
