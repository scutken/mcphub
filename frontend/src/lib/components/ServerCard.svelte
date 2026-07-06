<script lang="ts">
  import { Circle, ChevronRight } from 'lucide-svelte';

  interface Props {
    server: {
      name: string;
      status: string;
      url: string;
      transport: string;
    };
    active: boolean;
    onclick: () => void;
  }

  let { server, active, onclick }: Props = $props();

  let statusColor = $derived(
    server.status === 'connected' ? 'var(--color-shiqing)' :
    server.status === 'error' ? 'var(--color-zhusha)' :
    'var(--color-huiye)'
  );
</script>

<button class="server-card" class:active {onclick} type="button">
  <div class="server-indicator" style="color: {statusColor}">
    <Circle size={8} fill="currentColor" stroke="none" />
  </div>
  <div class="server-info">
    <div class="server-name">{server.name}</div>
    <div class="server-url">{server.url}</div>
  </div>
  <ChevronRight size={14} class="chevron" />
</button>

<style>
  .server-card {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 10px 12px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: transparent;
    color: var(--color-supai);
    cursor: pointer;
    transition: background var(--transition-fast), border-color var(--transition-fast);
    text-align: left;
    font-family: var(--font-sans);
    font-size: 13px;
  }

  .server-card:hover {
    background: var(--color-dailan);
  }

  .server-card.active {
    background: color-mix(in srgb, var(--color-liujin) 8%, transparent);
    border-color: color-mix(in srgb, var(--color-liujin) 25%, transparent);
  }

  .server-indicator {
    flex-shrink: 0;
    line-height: 0;
  }

  .server-info {
    flex: 1;
    min-width: 0;
  }

  .server-name {
    font-weight: 500;
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .server-url {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--color-shuang);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-top: 2px;
  }

  .chevron {
    flex-shrink: 0;
    color: var(--color-huiye);
    opacity: 0;
    transition: opacity var(--transition-fast);
  }

  .server-card:hover .chevron,
  .server-card.active .chevron {
    opacity: 1;
  }
</style>
