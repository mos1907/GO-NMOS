<script>
  import { api } from "../lib/api.js";
  import { setAuth } from "../stores/auth.js";

  let username = "admin";
  let password = "admin";
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
    } finally {
      loading = false;
    }
  }
</script>

<main style="max-width: 420px; margin: 40px auto; font-family: sans-serif;">
  <h1>go-NMOS Login</h1>
  <label>Username</label>
  <input bind:value={username} style="width:100%;padding:10px;margin:8px 0;" />
  <label>Password</label>
  <input type="password" bind:value={password} style="width:100%;padding:10px;margin:8px 0;" />

  {#if error}
    <p style="color:#c00">{error}</p>
  {/if}

  <button on:click={submit} disabled={loading} style="padding:10px 16px;">
    {loading ? "Signing in..." : "Sign in"}
  </button>
</main>
