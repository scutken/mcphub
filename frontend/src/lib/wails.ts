// 等待 Wails runtime (window.go) 就绪
export function waitForWails(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined') {
      reject(new Error('Wails runtime 仅可在浏览器环境中使用'));
      return;
    }
    // 已就绪
    if ((window as any).go) {
      resolve();
      return;
    }
    // 轮询等待，最多等 5 秒
    let attempts = 0;
    const maxAttempts = 100;
    const interval = setInterval(() => {
      if ((window as any).go) {
        clearInterval(interval);
        resolve();
        return;
      }
      attempts++;
      if (attempts >= maxAttempts) {
        clearInterval(interval);
        reject(new Error('Wails runtime 初始化超时'));
      }
    }, 50);
  });
}
