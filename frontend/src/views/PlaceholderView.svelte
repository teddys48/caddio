<script lang="ts">
  import Card from '../lib/components/common/Card.svelte';
  import Badge from '../lib/components/common/Badge.svelte';
  import { Construction, Sparkles } from 'lucide-svelte';

  interface Props {
    title: string;
    description: string;
    plannedPhase?: string;
  }

  let { title, description, plannedPhase = 'Phase 2' }: Props = $props();
</script>

<div class="space-y-6">
  <div class="pb-4 border-b border-slate-200 dark:border-slate-800">
    <div class="flex items-center gap-3">
      <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
        {title}
      </h1>
      <Badge variant="info">
        <Sparkles class="w-3 h-3" />
        <span>{plannedPhase}</span>
      </Badge>
    </div>
    <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
      {description}
    </p>
  </div>

  <Card class="p-12 text-center">
    <div class="w-12 h-12 rounded-xl bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 mx-auto flex items-center justify-center mb-4">
      <Construction class="w-6 h-6" />
    </div>

    <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100">
      Feature Planned for {plannedPhase}
    </h3>

    <p class="text-xs text-slate-500 dark:text-slate-400 max-w-md mx-auto mt-2 leading-relaxed">
      This management surface is intentionally deferred to subsequent releases as part of the Caddio roadmap. Phase 1 focuses on core foundation, local security, SQLite migrations, and operational server health monitoring.
    </p>

    <div class="mt-6">
      <span class="inline-flex items-center gap-1 text-[11px] font-mono text-slate-400 bg-slate-100 dark:bg-slate-800/80 px-3 py-1.5 rounded-md border border-slate-200 dark:border-slate-700/60">
        Status: Scheduled for future phase
      </span>
    </div>
  </Card>
</div>
