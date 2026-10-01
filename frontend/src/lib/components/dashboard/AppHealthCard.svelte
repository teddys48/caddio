<script lang="ts">
  import type { AppHealth } from '../../types';
  import Badge from '../common/Badge.svelte';
  import { Activity, Database, CheckCircle2, XCircle } from 'lucide-svelte';

  interface Props {
    health?: AppHealth;
  }

  let { health }: Props = $props();

  const isHealthy = $derived(health?.status === 'healthy');
  const dbConnected = $derived(health?.database === 'connected');
</script>

<div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-5 shadow-xs transition-colors">
  <div class="flex items-center justify-between gap-4 mb-4">
    <div class="flex items-center gap-2.5">
      <div class="p-2 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
        <Activity class="w-5 h-5" />
      </div>
      <div>
        <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100 tracking-tight">
          Application Health
        </h3>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
          Local services & database status
        </p>
      </div>
    </div>

    {#if isHealthy}
      <Badge variant="healthy">
        <CheckCircle2 class="w-3.5 h-3.5" />
        <span>Healthy</span>
      </Badge>
    {:else}
      <Badge variant="unhealthy">
        <XCircle class="w-3.5 h-3.5" />
        <span>Unhealthy</span>
      </Badge>
    {/if}
  </div>

  <div class="space-y-2.5 text-xs">
    <div class="p-3 rounded-md bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 flex items-center justify-between">
      <div class="flex items-center gap-2 text-slate-700 dark:text-slate-300 font-medium">
        <Database class="w-4 h-4 text-slate-400" />
        <span>SQLite Database</span>
      </div>
      {#if dbConnected}
        <span class="font-mono text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1">
          <CheckCircle2 class="w-3.5 h-3.5" />
          <span>Connected</span>
        </span>
      {:else}
        <span class="font-mono text-rose-600 dark:text-rose-400 font-semibold flex items-center gap-1">
          <XCircle class="w-3.5 h-3.5" />
          <span>Disconnected</span>
        </span>
      {/if}
    </div>

    <div class="p-3 rounded-md bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 flex items-center justify-between">
      <span class="text-slate-700 dark:text-slate-300 font-medium">Caddy Connectivity</span>
      <span class="font-mono font-semibold text-slate-900 dark:text-slate-100">
        {health?.caddy_status || 'Unknown'}
      </span>
    </div>
  </div>
</div>
