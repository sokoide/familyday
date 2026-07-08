// infra: ポートの fetch 実装。
import type {
  EndingApi,
  EndingInput,
  EndingResult,
  JudgeApi,
  JudgeInput,
  JudgeResult,
  ResultApi,
} from "../ports/api";

const JUDGE_TIMEOUT_MS = 30_000;
const ENDING_TIMEOUT_MS = 60_000;
const RESULT_TIMEOUT_MS = 10_000;

async function withTimeout<T>(timeoutMs: number, run: (signal: AbortSignal) => Promise<T>): Promise<T> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), timeoutMs);
  try {
    return await run(controller.signal);
  } catch (e) {
    if (e instanceof DOMException && e.name === "AbortError") {
      throw new Error("TIMEOUT");
    }
    throw e;
  } finally {
    window.clearTimeout(timeout);
  }
}

async function postJSON<T>(url: string, body: unknown, timeoutMs: number): Promise<T> {
  return withTimeout(timeoutMs, async (signal) => {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      signal,
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      const code = (data as { error?: string }).error ?? `HTTP_${res.status}`;
      throw new Error(code);
    }
    return data as T;
  });
}

export class FetchJudgeApi implements JudgeApi {
  async judge(input: JudgeInput): Promise<JudgeResult> {
    return postJSON<JudgeResult>("/api/judge", input, JUDGE_TIMEOUT_MS);
  }
}

export class FetchEndingApi implements EndingApi {
  async resolve(input: EndingInput): Promise<EndingResult> {
    return postJSON<EndingResult>("/api/ending", input, ENDING_TIMEOUT_MS);
  }
}

export class FetchResultApi implements ResultApi {
  async load(id: string) {
    return withTimeout(RESULT_TIMEOUT_MS, async (signal) => {
      const res = await fetch(`/api/result/${encodeURIComponent(id)}`, { signal });
      if (!res.ok) throw new Error(`HTTP_${res.status}`);
      return res.json();
    });
  }
}
