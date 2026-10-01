<script lang="ts">
  import { onMount } from 'svelte';
  import type { Site, SiteDiscoveryResult } from '../lib/types';
  import { api } from '../lib/api/client';
  import Button from '../lib/components/common/Button.svelte';
  import SiteListTable from '../lib/components/sites/SiteListTable.svelte';
  import SiteFormModal from '../lib/components/sites/SiteFormModal.svelte';
  import DeleteSiteDialog from '../lib/components/sites/DeleteSiteDialog.svelte';
  import {
    Globe,
    Plus,
    RefreshCw,
    AlertTriangle,
    CheckCircle2,
    Lock,
    Layers
  } from 'lucide-svelte';

  let sites = $state<Site[]>([]);
  let unmanagedCount = $state(0);
  let loading = $state(true);
  let error = $state('');

  // Modal states
  let isFormOpen = $state(false);
  let selectedSite = $state<Site | null>(null);
  let isDeleteOpen = $state(false);
  let siteToDelete = $state<Site | null>(null);

  async function loadSites() {
    loading = true;
    error = '';
    try {
      const res: SiteDiscoveryResult = await api.getSites();
      sites = res.sites || [];
      unmanagedCount = res.unmanaged_count || 0;
    } catch (err: any) {
      error = err.message || 'Failed to load sites from Caddy server';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadSites();
  });

  function handleOpenCreate() {
    selectedSite = null;
    isFormOpen = true;
  }

  function handleOpenEdit(site: Site) {
    selectedSite = site;
    isFormOpen = true;
  }

  function handleOpenDelete(site: Site) {
    siteToDelete = site;
    isDeleteOpen = true;
  }
</script>

<div class="space-y-6">
  <!-- Page Header -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight flex items-center gap-2">
        <Globe class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        <span>Reverse Proxy Sites</span>
      </h1>
      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
        Discover, configure, and manage Caddy reverse proxy domain routes and backend upstreams
      </p>
    </div>

    <div class="flex items-center gap-2.5">
      <Button variant="secondary" size="sm" onclick={loadSites} {loading}>
        <RefreshCw class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" />
        <span>Refresh</span>
      </Button>

      <Button variant="primary" size="sm" onclick={handleOpenCreate}>
        <Plus class="w-4 h-4" />
        <span>Add Site</span>
      </Button>
    </div>
  </div>

  <!-- Operational Overview Metric Cards -->
  <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-4 shadow-xs flex items-center justify-between">
      <div>
        <div class="text-xs text-slate-500 dark:text-slate-400 font-medium">Total Configured Routes</div>
        <div class="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono mt-1">
          {sites.length}
        </div>
      </div>
      <div class="p-2.5 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
        <Layers class="w-5 h-5" />
      </div>
    </div>

    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-4 shadow-xs flex items-center justify-between">
      <div>
        <div class="text-xs text-slate-500 dark:text-slate-400 font-medium">Caddio Managed</div>
        <div class="text-2xl font-bold text-emerald-600 dark:text-emerald-400 font-mono mt-1">
          {sites.filter((s) => s.is_managed).length}
        </div>
      </div>
      <div class="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
        <CheckCircle2 class="w-5 h-5" />
      </div>
    </div>

    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-4 shadow-xs flex items-center justify-between">
      <div>
        <div class="text-xs text-slate-500 dark:text-slate-400 font-medium">External (Read-Only)</div>
        <div class="text-2xl font-bold text-slate-600 dark:text-slate-400 font-mono mt-1">
          {sites.filter((s) => !s.is_managed).length}
        </div>
      </div>
      <div class="p-2.5 rounded-lg bg-slate-500/10 text-slate-600 dark:text-slate-400">
        <Lock class="w-5 h-5" />
      </div>
    </div>
  </div>

  <!-- Global Error Banner -->
  {#if error}
    <div class="p-4 rounded-lg bg-rose-500/10 border border-rose-500/20 text-xs text-rose-700 dark:text-rose-300 flex items-start gap-3">
      <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
      <div class="space-y-1">
        <div class="font-semibold">Unable to fetch Caddy configuration</div>
        <p class="leading-relaxed">{error}</p>
        <Button variant="danger" size="sm" class="mt-2" onclick={loadSites}>
          Retry Connection
        </Button>
      </div>
    </div>
  {/if}

  <!-- Main Sites Table Component -->
  <SiteListTable
    {sites}
    {unmanagedCount}
    {loading}
    onEdit={handleOpenEdit}
    onDelete={handleOpenDelete}
  />

  <!-- Create / Edit Site Modal -->
  <SiteFormModal
    open={isFormOpen}
    site={selectedSite}
    onClose={() => (isFormOpen = false)}
    onSuccess={loadSites}
  />

  <!-- Delete Site Dialog -->
  <DeleteSiteDialog
    open={isDeleteOpen}
    site={siteToDelete}
    onClose={() => (isDeleteOpen = false)}
    onSuccess={loadSites}
  />
</div>
