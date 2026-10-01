<script lang="ts">
  import type { Snippet } from 'svelte';
  import Header from './Header.svelte';
  import Sidebar from './Sidebar.svelte';
  import { useAuth } from '../../stores/auth.svelte';

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();
  const auth = useAuth();
  let mobileSidebarOpen = $state(false);

  function toggleMobileSidebar() {
    mobileSidebarOpen = !mobileSidebarOpen;
  }
</script>

{#if auth.authenticated}
  <div class="min-h-screen flex flex-col bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans">
    <Header onToggleMobileSidebar={toggleMobileSidebar} />
    <div class="flex-1 flex w-full">
      <Sidebar isOpen={mobileSidebarOpen} onClose={() => (mobileSidebarOpen = false)} />
      <main class="flex-1 p-4 md:p-8 max-w-7xl mx-auto w-full overflow-x-hidden">
        {@render children?.()}
      </main>
    </div>
  </div>
{:else}
  <div class="min-h-screen flex flex-col bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans">
    <div class="p-4 flex justify-end max-w-7xl mx-auto w-full">
      <Header />
    </div>
    <main class="flex-1 flex items-center justify-center p-4">
      {@render children?.()}
    </main>
  </div>
{/if}
