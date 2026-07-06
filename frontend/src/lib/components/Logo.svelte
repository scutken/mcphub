<script lang="ts">
  interface Props {
    size?: number;
  }

  let { size = 32 }: Props = $props();
</script>

<svg
  viewBox="0 0 256 256"
  width={size}
  height={size}
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  aria-label="MCPHub Logo"
>
  <!-- 外圈八边形 -->
  <polygon
    points="128,28 188,52 220,100 220,156 188,204 128,228 68,204 36,156 36,100 68,52"
    fill="none"
    stroke="var(--color-liujin, #d4a853)"
    stroke-width="5"
    opacity="0.6"
  />
  <!-- 对角线装饰 -->
  <line x1="36" y1="100" x2="220" y2="156" stroke="var(--color-liujin, #d4a853)" stroke-width="1.5" opacity="0.12" />
  <line x1="36" y1="156" x2="220" y2="100" stroke="var(--color-liujin, #d4a853)" stroke-width="1.5" opacity="0.12" />
  <!-- 四向连接线 -->
  <line x1="128" y1="52" x2="128" y2="104" stroke="var(--color-liujin, #d4a853)" stroke-width="2.5" opacity="0.45" />
  <line x1="128" y1="152" x2="128" y2="204" stroke="var(--color-liujin, #d4a853)" stroke-width="2.5" opacity="0.45" />
  <line x1="68" y1="128" x2="104" y2="128" stroke="var(--color-liujin, #d4a853)" stroke-width="2.5" opacity="0.45" />
  <line x1="152" y1="128" x2="188" y2="128" stroke="var(--color-liujin, #d4a853)" stroke-width="2.5" opacity="0.45" />
  <!-- 四个外层节点 -->
  <circle cx="128" cy="46" r="10" fill="none" stroke="var(--color-liujin, #d4a853)" stroke-width="3" />
  <circle cx="128" cy="46" r="4" fill="var(--color-liujin, #d4a853)" />
  <circle cx="128" cy="210" r="10" fill="none" stroke="var(--color-liujin, #d4a853)" stroke-width="3" />
  <circle cx="128" cy="210" r="4" fill="var(--color-liujin, #d4a853)" />
  <circle cx="62" cy="128" r="10" fill="none" stroke="var(--color-liujin, #d4a853)" stroke-width="3" />
  <circle cx="62" cy="128" r="4" fill="var(--color-liujin, #d4a853)" />
  <circle cx="194" cy="128" r="10" fill="none" stroke="var(--color-liujin, #d4a853)" stroke-width="3" />
  <circle cx="194" cy="128" r="4" fill="var(--color-liujin, #d4a853)" />
  <!-- 中心枢纽 -->
  <circle cx="128" cy="128" r="26" fill="var(--color-dailan, #0f3460)" stroke="var(--color-liujin, #d4a853)" stroke-width="3.5" />
  <circle cx="128" cy="128" r="12" fill="var(--color-liujin, #d4a853)" opacity="0.9" />
</svg>
