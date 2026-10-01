<script lang="ts">
  import { useTheme, type Theme } from '../../stores/theme.svelte';
  import { Sun, Moon, Monitor } from 'lucide-svelte';

  const themeStore = useTheme();

  function selectTheme(t: Theme) {
    themeStore.setTheme(t);
  }
</script>

<div class="inline-flex items-center p-1 rounded-lg bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-xs font-medium">
  <button
    type="button"
    onclick={() => selectTheme('light')}
    title="Light theme"
    aria-label="Light theme"
    class="flex items-center justify-center p-1.5 rounded-md transition-colors cursor-pointer {themeStore.theme === 'light' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs' : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100'}"
  >
    <Sun class="w-3.5 h-3.5" />
  </button>
  <button
    type="button"
    onclick={() => selectTheme('dark')}
    title="Dark theme"
    aria-label="Dark theme"
    class="flex items-center justify-center p-1.5 rounded-md transition-colors cursor-pointer {themeStore.theme === 'dark' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs' : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100'}"
  >
    <Moon class="w-3.5 h-3.5" />
  </button>
  <button
    type="button"
    onclick={() => selectTheme('system')}
    title="System theme"
    aria-label="System theme"
    class="flex items-center justify-center p-1.5 rounded-md transition-colors cursor-pointer {themeStore.theme === 'system' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs' : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100'}"
  >
    <Monitor class="w-3.5 h-3.5" />
  </button>
</div>
