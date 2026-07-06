<script lang="ts">
  import { X, SendHorizontal } from '@lucide/svelte';

  interface Props {
    tool: string;
    toolSchema?: { type: string; properties?: Record<string, any>; required?: string[] };
    oncancel: () => void;
    onsubmit: (args: Record<string, any>) => void;
  }

  let { tool, toolSchema, oncancel, onsubmit }: Props = $props();

  let argsText = $state('{}');
  let error = $state('');

  function handleSubmit() {
    try {
      const args = JSON.parse(argsText);
      onsubmit(args);
    } catch (e) {
      error = 'JSON 格式错误';
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') oncancel();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="overlay" onclick={oncancel} role="presentation">
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
  <div class="dialog" onclick={(e) => e.stopPropagation()} role="dialog" aria-label="调用工具" tabindex="-1">
    <div class="dialog-header">
      <h3>调用工具</h3>
      <button class="close-btn" onclick={oncancel} type="button" aria-label="关闭">
        <X size={16} />
      </button>
    </div>

    <div class="dialog-body">
      <div class="dialog-tool-name">
        <span class="label">工具</span>
        <code>{tool}</code>
      </div>

      <div class="dialog-args">
        <span class="label">参数 (JSON)</span>
        <textarea
          bind:value={argsText}
          class="args-input"
          rows={6}
          placeholder={'{"key": "value"}'}
        ></textarea>
        {#if error}
          <p class="error-msg">{error}</p>
        {/if}
      </div>
    </div>

    <div class="dialog-footer">
      <button class="btn-cancel" onclick={oncancel} type="button">取消</button>
      <button class="btn-submit" onclick={handleSubmit} type="button">
        <SendHorizontal size={14} />
        <span>发送</span>
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: color-mix(in srgb, var(--color-mo) 70%, transparent);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    backdrop-filter: blur(2px);
  }

  .dialog {
    width: 480px;
    max-width: 90vw;
    max-height: 80vh;
    background: var(--color-dailan);
    border: 1px solid var(--color-yaqing);
    border-radius: 14px;
    box-shadow: 0 24px 64px rgba(0,0,0,0.4);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--color-yaqing);
  }

  .dialog-header h3 {
    font-size: 15px;
    font-weight: 600;
    color: var(--color-supai);
    margin: 0;
  }

  .close-btn {
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

  .close-btn:hover {
    background: var(--color-yaqing);
    color: var(--color-supai);
  }

  .dialog-body {
    padding: 20px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .dialog-tool-name {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .label {
    font-size: 11px;
    font-weight: 600;
    color: var(--color-shuang);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    flex-shrink: 0;
  }

  .dialog-tool-name code {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--color-liujin);
  }

  .dialog-args {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .args-input {
    width: 100%;
    padding: 12px;
    border: 1px solid var(--color-yaqing);
    border-radius: 8px;
    background: var(--color-mo);
    color: var(--color-supai);
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.6;
    resize: vertical;
    transition: border-color var(--transition-fast);
  }

  .args-input:focus {
    outline: none;
    border-color: var(--color-liujin);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-liujin) 15%, transparent);
  }

  .error-msg {
    color: var(--color-zhusha);
    font-size: 12px;
    margin: 0;
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 20px;
    border-top: 1px solid var(--color-yaqing);
  }

  .btn-cancel {
    padding: 8px 20px;
    border: 1px solid var(--color-yaqing);
    border-radius: 8px;
    background: transparent;
    color: var(--color-shuang);
    font-size: 13px;
    font-weight: 500;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: border-color var(--transition-fast), color var(--transition-fast);
  }

  .btn-cancel:hover {
    border-color: var(--color-die);
    color: var(--color-supai);
  }

  .btn-submit {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 20px;
    border: none;
    border-radius: 8px;
    background: var(--color-liujin);
    color: var(--color-mo);
    font-size: 13px;
    font-weight: 600;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: opacity var(--transition-fast);
  }

  .btn-submit:hover {
    opacity: 0.88;
  }
</style>
