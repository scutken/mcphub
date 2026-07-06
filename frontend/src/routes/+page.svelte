<script lang="ts">
  import { Wrench, ArrowRight } from '@lucide/svelte';
  import { goto } from '$app/navigation';
  import Logo from '$lib/components/Logo.svelte';
  import AddServerDialog from '$lib/components/AddServerDialog.svelte';
  import { ListServers } from '../../wailsjs/go/main/App';

  let showAddDialog = $state(false);

  async function browseTools() {
    try {
      const servers = await ListServers();
      const first = servers.find(s => s.status === 'connected') || servers[0];
      if (first) {
        goto(`/servers/${encodeURIComponent(first.name)}`);
      }
    } catch (e) {
      // no servers available, do nothing
    }
  }
</script>

<svelte:head>
  <title>MCPHub — MCP Server Manager</title>
</svelte:head>

<div class="welcome">
  <div class="welcome-icon">
    <Logo size={48} />
  </div>
  <h1 class="welcome-title">MCP<span class="accent">Hub</span></h1>
  <p class="welcome-desc">管理 MCP 服务器，发现和调用工具</p>

  <div class="quick-actions">
    <button class="quick-action" type="button" onclick={() => showAddDialog = true}>
      <span class="qa-icon">+</span>
      <span class="qa-label">添加服务器</span>
      <span class="qa-arrow"><ArrowRight size={14} /></span>
    </button>
    <button class="quick-action" type="button" onclick={browseTools}>
      <Wrench size={16} />
      <span class="qa-label">浏览工具</span>
      <span class="qa-arrow"><ArrowRight size={14} /></span>
    </button>
  </div>

  <div class="cli-hint">
    <p>CLI 快速开始</p>
    <code>mcphub connect &lt;name&gt; &lt;url&gt;</code>
    <code>mcphub tools &lt;server&gt;</code>
    <code>mcphub call &lt;server&gt; &lt;tool&gt;</code>
  </div>
</div>

{#if showAddDialog}
  <AddServerDialog
    oncancel={() => showAddDialog = false}
    onsuccess={() => showAddDialog = false}
  />
{/if}

<style>
  .welcome {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
    gap: 16px;
    padding-bottom: 60px;
  }

  .welcome-icon {
    width: 80px;
    height: 80px;
    border-radius: 20px;
    background: color-mix(in srgb, var(--color-liujin) 12%, transparent);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-liujin);
  }

  .welcome-title {
    font-size: 28px;
    font-weight: 600;
    margin: 0;
    color: var(--color-supai);
  }

  .accent {
    color: var(--color-liujin);
    font-weight: 400;
  }

  .welcome-desc {
    color: var(--color-shuang);
    font-size: 14px;
    margin: 0;
  }

  .quick-actions {
    display: flex;
    gap: 12px;
    margin-top: 8px;
  }

  .quick-action {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    border: 1px solid var(--color-yaqing);
    border-radius: 10px;
    background: var(--color-dailan);
    color: var(--color-supai);
    font-size: 13px;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: border-color var(--transition-fast), background var(--transition-fast);
  }

  .quick-action:hover {
    border-color: var(--color-liujin);
    background: var(--color-xuanqing);
  }

  .qa-icon {
    font-size: 16px;
    color: var(--color-liujin);
  }

  .qa-label {
    font-weight: 500;
  }

  .qa-arrow {
    color: var(--color-shuang);
    transition: transform var(--transition-fast);
  }

  .quick-action:hover .qa-arrow {
    transform: translateX(2px);
    color: var(--color-liujin);
  }

  .cli-hint {
    margin-top: 32px;
    padding: 16px 24px;
    background: var(--color-dailan);
    border: 1px solid var(--color-yaqing);
    border-radius: 10px;
    text-align: left;
    min-width: 340px;
  }

  .cli-hint p {
    font-size: 11px;
    font-weight: 600;
    color: var(--color-shuang);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin: 0 0 10px;
  }

  .cli-hint code {
    display: block;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--color-supai);
    padding: 4px 0;
  }
</style>
