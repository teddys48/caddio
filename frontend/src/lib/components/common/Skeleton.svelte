<script lang="ts">
  interface Props {
    class?: string;
  }
  let { class: className = '' }: Props = $props();
</script>

<div
  class="animate-pulse bg-slate-200 dark:bg-slate-800 rounded-md {className}"
></div>
