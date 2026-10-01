<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api/client';
  import type { ConfigResponse, ConfigDiffResponse } from '../lib/types';
  import ConfigEditor from '../lib/components/config/ConfigEditor.svelte';
  import ConfigDiff from '../lib/components/config/ConfigDiff.svelte';
  import ConfigToolbar from '../lib/components/config/ConfigToolbar.svelte';
  import ApplyConfigDialog from '../lib/components/config/ApplyConfigDialog.svelte';
  import { History, RefreshCw, AlertTriangle, CheckCircle2, Sliders } from 'lucide-svelte';

  let configRes = $state<ConfigResponse | null>(null);
  let editorContent = $state('');
  let initialHash = $state('');
  let currentServerHash = $state('');
  let caddyVersion = $state('');
  
  let isLoading = $state(true);
  let isValidating = $state(false);
  let isDiffing = $state(false);
  let isApplying = $state(false);
  
  let errorMessage = $state<string | null>(null);
  let successMessage = $state<string | null>(null);
  
  let validation = $state<{ valid: boolean; message: string; details?: string[] } | null>(null);
  let diffRes = $state<ConfigDiffResponse | null>(null);
  
  let showApplyDialog = $state(false);
  let activeTab = $state<'split' | 'editor' | 'diff'>('split');
  let editorRef = $state<any>(null);

  let workspaceStatus = $derived<'clean' | 'modified' | 'valid' | 'invalid' | 'conflict'>(
    (initialHash && currentServerHash && initialHash !== currentServerHash)
      ? 'conflict'
      : (validation && !validation.valid)
        ? 'invalid'
        : (validation && validation.valid)
          ? 'valid'
          : (configRes && editorContent !== configRes.config)
            ? 'modified'
            : 'clean'
  );

  onMount(() => {
    loadConfig();
  });

  async function loadConfig() {
    isLoading = true;
    errorMessage = null;
    try {
      const res = await api.getConfig();
      configRes = res;
      editorContent = res.config;
      initialHash = res.hash;
      currentServerHash = res.hash;
      caddyVersion = res.version || '';
      validation = null;
      diffRes = null;
    } catch (err: any) {
      errorMessage = err.message || 'Failed to fetch active Caddy configuration';
    } finally {
      isLoading = false;
    }
  }

  async function handleValidate() {
    isValidating = true;
    errorMessage = null;
    successMessage = null;
    try {
      const res = await api.validateConfig({
        base_hash: initialHash,
        config: editorContent
      });
      validation = {
        valid: res.valid,
        message: res.message,
        details: res.details
      };
      if (res.current_hash) {
        currentServerHash = res.current_hash;
      }
      if (res.has_conflict) {
        errorMessage = 'Conflict detected: Caddy configuration was modified on the server since session loaded.';
      } else if (res.valid) {
        successMessage = 'Configuration syntax and Caddy structure validation passed!';
      }
    } catch (err: any) {
      validation = {
        valid: false,
        message: err.message || 'Validation request failed'
      };
    } finally {
      isValidating = false;
    }
  }

  async function handlePreviewDiff() {
    isDiffing = true;
    errorMessage = null;
    try {
      const res = await api.diffConfig({
        base_hash: initialHash,
        config: editorContent
      });
      diffRes = res;
      currentServerHash = res.current_hash;
      if (res.has_conflict) {
        errorMessage = 'Conflict detected: Base hash does not match current server hash.';
      }
    } catch (err: any) {
      errorMessage = err.message || 'Failed to compute backend diff';
    } finally {
      isDiffing = false;
    }
  }

  function handleOpenApply() {
    errorMessage = null;
    showApplyDialog = true;
  }

  async function handleConfirmApply() {
    isApplying = true;
    errorMessage = null;
    try {
      const res = await api.applyConfig({
        base_hash: initialHash,
        config: editorContent
      });
      configRes = res;
      editorContent = res.config;
      initialHash = res.hash;
      currentServerHash = res.hash;
      caddyVersion = res.version || caddyVersion;
      diffRes = null;
      validation = null;
      showApplyDialog = false;
      successMessage = 'Caddy configuration successfully applied and verified!';
      setTimeout(() => {
        successMessage = null;
      }, 5000);
    } catch (err: any) {
      errorMessage = err.message || 'Failed to apply Caddy configuration';
    } finally {
      isApplying = false;
    }
  }

  function handleFormatJSON() {
    if (editorRef) {
      editorRef.formatJSON();
    }
  }
</script>

<div class="space-y-6">
  <!-- Page Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center space-x-2 tracking-tight">
        <Sliders class="w-5 h-5 text-sky-600 dark:text-sky-400" />
        <span>Configuration Workspace</span>
      </h1>
      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
        Safe Caddy JSON editor with optimistic concurrency protection, dry-run validation, and line diff preview.
      </p>
    </div>

    <div class="flex items-center space-x-3">
      <a
        href="/config/history"
        class="inline-flex items-center space-x-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 transition-colors shadow-xs"
      >
        <History class="w-4 h-4 text-purple-600 dark:text-purple-400" />
        <span>History & Snapshots</span>
      </a>
    </div>
  </div>

  {#if successMessage}
    <div class="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-800 dark:text-emerald-300 font-mono text-xs flex items-center justify-between animate-in fade-in">
      <div class="flex items-center space-x-2">
        <CheckCircle2 class="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0" />
        <span>{successMessage}</span>
      </div>
      <button type="button" onclick={() => successMessage = null} class="text-emerald-600 dark:text-emerald-400 hover:opacity-75">
        &times;
      </button>
    </div>
  {/if}

  {#if errorMessage}
    <div class="p-4 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-800 dark:text-rose-300 font-mono text-xs flex items-start justify-between animate-in fade-in">
      <div class="flex items-start space-x-2">
        <AlertTriangle class="w-4 h-4 text-rose-600 dark:text-rose-400 shrink-0 mt-0.5" />
        <span class="break-all">{errorMessage}</span>
      </div>
      <button type="button" onclick={() => errorMessage = null} class="text-rose-600 dark:text-rose-400 hover:opacity-75 shrink-0 ml-2">
        &times;
      </button>
    </div>
  {/if}

  <!-- Toolbar -->
  <ConfigToolbar
    activeHash={initialHash}
    version={caddyVersion}
    status={workspaceStatus}
    {isValidating}
    {isDiffing}
    {isApplying}
    onValidate={handleValidate}
    onPreviewDiff={handlePreviewDiff}
    onApply={handleOpenApply}
    onFormatJSON={handleFormatJSON}
    onRefresh={loadConfig}
  />

  <!-- Validation Status Box -->
  {#if validation}
    <div
      class="p-4 rounded-lg border text-xs font-mono space-y-1 {validation.valid ? 'bg-emerald-50 dark:bg-emerald-950/30 border-emerald-200 dark:border-emerald-800/50 text-emerald-800 dark:text-emerald-300' : 'bg-rose-50 dark:bg-rose-950/30 border-rose-200 dark:border-rose-800/50 text-rose-800 dark:text-rose-300'}"
    >
      <div class="font-bold flex items-center space-x-2">
        {#if validation.valid}
          <CheckCircle2 class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
        {:else}
          <AlertTriangle class="w-4 h-4 text-rose-600 dark:text-rose-400" />
        {/if}
        <span>{validation.message}</span>
      </div>
      {#if validation.details && validation.details.length > 0}
        <ul class="list-disc list-inside space-y-0.5 pt-1 text-slate-700 dark:text-slate-300">
          {#each validation.details as detail}
            <li>{detail}</li>
          {/each}
        </ul>
      {/if}
    </div>
  {/if}

  <!-- View Mode Switcher for Editor & Diff -->
  <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-2">
    <div class="flex items-center space-x-1 bg-slate-100 dark:bg-slate-900 p-1 rounded-lg border border-slate-200 dark:border-slate-800 text-xs">
      <button
        type="button"
        onclick={() => activeTab = 'split'}
        class="px-3 py-1 rounded-md font-medium transition-colors cursor-pointer"
        class:bg-white={activeTab === 'split'}
        class:dark:bg-slate-800={activeTab === 'split'}
        class:text-slate-900={activeTab === 'split'}
        class:dark:text-slate-100={activeTab === 'split'}
        class:shadow-xs={activeTab === 'split'}
        class:text-slate-600={activeTab !== 'split'}
        class:dark:text-slate-400={activeTab !== 'split'}
      >
        Split View
      </button>
      <button
        type="button"
        onclick={() => activeTab = 'editor'}
        class="px-3 py-1 rounded-md font-medium transition-colors cursor-pointer"
        class:bg-white={activeTab === 'editor'}
        class:dark:bg-slate-800={activeTab === 'editor'}
        class:text-slate-900={activeTab === 'editor'}
        class:dark:text-slate-100={activeTab === 'editor'}
        class:shadow-xs={activeTab === 'editor'}
        class:text-slate-600={activeTab !== 'editor'}
        class:dark:text-slate-400={activeTab !== 'editor'}
      >
        Editor Only
      </button>
      <button
        type="button"
        onclick={() => activeTab = 'diff'}
        class="px-3 py-1 rounded-md font-medium transition-colors cursor-pointer"
        class:bg-white={activeTab === 'diff'}
        class:dark:bg-slate-800={activeTab === 'diff'}
        class:text-slate-900={activeTab === 'diff'}
        class:dark:text-slate-100={activeTab === 'diff'}
        class:shadow-xs={activeTab === 'diff'}
        class:text-slate-600={activeTab !== 'diff'}
        class:dark:text-slate-400={activeTab !== 'diff'}
      >
        Diff Preview
      </button>
    </div>
  </div>

  <!-- Workspace Grid Layout -->
  {#if isLoading}
    <div class="flex items-center justify-center h-80 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg text-slate-500 dark:text-slate-400 space-x-2">
      <RefreshCw class="w-6 h-6 animate-spin text-sky-600 dark:text-sky-400" />
      <span>Loading active Caddy JSON configuration...</span>
    </div>
  {:else}
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 min-h-[500px]">
      {#if activeTab === 'split' || activeTab === 'editor'}
        <div class="h-full" class:col-span-2={activeTab === 'editor'}>
          <ConfigEditor bind:this={editorRef} bind:value={editorContent} />
        </div>
      {/if}

      {#if activeTab === 'split' || activeTab === 'diff'}
        <div class="h-full" class:col-span-2={activeTab === 'diff'}>
          <ConfigDiff diff={diffRes} loading={isDiffing} />
        </div>
      {/if}
    </div>
  {/if}

  <!-- Apply Modal Dialog -->
  <ApplyConfigDialog
    open={showApplyDialog}
    baseHash={initialHash}
    currentHash={currentServerHash}
    diff={diffRes}
    {isApplying}
    error={errorMessage}
    onConfirm={handleConfirmApply}
    onCancel={() => showApplyDialog = false}
  />
</div>
