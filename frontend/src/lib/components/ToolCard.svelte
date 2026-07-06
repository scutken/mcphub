<script lang="ts">
  import { ChevronRight, Play } from 'lucide-svelte';

  interface Props {
    tool: {
      name: string;
      description: string;
      inputSchema: {
        type: string;
        properties?: Record<string, any>;
        required?: string[];
      };
    };
    oncall: (args: Record<string, any>) => void;
  }

  let { tool, oncall }: Props = $props();

  let showParams = $state(false);

  let params = $derived(
    tool.inputSchema?.properties
      ? Object.entries(tool.inputSchema.properties).map(([name, schema]) => ({
          name,
          type: schema.type || 'string',
          description: schema.description || '',
          required: tool.inputSchema.required?.includes(name) ?? false,
        }))
      : []
  );
</script>

<div class="tool-card" class:expanded={showParams}>
  <button class="tool-header" onclick={() => showParams = !showParams} type="button">
    <div class="tool-name-row">
      <span class="tool-name">{tool.name}</span>
      {#if params.length > 0}
        <span class="param-count">{params.length}</span>
      {/if}
    </div>
    <ChevronRight size={14} class="expand-icon {showParams ? 'rotated' : ''}" />
  </button>

  {#if tool.description}
    <p class="tool-desc">{tool.description}</p>
  {/if}

  {#if showParams && params.length > 0}
    <div class="tool-params">
      <div class="params-label">参数</div>
      {#each params as param}
        <div class="param-row">
          <span class="param-name">{param.name}</span>
          <span class="param-type">{param.type}</span>
          {#if param.required}
            <span class="param-required">必填</span>
          {/if}
          {#if param.description}
            <span class="param-desc">{param.description}</span>
          {/if}
        </div>
      {/each}
    </div>

    <button class="call-btn" type="button" onclick={() => oncall({})}>
      <Play size={14} />
      <span>调用</span>
    </button>
  {/if}
</div>

<style>
  .tool-card {
    border: 1px solid var(--color-yaqing);
    border-radius: 10px;
    background: var(--color-dailan);
    overflow: hidden;
    transition: border-color var(--transition-fast);
  }

  .tool-card:hover {
    border-color: var(--color-die);
  }

  .tool-card.expanded {
    border-color: var(--color-liujin);
  }

  .tool-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 14px 16px;
    border: none;
    background: transparent;
    color: var(--color-supai);
    cursor: pointer;
    font-family: var(--font-sans);
    font-size: 13px;
    transition: background var(--transition-fast);
  }

  .tool-header:hover {
    background: color-mix(in srgb, var(--color-supai) 3%, transparent);
  }

  .tool-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .tool-name {
    font-family: var(--font-mono);
    font-weight: 500;
    font-size: 13px;
  }

  .param-count {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--color-shuang);
    background: var(--color-yaqing);
    padding: 1px 5px;
    border-radius: 3px;
  }

  .expand-icon {
    color: var(--color-huiye);
    transition: transform var(--transition-base);
    flex-shrink: 0;
  }

  .expand-icon :global(.rotated) {
    transform: rotate(90deg);
  }

  .tool-desc {
    margin: 0;
    padding: 0 16px 4px;
    font-size: 12px;
    color: var(--color-shuang);
    line-height: 1.6;
  }

  .tool-params {
    padding: 12px 16px;
    border-top: 1px solid var(--color-yaqing);
    background: color-mix(in srgb, var(--color-mo) 50%, var(--color-dailan));
  }

  .params-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--color-shuang);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 8px;
  }

  .param-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 0;
    font-size: 12px;
  }

  .param-name {
    font-family: var(--font-mono);
    color: var(--color-supai);
  }

  .param-type {
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--color-liujin);
    background: color-mix(in srgb, var(--color-liujin) 12%, transparent);
    padding: 1px 5px;
    border-radius: 3px;
  }

  .param-required {
    font-size: 10px;
    color: var(--color-zhusha);
  }

  .param-desc {
    color: var(--color-shuang);
    flex: 1;
  }

  .call-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: calc(100% - 32px);
    margin: 12px 16px 16px;
    padding: 8px 0;
    border: 1px solid var(--color-liujin);
    border-radius: 8px;
    background: color-mix(in srgb, var(--color-liujin) 8%, transparent);
    color: var(--color-liujin);
    font-size: 13px;
    font-weight: 500;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: background var(--transition-fast), color var(--transition-fast);
  }

  .call-btn:hover {
    background: var(--color-liujin);
    color: var(--color-mo);
  }
</style>
