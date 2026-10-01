<script lang="ts">
  import type { Site } from '../../types';
  import Button from '../common/Button.svelte';
  import { api } from '../../api/client';
  import { AlertTriangle, Trash2, X } from 'lucide-svelte';

  interface Props {
    open: boolean;
    site: Site | null;
    onClose: () => void;
    onSuccess: () => void;
  }

  let { open = false, site = null, onClose, onSuccess }: Props = $props();

  let deleting = $state(false);
  let errorMessage = $state('');

  async function handleDelete() {
    if (!site) return;
    deleting = true;
    errorMessage = '';

    try {
      await api.deleteSite(site.id);
      onSuccess();
      onClose();
    } catch (err: any) {
      errorMessage = err.message || 'Failed to delete site from Caddy';
    } finally {
      deleting = false;
    }
  }
</script>

{#if open && site}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-y-auto">
    <!-- Backdrop -->
    <div
      class="fixed inset-0 bg-slate-950/60 backdrop-blur-xs transition-opacity"
      onclick={onClose}
      aria-hidden="true"
    ></div>

    <!-- Modal Card -->
    <div class="relative w-full max-w-md bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-xl overflow-hidden z-10 my-auto p-6 space-y-4">
      <div class="flex items-start justify-between gap-3">
        <div class="p-3 rounded-full bg-rose-500/10 text-rose-600 dark:text-rose-400 shrink-0">
          <AlertTriangle class="w-6 h-6" />
        </div>
        <button
          type="button"
          onclick={onClose}
          class="p-1.5 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <div class="space-y-2">
        <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">
          Delete Reverse Proxy Site?
        </h3>
        <p class="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
          Are you sure you want to delete <span class="font-mono font-semibold text-slate-900 dark:text-slate-100 bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded">{site.domain}</span>?
          This action will immediately remove the HTTP/HTTPS route from Caddy.
        </p>
      </div>

      {#if errorMessage}
        <div class="p-3 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-700 dark:text-rose-300 leading-relaxed">
          {errorMessage}
        </div>
      {/if}

      <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
        <Button variant="secondary" size="sm" onclick={onClose} disabled={deleting}>
          Cancel
        </Button>

        <Button variant="danger" size="sm" onclick={handleDelete} loading={deleting}>
          <Trash2 class="w-3.5 h-3.5" />
          <span>Delete & Apply</span>
        </Button>
      </div>
    </div>
  </div>
{/if}
