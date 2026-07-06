<script lang="ts">
  import '../app.css';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import Titlebar from '$lib/components/Titlebar.svelte';
  import { page } from '$app/stores';

  let { children } = $props();
</script>

<Titlebar />

<div class="app-shell">
  <Sidebar />
  <main class="main-content">
    {@render children?.()}
  </main>
</div>

<style>
  .app-shell {
    display: flex;
    height: calc(100vh - 36px); /* minus titlebar */
    overflow: hidden;
  }

  .main-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    background: var(--color-mo);
  }
</style>
