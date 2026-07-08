<script lang="ts">
  import { X, SendHorizontal, FileJson, List } from '@lucide/svelte';
  import { ConnectServer } from '$lib/../../wailsjs/go/main/App';
  import { waitForWails } from '$lib/wails';

  interface Props {
    oncancel: () => void;
    onsuccess: () => void;
  }

  let { oncancel, onsuccess }: Props = $props();

  let tab = $state<'manual' | 'json'>('manual');
  let name = $state('');
  let url = $state('');
  let transport = $state('auto');
  let headersText = $state('');
  let loading = $state(false);
  let error = $state('');

  // JSON 导入相关状态
  let jsonText = $state('');
  let jsonParseError = $state('');

  const transportOptions = [
    { value: 'auto', label: 'Auto (自动检测)' },
    { value: 'streamable', label: 'Streamable HTTP' },
  ];

  function parseHeaders(text: string): Record<string, string> {
    const headers: Record<string, string> = {};
    for (const line of text.split('\n')) {
      const trimmed = line.trim();
      if (!trimmed) continue;
      const colonIdx = trimmed.indexOf(':');
      if (colonIdx === -1) continue;
      const key = trimmed.slice(0, colonIdx).trim();
      const value = trimmed.slice(colonIdx + 1).trim();
      if (key) headers[key] = value;
    }
    return headers;
  }

  /** 解析 JSON 配置并填充表单 */
  function parseJsonConfig() {
    jsonParseError = '';
    try {
      const parsed = JSON.parse(jsonText);

      // 支持两种格式：直接 server 对象 或 { mcpServers: {...} }
      let servers: Record<string, any>;
      if (parsed.mcpServers) {
        servers = parsed.mcpServers;
      } else {
        servers = parsed;
      }

      const keys = Object.keys(servers);
      if (keys.length === 0) {
        jsonParseError = 'JSON 中未找到任何服务器配置';
        return;
      }

      const firstKey = keys[0];
      const config = servers[firstKey];
      if (!config || typeof config !== 'object') {
        jsonParseError = `服务器 "${firstKey}" 的配置无效`;
        return;
      }

      // 提取字段
      name = firstKey;
      url = config.url || '';

      // 提取 headers → 多行文本
      if (config.headers && typeof config.headers === 'object') {
        headersText = Object.entries(config.headers)
          .map(([k, v]) => `${k}: ${v}`)
          .join('\n');
      } else {
        headersText = '';
      }

      transport = 'auto';

      // 解析成功，切回手动 tab 让用户确认
      tab = 'manual';
    } catch (e: any) {
      jsonParseError = 'JSON 格式错误：' + (e?.message || '无法解析');
    }
  }

  async function handleSubmit() {
    // 前端校验
    if (!name.trim()) { error = '服务器名称不能为空'; return; }
    if (!url.trim()) { error = 'URL 不能为空'; return; }

    loading = true;
    error = '';
    try {
      const headers = parseHeaders(headersText);
      await waitForWails();
      await ConnectServer(name.trim(), url.trim(), headers, transport);
      onsuccess();
    } catch (e: any) {
      error = e?.message || '连接失败，请检查地址和网络';
    } finally {
      loading = false;
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
  <div class="dialog" onclick={(e) => e.stopPropagation()} role="dialog" aria-label="添加服务器" tabindex="-1">
    <div class="dialog-header">
      <h3>添加服务器</h3>
      <button class="close-btn" onclick={oncancel} type="button" aria-label="关闭">
        <X size={16} />
      </button>
    </div>

    <div class="dialog-body">
      <!-- Tab 切换 -->
      <div class="tab-bar">
        <button
          class="tab-btn"
          class:active={tab === 'manual'}
          onclick={() => tab = 'manual'}
          type="button"
        >
          <List size={14} />
          <span>手动填写</span>
        </button>
        <button
          class="tab-btn"
          class:active={tab === 'json'}
          onclick={() => tab = 'json'}
          type="button"
        >
          <FileJson size={14} />
          <span>粘贴 JSON</span>
        </button>
      </div>

      {#if tab === 'manual'}
        <!-- 手动填写表单 -->
        <div class="field">
          <label class="field-label" for="as-name">名称 <span class="required">*</span></label>
          <input
            id="as-name"
            type="text"
            bind:value={name}
            class="field-input"
            placeholder="my-server"
            disabled={loading}
          />
        </div>

        <div class="field">
          <label class="field-label" for="as-url">URL <span class="required">*</span></label>
          <input
            id="as-url"
            type="text"
            bind:value={url}
            class="field-input"
            placeholder="http://localhost:8080/mcp"
            disabled={loading}
          />
        </div>

        <div class="field">
          <label class="field-label" for="as-transport">传输协议</label>
          <select
            id="as-transport"
            bind:value={transport}
            class="field-select"
            disabled={loading}
          >
            {#each transportOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>

        <div class="field">
          <label class="field-label" for="as-headers">
            Headers <span class="optional">(选填，每行一个 Key: Value)</span>
          </label>
          <textarea
            id="as-headers"
            bind:value={headersText}
            class="field-textarea"
            rows={4}
            placeholder={'Authorization: Bearer xxx\nX-Custom: value'}
            disabled={loading}
          ></textarea>
        </div>

        {#if error}
          <p class="error-msg">{error}</p>
        {/if}
      {:else}
        <!-- JSON 导入面板 -->
        <div class="json-import">
          <p class="json-hint">
            粘贴 Claude Desktop 标准 MCP 配置 JSON，点击解析后自动填充表单。
          </p>
          <pre class="json-example">{`{
  "mcpServers": {
    "my-server": {
      "url": "http://127.0.0.1:3501/mcp",
      "headers": { "x-key": "xxx" }
    }
  }
}`}</pre>
          <textarea
            bind:value={jsonText}
            class="json-textarea"
            rows={8}
            placeholder={'{\n  "mcpServers": {\n    "serverName": {\n      "url": "...",\n      "headers": {}\n    }\n  }\n}'}
          ></textarea>
          <button class="parse-btn" onclick={parseJsonConfig} type="button">
            <FileJson size={14} />
            <span>解析并填充</span>
          </button>
          {#if jsonParseError}
            <p class="error-msg">{jsonParseError}</p>
          {/if}
        </div>
      {/if}
    </div>

    <div class="dialog-footer">
      <button class="btn-cancel" onclick={oncancel} type="button" disabled={loading}>取消</button>
      <button class="btn-submit" onclick={handleSubmit} type="button" disabled={loading}>
        {#if loading}
          <span>连接中...</span>
        {:else}
          <SendHorizontal size={14} />
          <span>连接</span>
        {/if}
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
    width: 560px;
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

  .tab-bar {
    display: flex;
    gap: 4px;
    padding: 2px;
    background: var(--color-mo);
    border-radius: 8px;
    border: 1px solid var(--color-yaqing);
  }

  .tab-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    flex: 1;
    padding: 7px 12px;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--color-shuang);
    font-size: 12px;
    font-weight: 500;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: background var(--transition-fast), color var(--transition-fast);
  }

  .tab-btn:hover {
    color: var(--color-supai);
  }

  .tab-btn.active {
    background: var(--color-dailan);
    color: var(--color-liujin);
  }

  .dialog-body {
    padding: 20px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .json-import {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .json-hint {
    margin: 0;
    font-size: 12px;
    color: var(--color-shuang);
    line-height: 1.6;
  }

  .json-example {
    margin: 0;
    padding: 10px 12px;
    background: var(--color-mo);
    border: 1px solid var(--color-yaqing);
    border-radius: 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    line-height: 1.6;
    color: var(--color-huiye);
    overflow-x: auto;
    white-space: pre;
  }

  .json-textarea {
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

  .json-textarea:focus {
    outline: none;
    border-color: var(--color-liujin);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-liujin) 15%, transparent);
  }

  .parse-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px 16px;
    border: 1px solid var(--color-liujin);
    border-radius: 8px;
    background: color-mix(in srgb, var(--color-liujin) 10%, transparent);
    color: var(--color-liujin);
    font-size: 13px;
    font-weight: 500;
    font-family: var(--font-sans);
    cursor: pointer;
    transition: background var(--transition-fast), color var(--transition-fast);
    align-self: flex-start;
  }

  .parse-btn:hover {
    background: var(--color-liujin);
    color: var(--color-mo);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--color-shuang);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .required {
    color: var(--color-zhusha);
  }

  .optional {
    font-weight: 400;
    text-transform: none;
    letter-spacing: normal;
    color: var(--color-huiye);
    font-size: 11px;
  }

  .field-input,
  .field-select,
  .field-textarea {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid var(--color-yaqing);
    border-radius: 8px;
    background: var(--color-mo);
    color: var(--color-supai);
    font-family: var(--font-sans);
    font-size: 13px;
    line-height: 1.5;
    transition: border-color var(--transition-fast);
  }

  .field-input:focus,
  .field-select:focus,
  .field-textarea:focus {
    outline: none;
    border-color: var(--color-liujin);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-liujin) 15%, transparent);
  }

  .field-input:disabled,
  .field-select:disabled,
  .field-textarea:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .field-textarea {
    font-family: var(--font-mono);
    font-size: 12px;
    resize: vertical;
  }

  .field-select {
    cursor: pointer;
    appearance: auto;
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

  .btn-cancel:hover:not(:disabled) {
    border-color: var(--color-die);
    color: var(--color-supai);
  }

  .btn-cancel:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  .btn-submit:hover:not(:disabled) {
    opacity: 0.88;
  }

  .btn-submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
