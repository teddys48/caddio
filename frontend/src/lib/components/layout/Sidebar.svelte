<script lang="ts">
  import { useRouter, type Route } from '../../router/router.svelte';
  import {
    LayoutDashboard,
    Globe,
    FileCode,
    ScrollText,
    History,
    ShieldCheck,
    Settings,
  } from 'lucide-svelte';

  interface Props {
    isOpen?: boolean;
    onClose?: () => void;
  }

  let { isOpen = false, onClose }: Props = $props();
  const router = useRouter();

  const navigationGroups = [
    {
      title: 'Overview',
      items: [
        {
          name: 'Dashboard',
          route: '/dashboard' as Route,
          icon: LayoutDashboard,
          active: true,
        },
      ],
    },
    {
      title: 'Management',
      items: [
        {
          name: 'Sites',
          route: '/sites' as Route,
          icon: Globe,
          active: true,
        },
        {
          name: 'Configuration',
          route: '/config' as Route,
          icon: FileCode,
          active: true,
        },
      ],
    },
    {
      title: 'Observability',
      items: [
        {
          name: 'Runtime & Logs',
          route: '/runtime' as Route,
          icon: ScrollText,
          active: true,
        },
        {
          name: 'History',
          route: '/config/history' as Route,
          icon: History,
          active: true,
        },
      ],
    },
    {
      title: 'System',
      items: [
        {
          name: 'Audit Log',
          route: '/audit' as Route,
          icon: ShieldCheck,
          active: true,
        },
        {
          name: 'Settings',
          route: '/settings' as Route,
          icon: Settings,
          active: false,
        },
      ],
    },
  ];

  function handleNavigate(route: Route) {
    router.navigate(route);
    if (onClose) onClose();
  }
</script>

<!-- Mobile backdrop -->
{#if isOpen}
  <button
    type="button"
    onclick={onClose}
    class="fixed inset-0 bg-slate-900/50 backdrop-blur-xs z-40 md:hidden border-none cursor-pointer"
    aria-label="Close sidebar backdrop"
  ></button>
{/if}

<aside
  class="fixed md:sticky top-14 left-0 z-40 w-64 h-[calc(100vh-3.5rem)] bg-white text-slate-700 border-r border-slate-200 dark:bg-slate-900 dark:text-slate-300 dark:border-slate-800 transition-transform duration-200 ease-in-out shrink-0 {isOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}"
>
  <div class="p-4 flex flex-col gap-6 h-full overflow-y-auto">
    {#each navigationGroups as group}
      <div>
        <div class="px-2 mb-2 text-[10px] uppercase font-mono tracking-widest text-slate-400 dark:text-slate-500 font-semibold">
          {group.title}
        </div>
        <nav class="space-y-1">
          {#each group.items as item}
            {@const Icon = item.icon}
            {@const isSelected = router.route === item.route}
            <button
              type="button"
              onclick={() => handleNavigate(item.route)}
              class="w-full flex items-center justify-between px-2.5 py-2 text-xs font-medium rounded-md transition-colors cursor-pointer {isSelected ? 'bg-blue-600 text-white font-semibold shadow-xs' : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-800 dark:hover:text-white'}"
            >
              <div class="flex items-center gap-2.5">
                <Icon class="w-4 h-4 shrink-0" />
                <span>{item.name}</span>
              </div>

              {#if !item.active}
                <span class="text-[9px] font-mono uppercase bg-slate-100 text-slate-500 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700 px-1.5 py-0.5 rounded-sm">
                  Soon
                </span>
              {/if}
            </button>
          {/each}
        </nav>
      </div>
    {/each}

    <div class="mt-auto pt-4 border-t border-slate-200 dark:border-slate-800 text-[11px] font-mono text-slate-400 dark:text-slate-500 text-center">
      Caddio v1.0.0
    </div>
  </div>
</aside>
