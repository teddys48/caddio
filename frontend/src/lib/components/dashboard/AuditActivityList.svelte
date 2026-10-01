<script lang="ts">
  import type { AuditLog } from '../../types';
  import Badge from '../common/Badge.svelte';
  import { ShieldCheck, Clock, User, Layers } from 'lucide-svelte';

  interface Props {
    logs?: AuditLog[];
    loading?: boolean;
  }

  let { logs = [], loading = false }: Props = $props();

  function formatDate(iso: string) {
    if (!iso) return 'N/A';
    try {
      const d = new Date(iso);
      return d.toLocaleString(undefined, {
        dateStyle: 'medium',
        timeStyle: 'short',
      });
    } catch {
      return iso;
    }
  }

  function getActionBadgeVariant(action: string) {
    switch (action) {
      case 'SETUP_COMPLETED':
        return 'info';
      case 'LOGIN':
        return 'healthy';
      case 'LOGOUT':
        return 'neutral';
      default:
        return 'neutral';
    }
  }
</script>

<div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-5 shadow-xs transition-colors">
  <div class="flex items-center justify-between gap-4 mb-4">
    <div class="flex items-center gap-2.5">
      <div class="p-2 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
        <ShieldCheck class="w-5 h-5" />
      </div>
      <div>
        <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100 tracking-tight">
          Recent Activity
        </h3>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
          Audit event stream
        </p>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="space-y-2 py-4">
      <div class="h-10 bg-slate-100 dark:bg-slate-800 rounded-md animate-pulse"></div>
      <div class="h-10 bg-slate-100 dark:bg-slate-800 rounded-md animate-pulse"></div>
      <div class="h-10 bg-slate-100 dark:bg-slate-800 rounded-md animate-pulse"></div>
    </div>
  {:else if logs.length === 0}
    <div class="p-8 text-center border border-dashed border-slate-200 dark:border-slate-800 rounded-md my-2">
      <Clock class="w-8 h-8 text-slate-400 mx-auto mb-2 opacity-60" />
      <p class="text-sm font-medium text-slate-600 dark:text-slate-400">No recent activity</p>
      <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Audit log events will appear here when actions occur.</p>
    </div>
  {:else}
    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead>
          <tr class="border-b border-slate-200 dark:border-slate-800 text-slate-400 font-mono uppercase text-[10px] tracking-wider">
            <th class="py-2.5 px-3">Action</th>
            <th class="py-2.5 px-3">User</th>
            <th class="py-2.5 px-3">Resource</th>
            <th class="py-2.5 px-3 text-right">Timestamp</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono">
          {#each logs as log (log.id)}
            <tr class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
              <td class="py-3 px-3">
                <Badge variant={getActionBadgeVariant(log.action)}>
                  {log.action}
                </Badge>
              </td>
              <td class="py-3 px-3 text-slate-700 dark:text-slate-300">
                <div class="flex items-center gap-1.5 font-sans">
                  <User class="w-3.5 h-3.5 text-slate-400" />
                  <span>{log.username || 'System'}</span>
                </div>
              </td>
              <td class="py-3 px-3 text-slate-600 dark:text-slate-400">
                <div class="flex items-center gap-1.5 font-sans">
                  <Layers class="w-3.5 h-3.5 text-slate-400" />
                  <span>{log.resource}{log.resource_id ? ` (${log.resource_id})` : ''}</span>
                </div>
              </td>
              <td class="py-3 px-3 text-right text-slate-500 dark:text-slate-400">
                {formatDate(log.created_at)}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
