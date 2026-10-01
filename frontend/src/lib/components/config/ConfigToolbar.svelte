<script lang="ts">
  import {
    CheckCircle2, AlertTriangle, Eye, Send, RefreshCw, Code2, Copy, Check
  } from 'lucide-svelte';

  interface Props {
    activeHash: string;
    version?: string;
    status: 'clean' | 'modified' | 'valid' | 'invalid' | 'conflict';
    isValidating?: boolean;
    isDiffing?: boolean;
    isApplying?: boolean;
    onValidate: () => void;
    onPreviewDiff: () => void;
    onApply: () => void;
    onFormatJSON: () => void;
    onRefresh: () => void;
  }

  let {
    activeHash = '',
    version = '',
    status = 'clean',
    isValidating = false,
    isDiffing = false,
    isApplying = false,
    onValidate,
    onPreviewDiff,
    onApply,
    onFormatJSON,
    onRefresh
  }: Props = $props();

  let copiedHash = $state(false);

  function copyHash() {
    if (!activeHash) return;
    navigator.clipboard.writeText(activeHash);
    copiedHash = true;
    setTimeout(() => {
      copiedHash = false;
    }, 2000);
  }
</script>

<div class="flex flex-wrap items-center justify-between gap-3 p-4 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg shadow-xs">
  <!-- Left Side: Active Hash & Status Badge -->
  <div class="flex items-center space-x-3 flex-wrap gap-y-2">
    <div class="flex items-center space-x-2">
      <span class="text-xs font-medium text-slate-500 dark:text-slate-400">Active Hash:</span>
      <button
        type="button"
        onclick={copyHash}
        class="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 font-mono text-xs text-sky-700 dark:text-sky-400 hover:border-sky-500/50 transition-colors cursor-pointer"
        title="Click to copy full SHA-256 hash"
      >
        <span>{activeHash ? activeHash.substring(0, 12) + '...' : 'Loading...'}</span>
        {#if copiedHash}
          <Check class="w-3 h-3 text-emerald-600 dark:text-emerald-400" />
        {:else}
          <Copy class="w-3 h-3 text-slate-400 dark:text-slate-500" />
        {/if}
      </button>
    </div>

    {#if version}
      <span class="text-xs px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700 font-mono">
        Caddy {version}
      </span>
    {/if}

    <!-- Workspace Status Badge -->
    <div class="flex items-center">
      {#if status === 'conflict'}
        <span class="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-mono bg-rose-500/10 border border-rose-500/30 text-rose-700 dark:text-rose-300">
          <AlertTriangle class="w-3.5 h-3.5 text-rose-600 dark:text-rose-400" />
          <span>409 Conflict Detected</span>
        </span>
      {:else if status === 'invalid'}
        <span class="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-mono bg-amber-500/10 border border-amber-500/30 text-amber-700 dark:text-amber-300">
          <AlertTriangle class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" />
          <span>Invalid Config</span>
        </span>
      {:else if status === 'valid'}
        <span class="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-mono bg-emerald-500/10 border border-emerald-500/30 text-emerald-700 dark:text-emerald-300">
          <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
          <span>Validated</span>
        </span>
      {:else if status === 'modified'}
        <span class="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-mono bg-sky-500/10 border border-sky-500/30 text-sky-700 dark:text-sky-300">
          <span class="w-2 h-2 rounded-full bg-sky-500 dark:bg-sky-400 animate-pulse"></span>
          <span>Unsaved Changes</span>
        </span>
      {:else}
        <span class="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
          <CheckCircle2 class="w-3.5 h-3.5 text-slate-400 dark:text-slate-500" />
          <span>In Sync</span>
        </span>
      {/if}
    </div>
  </div>

  <!-- Right Side: Action Controls -->
  <div class="flex items-center space-x-2 flex-wrap gap-y-2">
    <button
      type="button"
      onclick={onRefresh}
      class="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors cursor-pointer"
      title="Reload active configuration from Caddy"
    >
      <RefreshCw class="w-3.5 h-3.5 text-slate-500 dark:text-slate-400" />
      <span>Reload</span>
    </button>

    <button
      type="button"
      onclick={onFormatJSON}
      class="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors cursor-pointer"
    >
      <Code2 class="w-3.5 h-3.5 text-slate-500 dark:text-slate-400" />
      <span>Format</span>
    </button>

    <button
      type="button"
      onclick={onValidate}
      disabled={isValidating}
      class="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 border border-slate-300 dark:border-slate-700 transition-colors cursor-pointer disabled:opacity-50"
    >
      {#if isValidating}
        <RefreshCw class="w-3.5 h-3.5 animate-spin text-sky-600 dark:text-sky-400" />
      {:else}
        <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
      {/if}
      <span>Validate</span>
    </button>

    <button
      type="button"
      onclick={onPreviewDiff}
      disabled={isDiffing}
      class="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-sky-50 hover:bg-sky-100 dark:bg-sky-950/60 dark:hover:bg-sky-900/60 text-sky-700 dark:text-sky-300 border border-sky-200 dark:border-sky-800/60 transition-colors cursor-pointer disabled:opacity-50"
    >
      {#if isDiffing}
        <RefreshCw class="w-3.5 h-3.5 animate-spin text-sky-600 dark:text-sky-400" />
      {:else}
        <Eye class="w-3.5 h-3.5 text-sky-600 dark:text-sky-400" />
      {/if}
      <span>Preview Diff</span>
    </button>

    <button
      type="button"
      onclick={onApply}
      disabled={isApplying}
      class="inline-flex items-center space-x-1.5 px-4 py-1.5 rounded-md text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white shadow-xs transition-colors cursor-pointer disabled:opacity-50"
    >
      {#if isApplying}
        <RefreshCw class="w-3.5 h-3.5 animate-spin" />
        <span>Applying...</span>
      {:else}
        <Send class="w-3.5 h-3.5" />
        <span>Apply Changes</span>
      {/if}
    </button>
  </div>
</div>
