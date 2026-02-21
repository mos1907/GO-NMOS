<script>
  import { api } from "../lib/api.js";
  import { setAuth } from "../stores/auth.js";
  import { addNotification } from "../lib/notifications.js";

  let username = "";
  let password = "";
  let loading = false;
  let error = "";

  export let onSuccess;

  async function submit() {
    loading = true;
    error = "";
    try {
      const data = await api("/login", {
        method: "POST",
        body: { username, password }
      });
      setAuth(data.token, data.user);
      onSuccess?.();
    } catch (e) {
      error = e.message;
      addNotification("error", e.message);
    } finally {
      loading = false;
    }
  }
</script>

<main class="min-h-screen flex items-center justify-center bg-slate-950 px-4">
  <div class="w-full max-w-md rounded-2xl border border-slate-800 bg-gradient-to-b from-slate-950 to-slate-900/90 shadow-xl shadow-slate-900/70 p-8 space-y-6">
    <header class="space-y-2">
      <div class="inline-flex items-center gap-2 rounded-full border border-svelte/50 bg-slate-900/80 px-3 py-1">
        <span class="h-2 w-2 rounded-full bg-emerald-400 shadow-[0_0_0_3px_rgba(52,211,153,0.35)]" />
        <span class="text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-200">
          go-NMOS Control
        </span>
      </div>
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-white">Sign in to your workspace</h1>
        <p class="mt-1 text-sm text-slate-400">
          Use your go-NMOS credentials to access routing and monitoring tools.
        </p>
      </div>
    </header>

    <form class="space-y-4" on:submit|preventDefault={submit}>
      <div class="space-y-1.5">
        <label for="username" class="block text-xs font-medium text-slate-300 uppercase tracking-wide">
          Username
        </label>
        <input
          id="username"
          bind:value={username}
          class="w-full rounded-lg border border-slate-700 bg-slate-900/70 px-3 py-2.5 text-sm text-slate-50 placeholder-slate-500 outline-none focus:border-svelte focus:ring-2 focus:ring-svelte/40 transition"
          placeholder="e.g. admin"
          autocomplete="username"
        />
      </div>

      <div class="space-y-1.5">
        <div class="flex items-center justify-between">
          <label for="password" class="block text-xs font-medium text-slate-300 uppercase tracking-wide">
            Password
          </label>
        </div>
        <input
          id="password"
          type="password"
          bind:value={password}
          class="w-full rounded-lg border border-slate-700 bg-slate-900/70 px-3 py-2.5 text-sm text-slate-50 placeholder-slate-500 outline-none focus:border-svelte focus:ring-2 focus:ring-svelte/40 transition"
          placeholder="Your password"
          autocomplete="current-password"
        />
      </div>

      {#if error}
        <p class="text-xs font-medium text-red-400 bg-red-950/40 border border-red-800/60 rounded-lg px-3 py-2">
          {error}
        </p>
      {/if}

      <button
        type="submit"
        on:click|preventDefault={submit}
        disabled={loading}
        class="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-svelte px-4 py-2.5 text-sm font-semibold text-slate-950 shadow-[0_10px_40px_rgba(248,117,55,0.45)] hover:bg-orange-400 hover:shadow-[0_14px_45px_rgba(248,117,55,0.6)] active:translate-y-px disabled:opacity-60 disabled:shadow-none transition"
      >
        {#if loading}
          <span class="inline-flex h-3 w-3 rounded-full border-2 border-slate-950 border-t-transparent animate-spin" />
          <span>Signing in...</span>
        {:else}
          <span>Sign in</span>
        {/if}
      </button>
    </form>

    <p class="text-[11px] text-slate-500 text-center">
      Access is restricted to authorized operators. Contact your system administrator for credentials.
    </p>
  </div>
</main>
