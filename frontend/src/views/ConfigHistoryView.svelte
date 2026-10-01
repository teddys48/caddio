<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api/client';
  import type { ConfigHistoryItem, ConfigSnapshot } from '../lib/types';
  import ConfigHistoryTable from '../lib/components/config/ConfigHistoryTable.svelte';
  import RollbackDialog from '../lib/components/config/RollbackDialog.svelte';
  import ConfigEditor from '../lib/components/config/ConfigEditor.svelte';
  import { History, Sliders, RefreshCw, AlertTriangle, CheckCircle2, X, Eye } from 'lucide-svelte';

  let items = $state<ConfigHistoryItem[]>([]);
  let activeHash = $state('');
  let isLoading = $state(true);
  
  let selectedInspectSnapshot = $state<ConfigSnapshot | null>(null);
  let isInspectLoading = $state(false);
  let showInspectModal = $state(false);

  let rollbackItem = $state<ConfigHistoryItem | null>(null);
  let showRollbackModal = $state(false);
  let isRollingBack = $state(false);

  let successMessage = $state<string | null>(null);
  let errorMessage = $state<string | null>(null);

  onMount(() => {
    loadHistoryData();
  });

  async function loadHistoryData() {
    isLoading = true;
    errorMessage = null;
    try {
      const [historyList, activeCfg] = await Promise.all([
        api.getConfigHistory(),
        api.getConfig()
      ]);
      items = historyList;
      activeHash = activeCfg.hash;
    } catch (err: any) {
      errorMessage = err.message || 'Failed to load configuration history';
    } finally {
      isLoading = false;
    }
  }

  async function handleInspect(item: ConfigHistoryItem) {
    isInspectLoading = true;
    showInspectModal = true;
    selectedInspectSnapshot = null;
    try {
      const snap = await api.getConfigSnapshot(item.id);
      selectedInspectSnapshot = snap;
    } catch (err: any) {
      errorMessage = err.message || 'Failed to fetch snapshot details';
      showInspectModal = false;
    } finally {
      isInspectLoading = false;
    }
  }

  function handleOpenRollback(item: ConfigHistoryItem) {
    rollbackItem = item;
    showRollbackModal = true;
  }

  async function handleConfirmRollback() {
    if (!rollbackItem) return;
    isRollingBack = true;
    errorMessage = null;
    try {
      const res = await api.rollbackConfig(rollbackItem.id, {
        base_hash: activeHash
      });
      activeHash = res.hash;
      showRollbackModal = false;
      successMessage = `Successfully rolled back to configuration snapshot ${rollbackItem.id}!`;
      await loadHistoryData();
      setTimeout(() => {
        successMessage = null;
      }, 5000);
    } catch (err: any) {
      errorMessage = err.message || 'Rollback execution failed';
    } finally {
      isRollingBack = false;
    }
  }
</script>

<div class="space-y-6">
  <!-- Page Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center space-x-2 tracking-tight">
        <History class="w-5 h-5 text-purple-600 dark:text-purple-400" />
        <span>Configuration History & Snapshots</span>
      </h1>
      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
        Inspect historical Caddy JSON snapshots stored in SQLite, review audit signatures, and safely rollback.
      </p>
    </div>

    <div class="flex items-center space-x-3">
      <button
        type="button"
        onclick={loadHistoryData}
        class="inline-flex items-center space-x-1.5 px-3 py-2 rounded-lg text-xs font-medium bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors shadow-xs cursor-pointer"
      >
        <RefreshCw class="w-3.5 h-3.5 text-slate-500 dark:text-slate-400" />
        <span>Refresh</span>
      </button>

      <a
        href="/config"
        class="inline-flex items-center space-x-2 px-3.5 py-2 rounded-lg text-xs font-semibold bg-sky-600 hover:bg-sky-500 text-white shadow-xs transition-colors"
      >
        <Sliders class="w-4 h-4" />
        <span>Open Editor Workspace</span>
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

  <!-- History Table -->
  <ConfigHistoryTable
    {items}
    loading={isLoading}
    onInspect={handleInspect}
    onRollback={handleOpenRollback}
  />

  <!-- Inspect Snapshot Modal -->
  {#if showInspectModal}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-xs">
      <div class="w-full max-w-4xl max-h-[90vh] bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in duration-200">
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/90">
          <div class="flex items-center space-x-2 text-slate-900 dark:text-slate-100 font-semibold text-base font-mono">
            <Eye class="w-5 h-5 text-sky-600 dark:text-sky-400" />
            <span>Snapshot {selectedInspectSnapshot?.id || ''}</span>
          </div>
          <button
            type="button"
            onclick={() => showInspectModal = false}
            class="p-1 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 flex-1 overflow-auto space-y-4">
          {#if isInspectLoading}
            <div class="flex items-center justify-center h-48 text-slate-500 dark:text-slate-400 space-x-2">
              <RefreshCw class="w-5 h-5 animate-spin text-sky-600 dark:text-sky-400" />
              <span>Fetching snapshot content...</span>
            </div>
          {:else if selectedInspectSnapshot}
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-xs font-mono">
              <div>
                <div class="text-slate-500 dark:text-slate-400">Source</div>
                <div class="text-slate-900 dark:text-slate-200 font-bold">{selectedInspectSnapshot.source}</div>
              </div>
              <div>
                <div class="text-slate-500 dark:text-slate-400">Author</div>
                <div class="text-slate-900 dark:text-slate-200 font-bold">{selectedInspectSnapshot.created_by}</div>
              </div>
              <div>
                <div class="text-slate-500 dark:text-slate-400">Hash</div>
                <div class="text-sky-600 dark:text-sky-400 font-bold">{selectedInspectSnapshot.config_hash.substring(0, 12)}...</div>
              </div>
              <div>
                <div class="text-slate-500 dark:text-slate-400">Size</div>
                <div class="text-slate-900 dark:text-slate-200 font-bold">{selectedInspectSnapshot.size_bytes} B</div>
              </div>
            </div>

            <div class="h-[400px]">
              <ConfigEditor value={selectedInspectSnapshot.config || ''} readonly={true} />
            </div>
          {/if}
        </div>

        <div class="flex items-center justify-end px-6 py-4 bg-slate-50 dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800">
          <button
            type="button"
            onclick={() => showInspectModal = false}
            class="px-4 py-2 rounded-md text-xs font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 transition-colors cursor-pointer"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Rollback Modal Dialog -->
  <RollbackDialog
    open={showRollbackModal}
    snapshotItem={rollbackItem}
    {activeHash}
    {isRollingBack}
    error={errorMessage}
    onConfirm={handleConfirmRollback}
    onCancel={() => showRollbackModal = false}
  />
</div>
