<script lang="ts">
  import { useAuth } from '../../stores/auth.svelte';
  import { useRouter } from '../../router/router.svelte';
  import { api } from '../../api/client';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import Button from '../common/Button.svelte';
  import { LogOut, User as UserIcon, Server, Menu } from 'lucide-svelte';

  interface Props {
    onToggleMobileSidebar?: () => void;
  }

  let { onToggleMobileSidebar }: Props = $props();

  const auth = useAuth();
  const router = useRouter();
  let loggingOut = $state(false);

  async function handleLogout() {
    loggingOut = true;
    try {
      await api.logout();
      auth.clear();
      router.navigate('/login');
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      loggingOut = false;
    }
  }
</script>

<header class="h-14 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 px-4 md:px-6 flex items-center justify-between sticky top-0 z-30">
  <div class="flex items-center gap-3">
    <button
      type="button"
      onclick={onToggleMobileSidebar}
      class="md:hidden p-1.5 rounded-md text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100"
      aria-label="Toggle navigation menu"
    >
      <Menu class="w-5 h-5" />
    </button>

    <div class="flex items-center gap-2.5">
      <div class="w-7 h-7 rounded-md bg-slate-900 dark:bg-slate-100 flex items-center justify-center text-white dark:text-slate-900">
        <Server class="w-4 h-4" />
      </div>
      <div>
        <span class="font-bold text-sm tracking-tight text-slate-900 dark:text-slate-100">Caddio</span>
        <span class="ml-2 text-[10px] uppercase font-mono tracking-widest text-slate-400 border border-slate-200 dark:border-slate-800 px-1.5 py-0.5 rounded-sm">Phase 1</span>
      </div>
    </div>
  </div>

  <div class="flex items-center gap-3">
    <ThemeToggle />

    {#if auth.authenticated && auth.user}
      <div class="hidden sm:flex items-center gap-2 pl-3 border-l border-slate-200 dark:border-slate-800 text-xs font-mono text-slate-600 dark:text-slate-300">
        <UserIcon class="w-3.5 h-3.5 text-slate-400" />
        <span>{auth.user.username}</span>
      </div>

      <Button
        variant="ghost"
        size="sm"
        onclick={handleLogout}
        loading={loggingOut}
        title="Log out"
        class="text-slate-600 dark:text-slate-400 hover:text-rose-600 dark:hover:text-rose-400"
      >
        <LogOut class="w-4 h-4" />
        <span class="hidden sm:inline">Log out</span>
      </Button>
    {/if}
  </div>
</header>
