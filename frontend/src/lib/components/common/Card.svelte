<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title?: string;
    description?: string;
    action?: Snippet;
    children?: Snippet;
    class?: string;
  }

  let { title, description, action, children, class: className = '' }: Props = $props();
</script>

<div
  class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg p-5 shadow-xs transition-colors {className}"
>
  {#if title || action}
    <div class="flex items-center justify-between gap-4 mb-4">
      <div>
        {#if title}
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 tracking-tight">
            {title}
          </h3>
        {/if}
        {#if description}
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            {description}
          </p>
        {/if}
      </div>
      {#if action}
        <div>
          {@render action()}
        </div>
      {/if}
    </div>
  {/if}

  <div>
    {@render children?.()}
  </div>
</div>
