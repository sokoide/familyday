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

async function postJSON<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const code = (data as { error?: string }).error ?? `HTTP_${res.status}`;
    throw new Error(code);
  }
  return data as T;
}

export class FetchJudgeApi implements JudgeApi {
  async judge(input: JudgeInput): Promise<JudgeResult> {
    return postJSON<JudgeResult>("/api/judge", input);
  }
}

export class FetchEndingApi implements EndingApi {
  async resolve(input: EndingInput): Promise<EndingResult> {
    return postJSON<EndingResult>("/api/ending", input);
  }
}

export class FetchResultApi implements ResultApi {
  async load(id: string) {
    const res = await fetch(`/api/result/${encodeURIComponent(id)}`);
    if (!res.ok) throw new Error(`HTTP_${res.status}`);
    return res.json();
  }
}
