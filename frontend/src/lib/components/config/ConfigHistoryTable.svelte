<script lang="ts">
  import type { ConfigHistoryItem } from '../../types';
  import { History, Eye, RotateCcw, Copy, Check, HardDrive, Calendar, User } from 'lucide-svelte';

  interface Props {
    items: ConfigHistoryItem[];
    loading?: boolean;
    onInspect: (item: ConfigHistoryItem) => void;
    onRollback: (item: ConfigHistoryItem) => void;
  }

  let { items = [], loading = false, onInspect, onRollback }: Props = $props();

  let copiedId = $state<string | null>(null);

  function copyHash(hash: string, id: string) {
    navigator.clipboard.writeText(hash);
    copiedId = id;
    setTimeout(() => {
      copiedId = null;
    }, 2000);
  }

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function formatDate(dateStr: string): string {
    try {
      const d = new Date(dateStr);
      return d.toLocaleString();
    } catch {
      return dateStr;
    }
  }
</script>

<div class="border border-slate-200 dark:border-slate-800 rounded-lg overflow-hidden bg-white dark:bg-slate-900 shadow-xs">
  <div class="overflow-x-auto">
    <table class="w-full text-left text-xs">
      <thead class="bg-slate-50 dark:bg-slate-950 text-slate-500 dark:text-slate-400 font-mono uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
        <tr>
          <th scope="col" class="px-4 py-3">Snapshot ID / Time</th>
          <th scope="col" class="px-4 py-3">SHA-256 Hash</th>
          <th scope="col" class="px-4 py-3">Source</th>
          <th scope="col" class="px-4 py-3">Author</th>
          <th scope="col" class="px-4 py-3">Size</th>
          <th scope="col" class="px-4 py-3 text-right">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60 font-sans">
        {#if loading}
          <tr>
            <td colspan="6" class="px-4 py-8 text-center text-slate-500 dark:text-slate-400">
              <div class="flex items-center justify-center space-x-2">
                <History class="w-4 h-4 animate-spin text-purple-600 dark:text-purple-400" />
                <span>Loading configuration snapshot history...</span>
              </div>
            </td>
          </tr>
        {:else if items.length === 0}
          <tr>
            <td colspan="6" class="px-4 py-8 text-center text-slate-400 dark:text-slate-500">
              No historical configuration snapshots found.
            </td>
          </tr>
        {:else}
          {#each items as item (item.id)}
            <tr class="hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-mono">
                <div class="font-semibold text-slate-900 dark:text-slate-100">{item.id}</div>
                <div class="text-[11px] text-slate-500 dark:text-slate-400 flex items-center space-x-1 mt-0.5">
                  <Calendar class="w-3 h-3 text-slate-400 dark:text-slate-500" />
                  <span>{formatDate(item.created_at)}</span>
                </div>
              </td>
              <td class="px-4 py-3 font-mono">
                <button
                  type="button"
                  onclick={() => copyHash(item.config_hash, item.id)}
                  class="inline-flex items-center space-x-1 px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-sky-700 dark:text-sky-400 hover:border-sky-500/50 transition-colors cursor-pointer"
                >
                  <span>{item.config_hash.substring(0, 12)}...</span>
                  {#if copiedId === item.id}
                    <Check class="w-3 h-3 text-emerald-600 dark:text-emerald-400" />
                  {:else}
                    <Copy class="w-3 h-3 text-slate-400 dark:text-slate-500" />
                  {/if}
                </button>
              </td>
              <td class="px-4 py-3">
                {#if item.source === 'manual_apply'}
                  <span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono bg-sky-500/10 text-sky-700 dark:text-sky-300 border border-sky-500/20">
                    manual apply
                  </span>
                {:else if item.source === 'rollback'}
                  <span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono bg-purple-500/10 text-purple-700 dark:text-purple-300 border border-purple-500/20">
                    rollback
                  </span>
                {:else}
                  <span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
                    system backup
                  </span>
                {/if}
              </td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300">
                <div class="flex items-center space-x-1.5">
                  <User class="w-3.5 h-3.5 text-slate-400 dark:text-slate-500" />
                  <span>{item.created_by || 'system'}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400 font-mono">
                <div class="flex items-center space-x-1">
                  <HardDrive class="w-3.5 h-3.5 text-slate-400 dark:text-slate-500" />
                  <span>{formatBytes(item.size_bytes)}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-right">
                <div class="flex items-center justify-end space-x-2">
                  <button
                    type="button"
                    onclick={() => onInspect(item)}
                    class="inline-flex items-center space-x-1 px-2.5 py-1 rounded bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 text-xs font-medium transition-colors cursor-pointer"
                  >
                    <Eye class="w-3.5 h-3.5 text-sky-600 dark:text-sky-400" />
                    <span>Inspect</span>
                  </button>
                  <button
                    type="button"
                    onclick={() => onRollback(item)}
                    class="inline-flex items-center space-x-1 px-2.5 py-1 rounded bg-purple-50 hover:bg-purple-100 dark:bg-purple-950/60 dark:hover:bg-purple-900/60 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800/60 text-xs font-medium transition-colors cursor-pointer"
                  >
                    <RotateCcw class="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" />
                    <span>Rollback</span>
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</div>
