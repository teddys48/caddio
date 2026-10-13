<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { RuntimeLogFilters } from '../../types';

  export let filters: RuntimeLogFilters = {
    level: 'all',
    search: '',
    limit: 100
  };

  const dispatch = createEventDispatcher<{ change: RuntimeLogFilters }>();

  const levels = [
    { id: 'all', label: 'All Levels' },
    { id: 'error', label: 'Error', color: 'text-rose-600 dark:text-rose-400' },
    { id: 'warn', label: 'Warn', color: 'text-amber-600 dark:text-amber-400' },
    { id: 'info', label: 'Info', color: 'text-blue-600 dark:text-blue-400' },
    { id: 'debug', label: 'Debug', color: 'text-slate-600 dark:text-slate-400' }
  ];

  const limits = [50, 100, 250, 500];

  function setLevel(lvl: string) {
    filters.level = lvl;
    dispatchChange();
  }

  function handleSearchInput(e: Event) {
    const target = e.target as HTMLInputElement;
    filters.search = target.value;
    dispatchChange();
  }

  function handleLimitChange(e: Event) {
    const target = e.target as HTMLSelectElement;
    filters.limit = parseInt(target.value, 10);
    dispatchChange();
  }

  function clearFilters() {
    filters = {
      level: 'all',
      search: '',
      limit: 100
    };
    dispatchChange();
  }

  function dispatchChange() {
    dispatch('change', { ...filters });
  }
</script>

<div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-4 shadow-sm space-y-3">
  <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
    <!-- Level Selector Pills -->
    <div class="flex flex-wrap items-center gap-1.5">
      <span class="text-xs font-medium text-slate-500 dark:text-slate-400 mr-1 hidden sm:inline">Level:</span>
      {#each levels as lvl}
        <button
          type="button"
          on:click={() => setLevel(lvl.id)}
          class="px-2.5 py-1 text-xs font-medium rounded-lg transition-colors border {filters.level === lvl.id
            ? 'bg-slate-900 text-white dark:bg-indigo-600 dark:text-white border-transparent shadow-xs'
            : 'bg-slate-50 dark:bg-slate-800/80 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700/70 hover:bg-slate-100 dark:hover:bg-slate-800'}"
        >
          {lvl.label}
        </button>
      {/each}
    </div>

    <!-- Right Controls: Limit & Clear -->
    <div class="flex items-center space-x-2 shrink-0">
      <div class="flex items-center space-x-1.5">
        <label for="log-limit-select" class="text-xs text-slate-500 dark:text-slate-400 font-medium">Rows:</label>
        <select
          id="log-limit-select"
          value={filters.limit || 100}
          on:change={handleLimitChange}
          class="px-2 py-1 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          {#each limits as lim}
            <option value={lim}>{lim}</option>
          {/each}
        </select>
      </div>

      {#if filters.level !== 'all' || filters.search !== '' || filters.limit !== 100}
        <button
          type="button"
          on:click={clearFilters}
          class="px-2.5 py-1 text-xs text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200 font-medium underline transition-colors"
        >
          Reset
        </button>
      {/if}
    </div>
  </div>

  <!-- Search Bar -->
  <div class="relative">
    <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    </div>
    <input
      type="text"
      placeholder="Filter logs by message, logger, host, URI, status..."
      value={filters.search || ''}
      on:input={handleSearchInput}
      class="w-full pl-9 pr-8 py-2 text-xs sm:text-sm rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/50 text-slate-800 dark:text-slate-200 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500"
    />
    {#if filters.search}
      <button
        type="button"
        on:click={() => { filters.search = ''; dispatchChange(); }}
        aria-label="Clear search"
        class="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>
</div>
