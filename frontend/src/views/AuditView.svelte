<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api/client';
  import type { AuditLog } from '../lib/types';
  import AuditActivityList from '../lib/components/dashboard/AuditActivityList.svelte';
  import Button from '../lib/components/common/Button.svelte';
  import { ShieldCheck, RefreshCw } from 'lucide-svelte';

  let logs = $state<AuditLog[]>([]);
  let loading = $state<boolean>(true);
  let refreshing = $state<boolean>(false);
  let error = $state<string | undefined>(undefined);

  async function fetchLogs(isManual = false) {
    if (isManual) refreshing = true;
    error = undefined;

    try {
      logs = await api.getAuditLogs(50);
    } catch (err: any) {
      error = err.message || 'Failed to fetch audit logs';
    } finally {
      loading = false;
      refreshing = false;
    }
  }

  onMount(() => {
    fetchLogs(false);
  });
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200 dark:border-slate-800">
    <div>
      <div class="flex items-center gap-2">
        <ShieldCheck class="w-5 h-5 text-slate-700 dark:text-slate-300" />
        <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
          System Audit Log
        </h1>
      </div>
      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
        Complete history of administrative actions, authentication events, and system changes
      </p>
    </div>

    <div>
      <Button
        variant="outline"
        size="sm"
        onclick={() => fetchLogs(true)}
        loading={refreshing}
      >
        <RefreshCw class="w-3.5 h-3.5" />
        <span>Refresh</span>
      </Button>
    </div>
  </div>

  {#if error}
    <div class="p-4 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-600 dark:text-rose-400">
      <span class="font-semibold">Error:</span> {error}
    </div>
  {/if}

  <AuditActivityList {logs} {loading} />
</div>
