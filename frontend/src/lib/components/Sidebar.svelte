<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import ServerCard from './ServerCard.svelte';
  import { Layers, Plus, RefreshCw } from 'lucide-svelte';

  // TODO: Wire up to Wails bindings via Go
  let servers = $state([
    // Placeholder data; replaced by Wails binding at runtime
  ]);

  let activeServer = $state<string | null>(null);

  function selectServer(name: string) {
    activeServer = name;
    goto(`/servers/${name}`);
  }
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <div class="sidebar-title">
      <span class="text-liujin"><Layers size={16} /></span>
      <span>Servers</span>
    </div>
    <div class="sidebar-actions">
      <button class="icon-btn" title="Refresh all">
        <RefreshCw size={14} />
      </button>
      <button class="icon-btn icon-btn-primary" title="Add server">
        <Plus size={14} />
      </button>
    </div>
  </div>

  <div class="server-list">
    {#if servers.length === 0}
      <div class="empty-state">
        <p class="empty-text">没连接任何服务器</p>
        <p class="empty-hint">点击 + 添加 MCP 服务器</p>
      </div>
    {:else}
      {#each servers as server}
        <ServerCard
          {server}
          active={activeServer === server.name}
          onclick={() => selectServer(server.name)}
        />
      {/each}
    {/if}
  </div>

  <div class="sidebar-footer">
    <span class="mono text-xs text-shuang">
      <!-- Go binding provides connected count -->
    </span>
  </div>
</aside>

<style>
  .sidebar {
    width: 240px;
    min-width: 240px;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--color-xuanqing);
    border-right: 1px solid var(--color-yaqing);
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 16px 12px;
    border-bottom: 1px solid var(--color-yaqing);
  }

  .sidebar-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    font-weight: 600;
    color: var(--color-shuang);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .sidebar-actions {
    display: flex;
    gap: 4px;
  }

  .icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    border: none;
    background: transparent;
    color: var(--color-shuang);
    cursor: pointer;
    transition: background var(--transition-fast), color var(--transition-fast);
  }

  .icon-btn:hover {
    background: var(--color-yaqing);
    color: var(--color-supai);
  }

  .icon-btn-primary {
    color: var(--color-liujin);
  }

  .icon-btn-primary:hover {
    background: color-mix(in srgb, var(--color-liujin) 15%, transparent);
    color: var(--color-liujin);
  }

  .server-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }

  .empty-state {
    padding: 32px 16px;
    text-align: center;
  }

  .empty-text {
    color: var(--color-shuang);
    font-size: 13px;
    margin: 0 0 4px;
  }

  .empty-hint {
    color: var(--color-huiye);
    font-size: 11px;
    margin: 0;
  }

  .sidebar-footer {
    padding: 12px 16px;
    border-top: 1px solid var(--color-yaqing);
  }
</style>
