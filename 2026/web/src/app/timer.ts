// app: 5分タイマー。setInterval ベース(テスト時は別実装に差し替え可能)。
import { TOTAL_SECONDS } from "./state";

export interface Timer {
  start(onTick: (remaining: number) => void, onZero: () => void, startFrom?: number): void;
  stop(): void;
  remaining(): number;
}

export function createTimer(intervalMs = 1000): Timer {
  let remaining = TOTAL_SECONDS;
  let handle: number | null = null;

  return {
    start(onTick, onZero, startFrom?: number) {
      remaining = startFrom !== undefined ? startFrom : TOTAL_SECONDS;
      onTick(remaining);
      handle = window.setInterval(() => {
        remaining -= 1;
        if (remaining <= 0) {
          remaining = 0;
          onTick(remaining);
          if (handle !== null) window.clearInterval(handle);
          handle = null;
          onZero();
          return;
        }
        onTick(remaining);
      }, intervalMs);
    },
    stop() {
      if (handle !== null) window.clearInterval(handle);
      handle = null;
    },
    remaining() {
      return remaining;
    },
  };
}

// 表示用 m:ss
export function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}
