<script lang="ts">
  import { page } from '$app/stores';
  import ToolCard from '$lib/components/ToolCard.svelte';
  import CallDialog from '$lib/components/CallDialog.svelte';
  import { Wrench } from '@lucide/svelte';

  let { data } = $props();

  let tools = $state([
    // Placeholder; replaced by Wails bindings
  ]);

  let callingTool = $state<string | null>(null);

  async function handleCall(toolName: string, args: Record<string, any>) {
    // TODO: Call via Go binding
    callingTool = null;
  }
</script>

<div class="server-detail">
  <div class="page-header">
    <div>
      <h2 class="page-title">{data.server}</h2>
      <p class="page-subtitle">
        <span class="status-badge connected">已连接</span>
        <span class="url-mono">{data.url || ''}</span>
      </p>
    </div>
  </div>

  <div class="tools-section">
    <div class="section-header">
      <Wrench size={16} class="text-liujin" />
      <h3>工具</h3>
      <span class="count">{tools.length}</span>
    </div>

    {#if tools.length === 0}
      <div class="empty-tools">
        <p>该服务器没有提供工具</p>
      </div>
    {:else}
      <div class="tools-grid">
        {#each tools as tool}
          <ToolCard {tool} oncall={(args) => handleCall(tool.name, args)} />
        {/each}
      </div>
    {/if}
  </div>
</div>

{#if callingTool}
  <CallDialog
    tool={callingTool}
    oncancel={() => callingTool = null}
    onsubmit={(args) => handleCall(callingTool, args)}
  />
{/if}

<style>
  .server-detail {
    max-width: 900px;
  }

  .page-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 32px;
  }

  .page-title {
    font-size: 22px;
    font-weight: 600;
    margin: 0 0 6px;
    color: var(--color-supai);
  }

  .page-subtitle {
    margin: 0;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .status-badge {
    font-size: 11px;
    font-weight: 500;
    padding: 2px 8px;
    border-radius: 4px;
  }

  .status-badge.connected {
    background: color-mix(in srgb, var(--color-shiqing) 15%, transparent);
    color: var(--color-shiqing);
  }

  .url-mono {
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--color-shuang);
  }

  .section-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
  }

  .section-header h3 {
    font-size: 14px;
    font-weight: 600;
    color: var(--color-supai);
    margin: 0;
  }

  .count {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--color-shuang);
    background: var(--color-yaqing);
    padding: 1px 6px;
    border-radius: 4px;
  }

  .tools-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 12px;
  }

  .empty-tools {
    padding: 48px;
    text-align: center;
    color: var(--color-shuang);
    background: var(--color-dailan);
    border: 1px dashed var(--color-yaqing);
    border-radius: 10px;
  }

  .empty-tools p {
    margin: 0;
    font-size: 13px;
  }
</style>
