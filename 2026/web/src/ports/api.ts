// ports: アプリ副作用の抽象インタフェース。infra 側で具象を実装し、テスト時はモックに差し替え可能。

export type Verdict = "Great" | "Good" | "Bad";
export type DragonRoute = "defeat" | "befriend" | "";
export type EndingType = "great" | "success" | "gameover";

export interface JudgeInput {
  stageId: string;
  sessionId: string;
  input: string;
  lang: "ja" | "en";
}

export interface JudgeResult {
  verdict: Verdict;
  route: DragonRoute;
  message: string;
  livesDelta: number;
  advance: boolean;
}

export interface EndingInput {
  lives: number;
  finalAction: "defeat" | "befriend" | "gameover";
  cleared: boolean;
  sessionId: string;
  lang: "ja" | "en";
}

export interface EndingResult {
  endingId: string;
  endingType: EndingType;
  story: string;
  imageUrl: string;
  resultUrl: string;
}

// JudgeApi / EndingApi: バックエンドへの通信ポート
export interface JudgeApi {
  judge(input: JudgeInput): Promise<JudgeResult>;
}
export interface EndingApi {
  resolve(input: EndingInput): Promise<EndingResult>;
}

// ResultApi: /r/{id} の復元用
export interface ResultApi {
  load(id: string): Promise<{
    endingType: EndingType;
    story: string;
    imageUrl: string;
    resultUrl: string;
    createdAt: string;
  }>;
}
