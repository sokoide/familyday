// app: 状態機械(純粋ロジック)。副作用を持たないので単体テスト可能。
import type { Lang } from "./messages";
import type { DragonRoute, JudgeResult } from "../ports/api";

export type Phase = "intro" | "stage" | "ending";

export const MAX_LIVES = 3;
export const TOTAL_SECONDS = 300; // 5分

export interface HistoryItem {
  stageIndex: number;
  spoken: string;
  verdict: string;
  livesDelta: number;
  reason: string;
}

export interface GameState {
  phase: Phase;
  stageIndex: number; // 0..2
  lives: number; // 0..3
  dragonRoute: DragonRoute;
  sessionId: string;
  lang: Lang;
  isProcessing: boolean;
  timeUp: boolean;
  timerRemaining: number;
  history: HistoryItem[];
}

export function initialState(sessionId: string, lang: Lang): GameState {
  return {
    phase: "intro",
    stageIndex: 0,
    lives: MAX_LIVES,
    dragonRoute: "",
    sessionId,
    lang,
    isProcessing: false,
    timeUp: false,
    timerRemaining: TOTAL_SECONDS,
    history: [],
  };
}

export function start(s: GameState): GameState {
  return { ...s, phase: "stage", stageIndex: 0, lives: MAX_LIVES, timerRemaining: TOTAL_SECONDS, history: [] };
}

// 判定適用。次フェーズ/ライフ/ルートを計算して新しい state を返す。
export function applyJudge(s: GameState, res: JudgeResult): GameState {
  let lives = s.lives + res.livesDelta;
  if (lives < 0) lives = 0;
  if (lives > MAX_LIVES) lives = MAX_LIVES;

  const isLast = s.stageIndex >= 2;
  const route: DragonRoute = isLast ? res.route : "";

  // ライフ0 → 強制エンディング(gameover)
  if (lives <= 0) {
    return { ...s, lives, dragonRoute: route, phase: "ending" };
  }
  // 進行: Great/Good で次へ(Bad は同じステージ)
  if (res.advance) {
    if (isLast) {
      return { ...s, lives, dragonRoute: route, phase: "ending" };
    }
    return { ...s, lives, stageIndex: s.stageIndex + 1 };
  }
  // Bad: 同ステージでリトライ
  return { ...s, lives };
}

export function forceEnding(s: GameState): GameState {
  return { ...s, phase: "ending", timeUp: true };
}

// エンディング種決定(サーバでも決めるが、UI 表示の先行計算用)
export function endingKind(s: GameState): {
  lives: number;
  cleared: boolean;
  finalAction: "defeat" | "befriend" | "gameover";
} {
  if (s.lives <= 0) {
    return { lives: 0, cleared: false, finalAction: "gameover" };
  }
  // 最終ステージを cleared したか: timeUp でなければ yes
  const actuallyCleared = s.phase === "ending" && !s.timeUp;
  const route = s.dragonRoute === "befriend" ? "befriend" : "defeat";
  return {
    lives: s.lives,
    cleared: actuallyCleared,
    finalAction: actuallyCleared ? route : "gameover",
  };
}
