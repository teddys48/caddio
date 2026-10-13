<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';

  export let loading: boolean = false;

  const dispatch = createEventDispatcher<{ refresh: void }>();

  let autoRefresh: boolean = false;
  let timer: ReturnType<typeof setInterval> | null = null;
  let secondsRemaining: number = 30;

  function triggerRefresh() {
    secondsRemaining = 30;
    dispatch('refresh');
  }

  function toggleAutoRefresh() {
    autoRefresh = !autoRefresh;
    if (autoRefresh) {
      startTimer();
    } else {
      stopTimer();
    }
  }

  function startTimer() {
    stopTimer();
    secondsRemaining = 30;
    timer = setInterval(() => {
      secondsRemaining -= 1;
      if (secondsRemaining <= 0) {
        secondsRemaining = 30;
        dispatch('refresh');
      }
    }, 1000);
  }

  function stopTimer() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  onDestroy(() => {
    stopTimer();
  });
</script>

<div class="flex items-center space-x-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-1.5 shadow-sm">
  <!-- Manual Refresh Button -->
  <button
    type="button"
    on:click={triggerRefresh}
    disabled={loading}
    class="px-2.5 py-1 text-xs font-medium rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center space-x-1.5 disabled:opacity-50"
  >
    <svg class="w-3.5 h-3.5 {loading ? 'animate-spin text-indigo-500' : 'text-slate-500'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
    </svg>
    <span>{loading ? 'Refreshing...' : 'Refresh'}</span>
  </button>

  <div class="h-4 w-[1px] bg-slate-200 dark:border-slate-800"></div>

  <!-- Auto Refresh Toggle -->
  <button
    type="button"
    on:click={toggleAutoRefresh}
    class="flex items-center space-x-2 text-xs text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white transition-colors"
  >
    <span
      class="w-7 h-4 flex items-center rounded-full p-0.5 transition-colors {autoRefresh ? 'bg-indigo-600' : 'bg-slate-300 dark:bg-slate-700'}"
    >
      <span
        class="bg-white w-3 h-3 rounded-full shadow-md transform transition-transform {autoRefresh ? 'translate-x-3' : 'translate-x-0'}"
      ></span>
    </span>
    <span class="font-medium select-none">
      Auto-refresh (30s)
      {#if autoRefresh}
        <span class="text-indigo-600 dark:text-indigo-400 font-mono text-[10px] ml-1">({secondsRemaining}s)</span>
      {/if}
    </span>
  </button>
</div>
