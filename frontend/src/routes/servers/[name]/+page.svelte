<script lang="ts">
  import { onMount, tick } from 'svelte';
  import ToolCard from '$lib/components/ToolCard.svelte';
  import CallDialog from '$lib/components/CallDialog.svelte';
  import { Wrench, X } from '@lucide/svelte';
  import { ListTools, CallTool } from '../../../../wailsjs/go/main/App';
  import { waitForWails } from '$lib/wails';

  let { data } = $props();

  let tools = $state<Array<{
    server: string;
    name: string;
    description?: string;
    inputSchema: { type: string; properties?: Record<string, any>; required?: string[] };
  }>>([]);

  let loading = $state(true);
  let error = $state('');

  let callingTool = $state<string | null>(null);

  // 调用结果
  let callResult = $state<{
    tool: string;
    isError: boolean;
    content: Array<{ type: string; text?: string; data?: string; mimeType?: string }>;
  } | null>(null);

  let callError = $state('');

  let resultRef = $state<HTMLDivElement | null>(null);

  async function loadTools() {
    loading = true;
    error = '';
    try {
      await waitForWails();
      tools = await ListTools(data.server);
    } catch (e: any) {
      error = e?.message || '加载工具列表失败';
      tools = [];
    } finally {
      loading = false;
    }
  }

  async function handleCall(toolName: string, args: Record<string, any>) {
    callingTool = null;
    callResult = null;
    callError = '';
    try {
      await waitForWails();
      const result = await CallTool(data.server, toolName, args);
      callResult = {
        tool: toolName,
        isError: result.isError,
        content: result.content,
      };
      // 滚动到结果区
      await tick();
      resultRef?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    } catch (e: any) {
      callError = e?.message || '调用失败';
      await tick();
      resultRef?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }

  // 页面挂载时加载工具列表
  onMount(() => {
    loadTools();
  });
</script>

<div class="server-detail">
  <div class="page-header">
    <h2 class="page-title">{data.server}</h2>
    <span class="status-badge connected">已连接</span>
    <span class="header-sep" aria-hidden="true"></span>
    <div class="tools-head">
      <Wrench size={16} class="text-liujin" />
      <h3>工具</h3>
      <span class="count">{tools.length}</span>
    </div>
  </div>

  <div class="tools-section">
    <!-- 调用结果展示（紧贴工具列表上方，方便查看） -->
    {#if callResult}
      <div class="result-section" class:fade-in={callResult} bind:this={resultRef}>
        <div class="result-header" class:is-error={callResult.isError}>
          <h3>调用结果 — {callResult.tool}</h3>
          {#if callResult.isError}
            <span class="result-badge error">Error</span>
          {:else}
            <span class="result-badge success">Success</span>
          {/if}
          <button class="close-result" onclick={() => callResult = null} type="button" aria-label="关闭结果">
            <X size={14} />
          </button>
        </div>
        <div class="result-content">
          {#each callResult.content as content, i}
            {#if content.type === 'text' && content.text}
              <pre class="result-text">{content.text}</pre>
            {:else if content.type === 'image' && content.data}
              <div class="result-image">
                <img src="data:{content.mimeType || 'image/png'};base64,{content.data}" alt="" />
              </div>
            {:else if content.data}
              <pre class="result-text">{content.data}</pre>
            {:else}
              <pre class="result-text">{JSON.stringify(content, null, 2)}</pre>
            {/if}
          {/each}
        </div>
      </div>
    {/if}

    {#if callError}
      <div class="result-section error" class:fade-in={callError} bind:this={resultRef}>
        <div class="result-header is-error">
          <h3>调用失败</h3>
          <button class="close-result" onclick={() => callError = ''} type="button" aria-label="关闭错误">
            <X size={14} />
          </button>
        </div>
        <div class="result-content">
          <pre class="result-text error-text">{callError}</pre>
        </div>
      </div>
    {/if}

    {#if loading}
      <div class="empty-tools">
        <p>加载中...</p>
      </div>
    {:else if error}
      <div class="empty-tools error">
        <p>{error}</p>
      </div>
    {:else if tools.length === 0}
      <div class="empty-tools">
        <p>该服务器没有提供工具</p>
      </div>
    {:else}
      <div class="tools-grid">
        {#each tools as tool}
          <ToolCard {tool} oncall={(_args) => callingTool = tool.name} />
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
    align-items: baseline;
    gap: 10px;
    margin-bottom: 24px;
    flex-wrap: wrap;
  }

  .page-title {
    font-size: 20px;
    font-weight: 600;
    margin: 0;
    color: var(--color-supai);
    line-height: 1;
  }

  .header-sep {
    width: 1px;
    height: 16px;
    background: var(--color-yaqing);
    align-self: center;
  }

  .tools-head {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .tools-head h3 {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-supai);
    margin: 0;
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

  .empty-tools.error {
    color: var(--color-zhusha);
    border-color: color-mix(in srgb, var(--color-zhusha) 30%, transparent);
  }

  /* 结果展示区（在 tools-section 内部，紧贴工具列表上方） */
  .result-section {
    margin-bottom: 16px;
    border: 1px solid var(--color-yaqing);
    border-radius: 10px;
    background: var(--color-dailan);
    overflow: hidden;
    opacity: 0;
    transform: translateY(-4px);
    transition: opacity var(--transition-base), transform var(--transition-base);
  }

  .result-section.fade-in {
    opacity: 1;
    transform: translateY(0);
  }

  .result-section.error {
    border-color: color-mix(in srgb, var(--color-zhusha) 30%, transparent);
  }

  .result-header {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 20px;
    border-bottom: 1px solid var(--color-yaqing);
  }

  .result-header.is-error {
    border-bottom-color: color-mix(in srgb, var(--color-zhusha) 30%, transparent);
  }

  .result-header h3 {
    font-size: 13px;
    font-weight: 600;
    color: var(--color-supai);
    margin: 0;
    flex: 1;
  }

  .result-badge {
    font-size: 10px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 4px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .result-badge.success {
    background: color-mix(in srgb, var(--color-shiqing) 15%, transparent);
    color: var(--color-shiqing);
  }

  .result-badge.error {
    background: color-mix(in srgb, var(--color-zhusha) 15%, transparent);
    color: var(--color-zhusha);
  }

  .close-result {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 4px;
    border: none;
    background: transparent;
    color: var(--color-shuang);
    cursor: pointer;
    transition: background var(--transition-fast), color var(--transition-fast);
  }

  .close-result:hover {
    background: var(--color-yaqing);
    color: var(--color-supai);
  }

  .result-content {
    padding: 16px 20px;
    overflow-x: auto;
  }

  .result-text {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.6;
    color: var(--color-supai);
    white-space: pre-wrap;
    word-break: break-all;
  }

  .error-text {
    color: var(--color-zhusha);
  }

  .result-image img {
    max-width: 100%;
    border-radius: 6px;
  }
</style>
