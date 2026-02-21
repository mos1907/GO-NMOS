<script>
  import { get } from "svelte/store";
  import LoginPage from "./pages/LoginPage.svelte";
  import DashboardPage from "./pages/DashboardPage.svelte";
  import GlobalNotificationPopups from "./components/GlobalNotificationPopups.svelte";
  import { auth, clearAuth } from "./stores/auth.js";

  let state = get(auth);
  auth.subscribe((value) => {
    state = value;
  });

  function handleLogout() {
    clearAuth();
  }

  function handleLoginSuccess() {
    // auth store already updated
  }
</script>

<GlobalNotificationPopups />
{#if state?.token}
  <DashboardPage token={state.token} user={state.user} onLogout={handleLogout} />
{:else}
  <LoginPage onSuccess={handleLoginSuccess} />
{/if}
