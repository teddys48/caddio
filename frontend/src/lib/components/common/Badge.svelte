<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'connected' | 'disconnected' | 'healthy' | 'unhealthy' | 'neutral' | 'info';
    children?: Snippet;
    class?: string;
  }

  let { variant = 'neutral', children, class: className = '' }: Props = $props();

  const variantStyles = {
    connected: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
    healthy: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
    disconnected: 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20',
    unhealthy: 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20',
    neutral: 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20',
    info: 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20',
  };
</script>

<span
  class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-md border font-mono tracking-tight {variantStyles[variant]} {className}"
>
  {@render children?.()}
</span>
