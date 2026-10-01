<script lang="ts">
  import type { Site, SaveSiteRequest, ValidationResult, Upstream } from '../../types';
  import Button from '../common/Button.svelte';
  import { api } from '../../api/client';
  import {
    X,
    Plus,
    Trash2,
    CheckCircle2,
    AlertTriangle,
    ShieldCheck,
    Globe,
    Server
  } from 'lucide-svelte';

  interface Props {
    open: boolean;
    site?: Site | null;
    onClose: () => void;
    onSuccess: () => void;
  }

  let { open = false, site = null, onClose, onSuccess }: Props = $props();

  let name = $state('');
  let domain = $state('');
  let upstreams = $state<string[]>(['']);
  let enabled = $state(true);

  let saving = $state(false);
  let validating = $state(false);
  let validationResult = $state<ValidationResult | null>(null);
  let errorMessage = $state('');

  // Sync state whenever site prop changes
  $effect(() => {
    if (site) {
      name = site.name || site.domain;
      domain = site.domain;
      upstreams = site.upstreams && site.upstreams.length > 0 ? site.upstreams.map((u: Upstream) => u.address) : [''];
      enabled = site.enabled;
    } else {
      name = '';
      domain = '';
      upstreams = [''];
      enabled = true;
    }
    validationResult = null;
    errorMessage = '';
  });

  function addUpstream() {
    upstreams = [...upstreams, ''];
  }

  function removeUpstream(index: number) {
    if (upstreams.length > 1) {
      upstreams = upstreams.filter((_, i) => i !== index);
    }
  }

  function buildPayload(): SaveSiteRequest {
    return {
      id: site?.id,
      name: name.trim() || domain.trim(),
      domain: domain.trim(),
      upstreams: upstreams
        .filter((u) => u.trim() !== '')
        .map((u) => ({ address: u.trim() })),
      enabled,
    };
  }

  async function handleValidate() {
    validating = true;
    errorMessage = '';
    validationResult = null;

    try {
      const payload = buildPayload();
      validationResult = await api.validateSite(payload);
    } catch (err: any) {
      errorMessage = err.message || 'Validation request failed';
    } finally {
      validating = false;
    }
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    saving = true;
    errorMessage = '';

    try {
      const payload = buildPayload();
      if (site?.id) {
        await api.updateSite(site.id, payload);
      } else {
        await api.createSite(payload);
      }
      onSuccess();
      onClose();
    } catch (err: any) {
      errorMessage = err.message || 'Failed to save and apply site configuration';
    } finally {
      saving = false;
    }
  }
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto">
    <!-- Backdrop -->
    <div
      class="fixed inset-0 bg-slate-950/60 backdrop-blur-xs transition-opacity"
      onclick={onClose}
      aria-hidden="true"
    ></div>

    <!-- Modal Card -->
    <div class="relative w-full max-w-lg bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-xl overflow-hidden z-10 my-auto">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/50">
        <div class="flex items-center gap-2.5">
          <div class="p-2 rounded-md bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <Globe class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-slate-900 dark:text-slate-100">
              {site ? 'Edit Reverse Proxy Site' : 'Add Reverse Proxy Site'}
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              {site ? 'Update route parameters and apply changes to Caddy' : 'Configure a new domain route and backend upstreams'}
            </p>
          </div>
        </div>
        <button
          type="button"
          onclick={onClose}
          class="p-1.5 rounded-md text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Form Content -->
      <form onsubmit={handleSubmit} class="p-6 space-y-5">
        {#if errorMessage}
          <div class="p-3.5 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-700 dark:text-rose-300 flex items-start gap-2">
            <AlertTriangle class="w-4 h-4 shrink-0 mt-0.5" />
            <div class="leading-relaxed">{errorMessage}</div>
          </div>
        {/if}

        <!-- Domain Name -->
        <div class="space-y-1.5">
          <label for="domain-input" class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
            Domain Name <span class="text-rose-500">*</span>
          </label>
          <input
            id="domain-input"
            type="text"
            bind:value={domain}
            required
            placeholder="e.g. app.example.com or api.mycompany.org"
            class="w-full px-3.5 py-2 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md text-xs font-mono text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition-colors"
          />
          <p class="text-[11px] text-slate-500 dark:text-slate-400">
            Caddy automatically provisions free SSL/TLS certificates for public domains. Do not include http:// or https://.
          </p>
        </div>

        <!-- Site Display Name (Optional) -->
        <div class="space-y-1.5">
          <label for="name-input" class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
            Display Label (Optional)
          </label>
          <input
            id="name-input"
            type="text"
            bind:value={name}
            placeholder="e.g. Customer Portal Frontend"
            class="w-full px-3.5 py-2 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md text-xs text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition-colors"
          />
        </div>

        <!-- Upstream Targets -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <label for="upstream-0" class="block text-xs font-semibold text-slate-800 dark:text-slate-200">
              Backend Upstream Server(s) <span class="text-rose-500">*</span>
            </label>
            <button
              type="button"
              onclick={addUpstream}
              class="text-xs font-medium text-emerald-600 dark:text-emerald-400 hover:underline flex items-center gap-1 cursor-pointer"
            >
              <Plus class="w-3.5 h-3.5" />
              <span>Add Upstream</span>
            </button>
          </div>

          <div class="space-y-2">
            {#each upstreams as _, index}
              <div class="flex items-center gap-2">
                <div class="relative flex-1">
                  <Server class="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    id="upstream-{index}"
                    type="text"
                    bind:value={upstreams[index]}
                    required
                    placeholder="e.g. 127.0.0.1:8000 or backend-service:3000"
                    class="w-full pl-9 pr-3 py-2 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-md text-xs font-mono text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition-colors"
                  />
                </div>
                {#if upstreams.length > 1}
                  <button
                    type="button"
                    onclick={() => removeUpstream(index)}
                    class="p-2 text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 rounded-md hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors cursor-pointer"
                    title="Remove Upstream"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                {/if}
              </div>
            {/each}
          </div>
          <p class="text-[11px] text-slate-500 dark:text-slate-400">
            Address of internal backend server (host:port). Multiple upstreams will automatically load-balance requests.
          </p>
        </div>

        <!-- Validation Status Results Card -->
        {#if validationResult}
          <div class="p-3.5 rounded-md border text-xs space-y-1.5 {validationResult.valid ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-800 dark:text-emerald-300' : 'bg-rose-500/10 border-rose-500/20 text-rose-800 dark:text-rose-300'}">
            <div class="flex items-center gap-1.5 font-semibold">
              {#if validationResult.valid}
                <CheckCircle2 class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                <span>Validation Passed</span>
              {:else}
                <AlertTriangle class="w-4 h-4 text-rose-600 dark:text-rose-400" />
                <span>Validation Errors</span>
              {/if}
            </div>
            <p class="text-[11px]">{validationResult.message}</p>
            {#if validationResult.details && validationResult.details.length > 0}
              <ul class="list-disc list-inside space-y-0.5 text-[11px] pt-1">
                {#each validationResult.details as detail}
                  <li>{detail}</li>
                {/each}
              </ul>
            {/if}
          </div>
        {/if}

        <!-- Footer Action Buttons -->
        <div class="flex items-center justify-between pt-4 border-t border-slate-200 dark:border-slate-800">
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onclick={handleValidate}
            loading={validating}
          >
            <ShieldCheck class="w-3.5 h-3.5" />
            <span>Validate Config</span>
          </Button>

          <div class="flex items-center gap-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onclick={onClose}
              disabled={saving}
            >
              Cancel
            </Button>

            <Button
              type="submit"
              variant="primary"
              size="sm"
              loading={saving}
            >
              <CheckCircle2 class="w-3.5 h-3.5" />
              <span>{site ? 'Save & Apply' : 'Create & Apply'}</span>
            </Button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
