<script lang="ts">
  import type { CaddyStatus } from '../../types';
  import Badge from '../common/Badge.svelte';
  import Button from '../common/Button.svelte';
  import { Server, RefreshCw, CheckCircle2, AlertTriangle, HelpCircle } from 'lucide-svelte';

  interface Props {
    status?: CaddyStatus;
    loading?: boolean;
    onRetry?: () => void;
  }

  let { status, loading = false, onRetry }: Props = $props();

  const isConnected = $derived(status?.status === 'Connected');
  const isDisconnected = $derived(status?.status === 'Disconnected');
</script>

<div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-5 shadow-xs transition-colors">
  <div class="flex items-center justify-between gap-4 mb-4">
    <div class="flex items-center gap-2.5">
      <div class="p-2 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
        <Server class="w-5 h-5" />
      </div>
      <div>
        <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100 tracking-tight">
          Caddy Server Status
        </h3>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
          Caddy Admin API connectivity
        </p>
      </div>
    </div>

    {#if isConnected}
      <Badge variant="connected">
        <CheckCircle2 class="w-3.5 h-3.5" />
        <span>Connected</span>
      </Badge>
    {:else if isDisconnected}
      <Badge variant="disconnected">
        <AlertTriangle class="w-3.5 h-3.5" />
        <span>Disconnected</span>
      </Badge>
    {:else}
      <Badge variant="neutral">
        <HelpCircle class="w-3.5 h-3.5" />
        <span>Unknown</span>
      </Badge>
    {/if}
  </div>

  <div class="space-y-3">
    <div class="p-3 rounded-md bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs">
      <span class="text-slate-500 dark:text-slate-400 font-medium">Caddy Version</span>
      <span class="font-mono font-semibold text-slate-900 dark:text-slate-100">
        {status?.version && status.version !== '' ? status.version : 'Unknown'}
      </span>
    </div>

    {#if isDisconnected}
      <div class="p-3.5 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-700 dark:text-rose-300 space-y-2">
        <div class="font-semibold flex items-center gap-1.5">
          <AlertTriangle class="w-4 h-4 shrink-0" />
          <span>Caddy Unavailable</span>
        </div>
        <p class="leading-relaxed">
          Unable to connect to the Caddy Admin API. Check that Caddy is running and that the configured Admin API address is reachable over the internal network.
        </p>
        <div class="pt-1">
          <Button variant="danger" size="sm" onclick={onRetry} {loading}>
            <RefreshCw class="w-3.5 h-3.5" />
            <span>Retry Connection</span>
          </Button>
        </div>
      </div>
    {/if}
  </div>
</div>
