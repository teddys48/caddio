<script lang="ts">
  import { useAuth } from '../lib/stores/auth.svelte';
  import { useRouter } from '../lib/router/router.svelte';
  import { api } from '../lib/api/client';
  import Card from '../lib/components/common/Card.svelte';
  import Input from '../lib/components/common/Input.svelte';
  import Button from '../lib/components/common/Button.svelte';
  import { Server, LogIn, AlertCircle } from 'lucide-svelte';

  const auth = useAuth();
  const router = useRouter();

  let username = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state<string | undefined>(undefined);

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    error = undefined;

    if (!username.trim() || !password) {
      error = 'Username and password are required';
      return;
    }

    loading = true;
    try {
      const user = await api.login({
        username: username.trim(),
        password,
      });

      auth.setUser(user);
      router.navigate('/dashboard');
    } catch (err: any) {
      error = err.message || 'Invalid username or password';
    } finally {
      loading = false;
    }
  }
</script>

<div class="w-full max-w-md mx-auto py-8">
  <div class="text-center mb-6">
    <div class="w-12 h-12 rounded-xl bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 mx-auto flex items-center justify-center mb-3 shadow-sm">
      <Server class="w-6 h-6" />
    </div>
    <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">Caddio</h1>
    <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
      Log in to access server management dashboard
    </p>
  </div>

  <Card>
    <form onsubmit={handleSubmit} class="space-y-4">
      {#if error}
        <div class="p-3 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-600 dark:text-rose-400 flex items-center gap-2">
          <AlertCircle class="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      {/if}

      <Input
        id="login-username"
        label="Username"
        bind:value={username}
        placeholder="Enter username"
        required
        disabled={loading}
      />

      <Input
        id="login-password"
        label="Password"
        type="password"
        bind:value={password}
        placeholder="Enter password"
        required
        disabled={loading}
      />

      <div class="pt-2">
        <Button type="submit" variant="primary" class="w-full" {loading}>
          <LogIn class="w-4 h-4" />
          <span>Log in</span>
        </Button>
      </div>
    </form>
  </Card>
</div>
