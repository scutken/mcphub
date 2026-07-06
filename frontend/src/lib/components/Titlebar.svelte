<script lang="ts">
  let { children } = $props();
</script>

<div class="titlebar">
  <div class="titlebar-left">
    <img src="/logo.svg" alt="MCPHub" class="titlebar-logo" />
    <span class="brand">MCP<span class="brand-accent">Hub</span></span>
  </div>
  <div class="titlebar-center">
    <!-- draggable region -->
  </div>
  <div class="titlebar-right">
    <span class="version">v1.0</span>
  </div>
</div>

<style>
  .titlebar {
    display: flex;
    align-items: center;
    height: 36px;
    padding: 0 16px;
    background: var(--color-dailan);
    border-bottom: 1px solid var(--color-yaqing);
    -webkit-app-region: drag;
    user-select: none;
  }

  .titlebar-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .titlebar-logo {
    width: 18px;
    height: 18px;
  }

  .brand {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-supai);
    letter-spacing: 0.02em;
  }

  .brand-accent {
    color: var(--color-liujin);
    font-weight: 400;
  }

  .titlebar-center {
    flex: 1;
    -webkit-app-region: drag;
  }

  .titlebar-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .version {
    font-size: 11px;
    color: var(--color-shuang);
    font-family: var(--font-mono);
  }
</style>
