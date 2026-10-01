<script lang="ts">
  import type { Site } from '../../types';
  import Badge from '../common/Badge.svelte';
  import Button from '../common/Button.svelte';
  import {
    Search,
    Globe,
    Server,
    ShieldCheck,
    Lock,
    Pencil,
    Trash2,
    CheckCircle2,
    ExternalLink
  } from 'lucide-svelte';

  interface Props {
    sites: Site[];
    unmanagedCount: number;
    loading?: boolean;
    onEdit: (site: Site) => void;
    onDelete: (site: Site) => void;
  }

  let { sites = [], unmanagedCount = 0, loading = false, onEdit, onDelete }: Props = $props();

  let searchQuery = $state('');
  let filterType = $state<'all' | 'managed' | 'external'>('all');

  const filteredSites = $derived(
    sites.filter((site) => {
      const matchesSearch =
        searchQuery === '' ||
        site.domain.toLowerCase().includes(searchQuery.toLowerCase()) ||
        site.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        site.upstreams.some((u) => u.address.toLowerCase().includes(searchQuery.toLowerCase()));

      if (!matchesSearch) return false;

      if (filterType === 'managed') return site.is_managed;
      if (filterType === 'external') return !site.is_managed;
      return true;
    })
  );

  const managedCount = $derived(sites.filter((s) => s.is_managed).length);
  const externalCount = $derived(sites.filter((s) => !s.is_managed).length);
</script>

<div class="space-y-4">
  <!-- Filters & Controls Toolbar -->
  <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 bg-white dark:bg-slate-900 p-4 rounded-lg border border-slate-200 dark:border-slate-800 shadow-xs">
    <!-- Search Bar -->
    <div class="relative flex-1 max-w-md">
      <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Filter by domain or upstream address..."
        class="w-full pl-9 pr-4 py-2 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md text-xs font-mono text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400 dark:focus:ring-slate-600 transition-colors"
      />
    </div>

    <!-- Filter Pills -->
    <div class="flex items-center gap-1.5 self-start sm:self-auto">
      <button
        onclick={() => (filterType = 'all')}
        class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors flex items-center gap-1.5 cursor-pointer {filterType === 'all' ? 'bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700'}"
      >
        <span>All</span>
        <span class="px-1.5 py-0.2 rounded-full text-[10px] bg-slate-700/20 dark:bg-slate-300/20 font-mono">
          {sites.length}
        </span>
      </button>

      <button
        onclick={() => (filterType = 'managed')}
        class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors flex items-center gap-1.5 cursor-pointer {filterType === 'managed' ? 'bg-emerald-600 text-white' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700'}"
      >
        <CheckCircle2 class="w-3.5 h-3.5" />
        <span>Managed</span>
        <span class="px-1.5 py-0.2 rounded-full text-[10px] bg-emerald-700/20 font-mono">
          {managedCount}
        </span>
      </button>

      <button
        onclick={() => (filterType = 'external')}
        class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors flex items-center gap-1.5 cursor-pointer {filterType === 'external' ? 'bg-slate-700 text-white dark:bg-slate-300 dark:text-slate-900' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700'}"
      >
        <Lock class="w-3.5 h-3.5" />
        <span>External</span>
        <span class="px-1.5 py-0.2 rounded-full text-[10px] bg-slate-700/20 font-mono">
          {externalCount}
        </span>
      </button>
    </div>
  </div>

  <!-- Unmanaged Routes Context Banner -->
  {#if externalCount > 0 && filterType !== 'managed'}
    <div class="p-3.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-800 dark:text-amber-300 flex items-start gap-2.5">
      <Lock class="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
      <div>
        <span class="font-semibold">Unmanaged Caddy Routes Detected:</span>
        Caddio discovered {externalCount} external route(s) in Caddy's configuration. External routes are displayed as read-only. Caddio only edits or deletes routes created within Caddio.
      </div>
    </div>
  {/if}

  <!-- Main Sites Table -->
  <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg overflow-hidden shadow-xs">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse text-xs">
        <thead>
          <tr class="bg-slate-50 dark:bg-slate-950/70 border-b border-slate-200 dark:border-slate-800 text-slate-500 dark:text-slate-400 font-medium select-none">
            <th class="py-3 px-4">Domain / Name</th>
            <th class="py-3 px-4">Upstream Target(s)</th>
            <th class="py-3 px-4">Ownership</th>
            <th class="py-3 px-4">Security</th>
            <th class="py-3 px-4 text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
          {#if loading}
            {#each Array(3) as _}
              <tr class="animate-pulse">
                <td class="py-4 px-4"><div class="h-4 bg-slate-200 dark:bg-slate-800 rounded w-40"></div></td>
                <td class="py-4 px-4"><div class="h-4 bg-slate-200 dark:bg-slate-800 rounded w-48"></div></td>
                <td class="py-4 px-4"><div class="h-4 bg-slate-200 dark:bg-slate-800 rounded w-24"></div></td>
                <td class="py-4 px-4"><div class="h-4 bg-slate-200 dark:bg-slate-800 rounded w-16"></div></td>
                <td class="py-4 px-4"><div class="h-4 bg-slate-200 dark:bg-slate-800 rounded w-16 ml-auto"></div></td>
              </tr>
            {/each}
          {:else if filteredSites.length === 0}
            <tr>
              <td colspan="5" class="py-12 px-4 text-center">
                <div class="max-w-xs mx-auto text-center space-y-2">
                  <div class="w-10 h-10 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-400 flex items-center justify-center mx-auto">
                    <Globe class="w-5 h-5" />
                  </div>
                  <div class="text-sm font-semibold text-slate-800 dark:text-slate-200">No sites found</div>
                  <p class="text-xs text-slate-500 dark:text-slate-400">
                    {searchQuery ? 'No reverse proxy routes match your search filter.' : 'No reverse proxy sites configured in Caddy yet.'}
                  </p>
                </div>
              </td>
            </tr>
          {:else}
            {#each filteredSites as site (site.id)}
              <tr class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors">
                <!-- Domain / Name -->
                <td class="py-3 px-4">
                  <div class="flex items-center gap-2">
                    <Globe class="w-4 h-4 text-slate-400 shrink-0" />
                    <div>
                      <a
                        href="https://{site.domain}"
                        target="_blank"
                        rel="noreferrer"
                        class="font-mono text-xs font-semibold text-slate-900 dark:text-slate-100 hover:text-emerald-600 dark:hover:text-emerald-400 flex items-center gap-1 group"
                      >
                        <span>{site.domain}</span>
                        <ExternalLink class="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity text-slate-400" />
                      </a>
                      {#if site.name && site.name !== site.domain}
                        <div class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">{site.name}</div>
                      {/if}
                    </div>
                  </div>
                </td>

                <!-- Upstream Targets -->
                <td class="py-3 px-4">
                  <div class="flex flex-wrap gap-1 items-center">
                    {#each site.upstreams as upstream}
                      <span class="inline-flex items-center gap-1 font-mono text-[11px] bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 px-2 py-0.5 rounded border border-slate-200/60 dark:border-slate-700/60">
                        <Server class="w-3 h-3 text-slate-400" />
                        <span>{upstream.address}</span>
                      </span>
                    {/each}
                  </div>
                </td>

                <!-- Managed Status -->
                <td class="py-3 px-4">
                  {#if site.is_managed}
                    <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20">
                      <CheckCircle2 class="w-3 h-3" />
                      <span>Caddio Managed</span>
                    </span>
                  {:else}
                    <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700" title="Managed externally outside Caddio">
                      <Lock class="w-3 h-3 text-slate-400" />
                      <span>External (Read-Only)</span>
                    </span>
                  {/if}
                </td>

                <!-- Security / HTTPS -->
                <td class="py-3 px-4">
                  {#if site.https}
                    <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-500/20">
                      <ShieldCheck class="w-3 h-3" />
                      <span>Auto HTTPS</span>
                    </span>
                  {:else}
                    <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] bg-slate-100 text-slate-500">
                      <span>HTTP</span>
                    </span>
                  {/if}
                </td>

                <!-- Actions -->
                <td class="py-3 px-4 text-right">
                  {#if site.is_managed}
                    <div class="flex items-center justify-end gap-1.5">
                      <button
                        onclick={() => onEdit(site)}
                        class="p-1.5 rounded-md text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                        title="Edit Site Configuration"
                      >
                        <Pencil class="w-3.5 h-3.5" />
                      </button>
                      <button
                        onclick={() => onDelete(site)}
                        class="p-1.5 rounded-md text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition-colors cursor-pointer"
                        title="Delete Site"
                      >
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                    </div>
                  {:else}
                    <span class="inline-flex items-center gap-1 text-[11px] text-slate-400 italic" title="External routes cannot be modified by Caddio">
                      <Lock class="w-3 h-3" />
                      <span>Read-Only</span>
                    </span>
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</div>
