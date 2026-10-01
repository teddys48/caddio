<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EditorView, basicSetup } from 'codemirror';
  import { json } from '@codemirror/lang-json';
  import { EditorState } from '@codemirror/state';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { Check, Copy, Code, AlertTriangle } from 'lucide-svelte';

  interface Props {
    value: string;
    readonly?: boolean;
    onChange?: (val: string) => void;
  }

  let { value = $bindable(''), readonly = false, onChange }: Props = $props();

  let container: HTMLDivElement;
  let view: EditorView | null = null;
  let copied = $state(false);
  let jsonError = $state<string | null>(null);

  $effect(() => {
    // Sync external value updates to CodeMirror if changed outside
    if (view && value !== view.state.doc.toString()) {
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: value }
      });
    }
  });

  onMount(() => {
    const updateListener = EditorView.updateListener.of((update) => {
      if (update.docChanged) {
        const docString = update.state.doc.toString();
        value = docString;
        validateSyntax(docString);
        if (onChange) {
          onChange(docString);
        }
      }
    });

    const theme = EditorView.theme({
      '&': {
        height: '100%',
        fontSize: '13px',
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace'
      }
    });

    const state = EditorState.create({
      doc: value,
      extensions: [
        basicSetup,
        json(),
        oneDark,
        theme,
        EditorState.readOnly.of(readonly),
        updateListener
      ]
    });

    view = new EditorView({
      state,
      parent: container
    });

    validateSyntax(value);
  });

  onDestroy(() => {
    if (view) {
      view.destroy();
    }
  });

  function validateSyntax(str: string) {
    if (!str.trim()) {
      jsonError = null;
      return;
    }
    try {
      JSON.parse(str);
      jsonError = null;
    } catch (e: any) {
      jsonError = e.message;
    }
  }

  function handleCopy() {
    navigator.clipboard.writeText(value);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }

  export function formatJSON() {
    try {
      const parsed = JSON.parse(value);
      const formatted = JSON.stringify(parsed, null, 2);
      value = formatted;
      if (view) {
        view.dispatch({
          changes: { from: 0, to: view.state.doc.length, insert: formatted }
        });
      }
      jsonError = null;
    } catch (e: any) {
      jsonError = `Cannot format invalid JSON: ${e.message}`;
    }
  }
</script>

<div class="flex flex-col h-full border border-slate-200 dark:border-slate-800 rounded-lg overflow-hidden bg-white dark:bg-[#090d16] shadow-xs">
  <!-- Editor Secondary Header -->
  <div class="flex items-center justify-between px-4 py-2 bg-slate-50 dark:bg-slate-900/90 border-b border-slate-200 dark:border-slate-800 text-xs">
    <div class="flex items-center space-x-2 font-mono text-slate-600 dark:text-slate-400">
      <Code class="w-3.5 h-3.5 text-sky-600 dark:text-sky-400" />
      <span>caddy.json</span>
      {#if jsonError}
        <span class="inline-flex items-center space-x-1 px-2 py-0.5 rounded bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-500/20">
          <AlertTriangle class="w-3 h-3" />
          <span>JSON Syntax Error</span>
        </span>
      {:else if value.trim()}
        <span class="inline-flex items-center space-x-1 px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20">
          <span>Valid JSON Syntax</span>
        </span>
      {/if}
    </div>
    <div class="flex items-center space-x-2">
      {#if !readonly}
        <button
          type="button"
          onclick={formatJSON}
          class="px-2 py-1 text-xs rounded bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 font-mono transition-colors cursor-pointer"
          title="Format JSON with 2 space indentation"
        >
          Format JSON
        </button>
      {/if}
      <button
        type="button"
        onclick={handleCopy}
        class="flex items-center space-x-1 px-2 py-1 text-xs rounded bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 font-mono transition-colors cursor-pointer"
      >
        {#if copied}
          <Check class="w-3 h-3 text-emerald-600 dark:text-emerald-400" />
          <span>Copied</span>
        {:else}
          <Copy class="w-3 h-3 text-slate-400 dark:text-slate-500" />
          <span>Copy</span>
        {/if}
      </button>
    </div>
  </div>

  <!-- CodeMirror Container -->
  <div bind:this={container} class="flex-1 overflow-auto min-h-[400px]"></div>

  {#if jsonError}
    <div class="px-4 py-2 bg-rose-50 dark:bg-rose-950/40 border-t border-rose-200 dark:border-rose-900/50 text-xs font-mono text-rose-700 dark:text-rose-300 flex items-start space-x-2">
      <AlertTriangle class="w-4 h-4 text-rose-600 dark:text-rose-400 shrink-0 mt-0.5" />
      <div class="break-all">{jsonError}</div>
    </div>
  {/if}
</div>
