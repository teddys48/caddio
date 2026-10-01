<script lang="ts">
  interface Props {
    id: string;
    label: string;
    type?: string;
    value?: string;
    placeholder?: string;
    required?: boolean;
    disabled?: boolean;
    error?: string;
    helperText?: string;
    oninput?: (e: Event) => void;
    class?: string;
  }

  let {
    id,
    label,
    type = 'text',
    value = $bindable(''),
    placeholder = '',
    required = false,
    disabled = false,
    error,
    helperText,
    oninput,
    class: className = '',
  }: Props = $props();
</script>

<div class="flex flex-col gap-1.5 w-full">
  <label for={id} class="text-xs font-semibold text-slate-700 dark:text-slate-300">
    {label}
    {#if required}
      <span class="text-rose-500 ml-0.5">*</span>
    {/if}
  </label>

  <input
    {id}
    {type}
    bind:value
    {placeholder}
    {required}
    {disabled}
    {oninput}
    class="w-full px-3 py-2 text-sm rounded-md bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 border border-slate-300 dark:border-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors placeholder:text-slate-400 dark:placeholder:text-slate-600 disabled:opacity-50 disabled:cursor-not-allowed {error ? 'border-rose-500 focus:ring-rose-500' : ''} {className}"
  />

  {#if error}
    <p class="text-xs text-rose-500 font-medium mt-0.5">{error}</p>
  {:else if helperText}
    <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{helperText}</p>
  {/if}
</div>
