// app: エントリ。状態機械 + infra ワイヤリング + UI バインディング + i18n。
import { FetchEndingApi, FetchJudgeApi, FetchResultApi } from "../infra/fetchApi";
import { WebSpeechRecognizer } from "../infra/webSpeech";
import type { EndingResult, EndingType, JudgeResult } from "../ports/api";
import { STAGES } from "./stages";
import {
  MAX_LIVES,
  TOTAL_SECONDS,
  applyJudge,
  endingKind,
  forceEnding,
  initialState,
  type GameState,
} from "./state";
import { createTimer, formatTime } from "./timer";
import { getMessages, isLang, type Lang, type Messages } from "./messages";
import { buildAdventureMailto, fallbackImage, isValidEmail } from "../ui/share";

const LANG_KEY = "fd-lang";
const SESSION_STATE_KEY = "fd-game-state";
const SESSION_ENDING_KEY = "fd-ending-result";

// メール本文に記載する画像URLのbase。window.__FD_IMAGE_BASE__ で上書き可能(無ければデフォルト)。
const DEFAULT_IMAGE_BASE = "https://lab.sokoide.com/familyday/2026/images";
const IMAGE_BASE: string =
  (typeof window !== "undefined" && (window as unknown as { __FD_IMAGE_BASE__?: string }).__FD_IMAGE_BASE__) ||
  DEFAULT_IMAGE_BASE;

// メール署名(言語非依存の固定文)。
const MAIL_SIGNATURE = "Thank you for joining,\nTokyo Family Day team";

interface SavedSession {
  state: GameState;
  endingResult?: EndingResult;
}

function saveSession(state: GameState, endingResult?: EndingResult): void {
  try {
    sessionStorage.setItem(SESSION_STATE_KEY, JSON.stringify(state));
    if (endingResult) {
      sessionStorage.setItem(SESSION_ENDING_KEY, JSON.stringify(endingResult));
    } else {
      sessionStorage.removeItem(SESSION_ENDING_KEY);
    }
  } catch (e) {
    console.error("Failed to save session state:", e);
  }
}

function loadSession(): SavedSession | null {
  try {
    const stateStr = sessionStorage.getItem(SESSION_STATE_KEY);
    if (!stateStr) return null;
    const state = JSON.parse(stateStr) as GameState;
    const endingResultStr = sessionStorage.getItem(SESSION_ENDING_KEY);
    const endingResult = endingResultStr ? JSON.parse(endingResultStr) as EndingResult : undefined;
    return { state, endingResult };
  } catch (e) {
    console.error("Failed to load session state:", e);
    return null;
  }
}

function clearSession(): void {
  sessionStorage.removeItem(SESSION_STATE_KEY);
  sessionStorage.removeItem(SESSION_ENDING_KEY);
}

function $(id: string): HTMLElement {
  const el = document.getElementById(id);
  if (!el) throw new Error(`missing #${id}`);
  return el;
}

function showPhase(phase: GameState["phase"]): void {
  document.querySelectorAll<HTMLElement>(".screen").forEach((s) => {
    s.hidden = s.dataset.phase !== phase;
  });
}

function hearts(lives: number): string {
  return "❤️".repeat(Math.max(0, Math.min(MAX_LIVES, lives))) +
    "🤍".repeat(Math.max(0, MAX_LIVES - lives));
}

function newSessionId(): string {
  // サーバのレートリミットキー。日内で一意なら十分。
  return "s" + Math.random().toString(36).slice(2, 10);
}

// 認識言語コード。Web Speech API 用。
function speechLang(lang: Lang): string {
  return lang === "en" ? "en-US" : "ja-JP";
}

// html lang 属性 + メッセージを一括反映。
function applyI18n(m: Messages): void {
  document.documentElement.lang = m.lang;
  setText("intro-title", m.intro.title);
  setText("intro-subtitle", m.intro.subtitle);
  setText("intro-line1", m.intro.lines[0]);
  setText("intro-line2", m.intro.lines[1]);
  setText("intro-hint", m.intro.hint);
  setText("btn-start", m.intro.start);
  setText("btn-practice", m.intro.practice);
  setText("manual-summary", m.input.manualSummary);
  setText("email-label", m.ending.emailLabel);
  setText("btn-email", m.ending.emailBtn);
  setText("btn-restart", m.ending.restart);
  const manualInput = document.getElementById("manual-input") as HTMLTextAreaElement | null;
  if (manualInput) manualInput.placeholder = m.input.manualPlaceholder;
  const emailInput = document.getElementById("email-input") as HTMLInputElement | null;
  if (emailInput) emailInput.placeholder = m.ending.emailPlaceholder;
  // 言語ボタンのアクティブ表示
  document.querySelectorAll<HTMLButtonElement>(".lang-switch button").forEach((b) => {
    b.classList.toggle("active", b.dataset.lang === m.lang);
  });
}

function setText(id: string, text: string): void {
  const el = document.getElementById(id);
  if (el) el.textContent = text;
}

function loadLang(): Lang {
  const saved = localStorage.getItem(LANG_KEY);
  return saved && isLang(saved) ? saved : "ja";
}

async function main(): Promise<void> {
  const judgeApi = new FetchJudgeApi();
  const endingApi = new FetchEndingApi();
  const resultApi = new FetchResultApi();
  const speech = new WebSpeechRecognizer();
  const timer = createTimer();

  let lang: Lang = loadLang();
  let m = getMessages(lang);
  applyI18n(m);

  let state = initialState(newSessionId(), lang);

  // --- /r/{id} ルート: 結果復元ビュー ---
  const pathMatch = location.pathname.match(/^\/r\/([^/]+)$/);
  if (pathMatch) {
    await renderResultPage(resultApi, pathMatch[1], m);
    return;
  }

  // --- ゲーム本体 ---
  const els = {
    lives: $("lives"),
    timer: $("timer"),
    stageTitle: $("stage-title"),
    stageImage: $("stage-image") as HTMLImageElement,
    stageSituation: $("stage-situation"),
    stageGoal: $("stage-goal"),
    judgeMsg: $("judge-message"),
    judgeReason: $("judge-reason") as HTMLElement,
    historyTitle: $("history-title") as HTMLElement,
    history: $("history") as HTMLUListElement,
    mic: $("btn-mic") as HTMLButtonElement,
    micLabel: document.querySelector<HTMLElement>(".mic-label")!,
    interim: $("interim"),
    manualInput: $("manual-input") as HTMLTextAreaElement,
    manualBtn: $("btn-manual") as HTMLButtonElement,
    manualDetails: document.querySelector<HTMLDetailsElement>(".manual")!,
    endingTitle: $("ending-title"),
    endingImage: $("ending-image") as HTMLImageElement,
    endingResult: $("ending-result") as HTMLElement,
    emailInput: $("email-input") as HTMLInputElement,
    emailBtn: $("btn-email") as HTMLButtonElement,
    restart: $("btn-restart") as HTMLButtonElement,
  };

  // --- セッション復元 ---
  const saved = loadSession();
  if (saved) {
    state = saved.state;
    lang = state.lang;
    m = getMessages(lang);
    applyI18n(m);

    els.lives.textContent = hearts(state.lives);
    els.micLabel.textContent = m.input.micLabel;
    els.historyTitle.textContent = m.judge.historyTitle;

    // 履歴復元
    els.history.textContent = "";
    if (state.history && state.history.length > 0) {
      state.history.forEach((hItem) => {
        const li = document.createElement("li");
        li.className = `history-item verdict-${hItem.verdict.toLowerCase()}`;
        const stageNo = hItem.stageIndex + 1;
        const lifeLabel = hItem.livesDelta < 0 ? `(${m.judge.lifeDown})` : `(${m.judge.lifeNone})`;
        const reason = hItem.reason?.trim() || m.judge.noReason;
        const quote = hItem.spoken.length > 30 ? hItem.spoken.slice(0, 30) + "…" : hItem.spoken;
        li.textContent = `[${m.stage.prefix(stageNo).trim()}] "${quote}" → ${hItem.verdict} ${lifeLabel}: ${reason}`;
        els.history.prepend(li);
      });
    } else {
      els.history.textContent = m.judge.historyEmpty;
    }

    if (!speech.isSupported()) {
      els.mic.disabled = true;
      els.micLabel.textContent = m.input.micUnsupported;
      els.manualDetails.open = true;
    }

    showPhase(state.phase);

    if (state.phase === "stage") {
      renderStage();
      timer.start(
        (rem) => {
          els.timer.textContent = formatTime(rem);
          state.timerRemaining = rem;
          saveSession(state);
        },
        () => {
          state = forceEnding(state);
          saveSession(state);
          void goEnding();
        },
        state.timerRemaining
      );
    } else if (state.phase === "ending") {
      if (saved.endingResult) {
        renderEnding(saved.endingResult);
      } else {
        void goEnding();
      }
    }
  } else {
    els.lives.textContent = hearts(state.lives);
    els.micLabel.textContent = m.input.micLabel;
    els.historyTitle.textContent = m.judge.historyTitle;
    els.history.textContent = m.judge.historyEmpty;

    if (!speech.isSupported()) {
      els.mic.disabled = true;
      els.micLabel.textContent = m.input.micUnsupported;
      els.manualDetails.open = true;
    }
  }

  function setLang(next: Lang): void {
    lang = next;
    m = getMessages(lang);
    state = { ...state, lang };
    localStorage.setItem(LANG_KEY, lang);
    applyI18n(m);
    if (state.phase === "stage") renderStage();
    els.historyTitle.textContent = m.judge.historyTitle;
    if (els.history.children.length === 0) {
      els.history.textContent = m.judge.historyEmpty;
    }
    if (!speech.isSupported()) {
      els.micLabel.textContent = m.input.micUnsupported;
    } else if (state.isProcessing) {
      els.micLabel.textContent = m.input.judging;
    } else {
      els.micLabel.textContent = m.input.micLabel;
    }
  }

  // 言語切替
  document.querySelectorAll<HTMLButtonElement>(".lang-switch button").forEach((b) => {
    b.addEventListener("click", () => {
      const next = b.dataset.lang;
      if (next && isLang(next)) setLang(next);
    });
  });

  function setInputsEnabled(enabled: boolean): void {
    els.mic.disabled = !enabled || !speech.isSupported();
    els.manualBtn.disabled = !enabled;
  }

  document.getElementById("btn-start")!.addEventListener("click", () => {
    // ぼうけんを1からやり直す: history・エンディング・sessionStorage を完全クリア
    clearSession();
    state = initialState(newSessionId(), lang);
    els.lives.textContent = hearts(state.lives);
    els.timer.textContent = formatTime(TOTAL_SECONDS);
    els.history.textContent = m.judge.historyEmpty;
    els.judgeReason.hidden = true;
    els.endingImage.src = "";
    els.endingTitle.textContent = "";
    els.endingResult.textContent = "";
    // エンディングで無効化したマイク・入力を再度有効化
    if (speech.isSupported()) {
      els.mic.disabled = false;
      els.micLabel.textContent = m.input.micLabel;
    }
    showPhase("stage");
    renderStage();
    timer.start(
      (rem) => {
        els.timer.textContent = formatTime(rem);
      },
      () => {
        // 時間切れ: 強制エンディング
        state = forceEnding(state);
        void goEnding();
      },
    );
  });

  // --- れんしゅうモード ---
  const practiceEls = {
    title: $("practice-title"),
    image: $("practice-image") as HTMLImageElement,
    situation: $("practice-situation"),
    goal: $("practice-goal"),
    message: $("practice-message"),
    result: $("practice-result") as HTMLElement,
    mic: $("btn-practice-mic") as HTMLButtonElement,
    micLabel: document.querySelector<HTMLElement>("#btn-practice-mic .mic-label")!,
    interim: $("practice-interim"),
    manualSummary: $("practice-manual-summary"),
    manualInput: $("practice-input") as HTMLTextAreaElement,
    manualBtn: $("btn-practice-manual") as HTMLButtonElement,
    manualDetails: document.querySelector<HTMLDetailsElement>('section[data-phase="practice"] .manual')!,
    back: $("btn-practice-back") as HTMLButtonElement,
  };
  let practiceProcessing = false;

  function initPractice(): void {
    practiceEls.title.textContent = m.practice.title;
    practiceEls.image.src = "/images/practice.jpg";
    practiceEls.image.alt = m.practice.title;
    practiceEls.situation.textContent = m.practice.situation;
    practiceEls.goal.textContent = m.practice.goal;
    practiceEls.message.hidden = true;
    practiceEls.result.textContent = "";
    practiceEls.result.className = "ending-result";
    practiceEls.interim.textContent = "";
    practiceEls.manualInput.value = "";
    practiceEls.manualSummary.textContent = m.practice.manualSummary;
    practiceEls.manualInput.placeholder = m.practice.manualPlaceholder;
    practiceEls.manualBtn.textContent = m.practice.manualBtn;
    practiceEls.back.textContent = m.practice.back;
    if (speech.isSupported()) {
      practiceEls.mic.disabled = false;
      practiceEls.micLabel.textContent = m.practice.micLabel;
    } else {
      practiceEls.mic.disabled = true;
      practiceEls.micLabel.textContent = m.input.micUnsupported;
      practiceEls.manualDetails.open = true;
    }
  }

  document.getElementById("btn-practice")!.addEventListener("click", () => {
    initPractice();
    showPhase("practice");
  });

  practiceEls.back.addEventListener("click", () => {
    showPhase("intro");
  });

  // 練習判定: フロント簡易キーワード判定(サーバー不要・即時)
  function judgePractice(text: string): boolean {
    const lower = text.toLowerCase();
    const successKeywords = [
      "こうばん", "交番", "とどけ", "届け", "とどける", "届ける", "わたす", "渡す",
      "もどす", "返す", "かえす",
      "police", "return", "deliver", "bring", "hand in", "koban",
    ];
    const failureKeywords = [
      "むし", "無視", "すて", "捨て", "あるく", "歩く", "いく", "行く", "とおる", "通る",
      "ignore", "walk", "leave", "pass", "skip", "abandon",
    ];
    if (successKeywords.some((kw) => lower.includes(kw.toLowerCase()))) return true;
    if (failureKeywords.some((kw) => lower.includes(kw.toLowerCase()))) return false;
    return false;
  }

  function submitPractice(text: string): void {
    const input = text.trim();
    if (!input || practiceProcessing) return;
    practiceProcessing = true;
    practiceEls.mic.disabled = true;
    practiceEls.manualBtn.disabled = true;
    practiceEls.interim.textContent = input;
    practiceEls.micLabel.textContent = m.practice.judging;
    practiceEls.message.hidden = true;

    const success = judgePractice(input);
    practiceEls.result.textContent = success ? m.practice.success : m.practice.failure;
    practiceEls.result.className = `ending-result ${success ? "clear" : "fail"}`;
    practiceEls.message.textContent = `"${input}"`;
    practiceEls.message.hidden = false;
    practiceEls.interim.textContent = "";

    if (speech.isSupported()) {
      practiceEls.mic.disabled = false;
      practiceEls.micLabel.textContent = m.practice.micLabel;
    }
    practiceEls.manualBtn.disabled = false;
    practiceProcessing = false;
  }

  practiceEls.mic.addEventListener("click", async () => {
    if (practiceProcessing) return;
    practiceEls.micLabel.textContent = m.input.listening;
    practiceEls.interim.textContent = "";
    try {
      const text = await speech.recognizeOnce(speechLang(lang), (t) => {
        practiceEls.interim.textContent = t;
      });
      practiceEls.micLabel.textContent = m.practice.micLabel;
      if (text) {
        submitPractice(text);
      } else {
        practiceEls.result.textContent = m.judge.noVoice;
        practiceEls.result.className = "ending-result fail";
      }
    } catch (err) {
      practiceEls.micLabel.textContent = m.practice.micLabel;
      console.error("[practice speech] error:", err);
      const msg = err instanceof Error ? err.message : "";
      if (msg.includes("not-allowed") || msg.includes("service-not-allowed")) {
        practiceEls.result.textContent = m.judge.micDenied;
      } else if (msg.includes("not-supported")) {
        practiceEls.result.textContent = m.judge.unsupported;
      } else if (msg.includes("no-speech")) {
        practiceEls.result.textContent = m.judge.noSpeech;
      } else {
        practiceEls.result.textContent = `${m.judge.generic} (${msg})`;
      }
      practiceEls.result.className = "ending-result fail";
    }
  });

  practiceEls.manualBtn.addEventListener("click", () => {
    submitPractice(practiceEls.manualInput.value);
  });
  practiceEls.manualInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submitPractice(practiceEls.manualInput.value);
    }
  });

  function renderStage(): void {
    const st = m.stage.stages[state.stageIndex];
    const ref = STAGES[state.stageIndex];
    els.stageTitle.textContent = m.stage.prefix(ref.number) + st.title;
    // 固定のステージ画像(public/images/s{N}.png)を表示
    els.stageImage.src = `/images/s${ref.number}.png`;
    els.stageImage.alt = st.title;
    els.stageSituation.textContent = st.situation;
    els.stageGoal.textContent = st.goal;
    els.judgeMsg.hidden = true;
    els.interim.textContent = "";
    els.manualInput.value = "";
    // マクラベルも「押して話す」に戻す(判定中から次ステージ遷移時に残るのを防ぐ)
    if (speech.isSupported()) {
      els.micLabel.textContent = m.input.micLabel;
    }
    setInputsEnabled(true);
  }

  function showJudge(msg: string): void {
    els.judgeMsg.textContent = msg;
    els.judgeMsg.hidden = false;
  }

  // 判定理由トースト(5秒でフェードアウト)。空欄時はフォールバック文。
  function showReason(res: JudgeResult, spoken: string): void {
    const reason = res.reason?.trim() || m.judge.noReason;
    const verdictLabel = res.verdict; // Great / Good / Bad
    const lifeLabel = res.livesDelta < 0 ? m.judge.lifeDown : m.judge.lifeNone;
    els.judgeReason.innerHTML = "";
    const head = document.createElement("div");
    head.className = "reason-head";
    head.textContent = `${verdictLabel}（${lifeLabel}）`;
    const body = document.createElement("div");
    body.className = "reason-body";
    body.textContent = `💡 ${reason}`;
    els.judgeReason.append(head, body);
    els.judgeReason.hidden = false;
    // アニメ: 表示 → 5秒後にフェード
    els.judgeReason.classList.remove("fade-out");
    void els.judgeReason.offsetWidth; // リフロー強制で再アニメ対応
    window.setTimeout(() => {
      els.judgeReason.classList.add("fade-out");
      window.setTimeout(() => {
        els.judgeReason.hidden = true;
      }, 600);
    }, 5000);
    void spoken; // (spoken は履歴で使用)
  }

  // 履歴へ「[ステージN] 発言 → 判定(ライフ変化): 理由」を prepend。
  function addHistory(stageIndex: number, spoken: string, res: JudgeResult): void {
    // 初回: 空表示をクリア
    if (els.history.textContent === m.judge.historyEmpty) {
      els.history.textContent = "";
    }
    const li = document.createElement("li");
    li.className = `history-item verdict-${res.verdict.toLowerCase()}`;
    const stageNo = stageIndex + 1;
    const lifeLabel = res.livesDelta < 0 ? `(${m.judge.lifeDown})` : `(${m.judge.lifeNone})`;
    const reason = res.reason?.trim() || m.judge.noReason;
    const quote = spoken.length > 30 ? spoken.slice(0, 30) + "…" : spoken;
    li.textContent = `[${m.stage.prefix(stageNo).trim()}] "${quote}" → ${res.verdict} ${lifeLabel}: ${reason}`;
    els.history.prepend(li);
  }

  // 判定API呼び出し共通
  async function submitInput(text: string): Promise<void> {
    const input = text.trim();
    if (!input || state.isProcessing) return;
    state = { ...state, isProcessing: true };
    setInputsEnabled(false);
    // 認識した文字をそのまま表示して「判定中」を明示(空クリアしない)
    els.interim.textContent = input;
    els.micLabel.textContent = m.input.judging;

    const stageId = STAGES[state.stageIndex].id;
    try {
      const res: JudgeResult = await judgeApi.judge({
        stageId,
        sessionId: state.sessionId,
        input,
        lang: state.lang,
      });
      onJudge(res, input);
    } catch {
      showJudge(m.judge.netError);
      els.micLabel.textContent = m.input.micLabel;
      setInputsEnabled(true);
      state = { ...state, isProcessing: false };
    }
    // 成功時の isProcessing 解除は onJudge 内で状態遷移に応じて行う(S2: 進行中の二重送信防止)
  }

  function releaseProcessing(): void {
    state = { ...state, isProcessing: false };
  }

  function onJudge(res: JudgeResult, spoken: string): void {
    const prevStageIndex = state.stageIndex;
    state = applyJudge(state, res);

    // 履歴追加
    state.history.push({
      stageIndex: prevStageIndex,
      spoken,
      verdict: res.verdict,
      livesDelta: res.livesDelta,
      reason: res.reason,
    });
    saveSession(state);

    els.lives.textContent = hearts(state.lives);

    // 判定理由をトースト表示(5秒) + 履歴へ蓄積
    showReason(res, spoken);
    addHistory(prevStageIndex, spoken, res);

    if (state.phase === "ending") {
      releaseProcessing();
      // エンディング中はマイクを「しゅうりょう」にして押せないようにする
      els.micLabel.textContent = m.input.ended;
      els.mic.disabled = true;
      els.manualBtn.disabled = true;
      void goEnding();
      return;
    }
    if (res.verdict === "Bad") {
      showJudge(m.judge.bad);
      els.interim.textContent = "";
      els.micLabel.textContent = m.input.micLabel;
      releaseProcessing();
      setInputsEnabled(true);
    } else if (res.verdict === "Good") {
      showJudge(`${res.message} ${m.judge.goodSuffix}`);
      // 進行遅延中は入力ロック維持 → renderStage で解除
      setTimeout(() => {
        renderStage();
        releaseProcessing();
      }, 1200);
    } else {
      // Great
      showJudge(`${res.message} ${m.judge.greatSuffix}`);
      setTimeout(() => {
        renderStage();
        releaseProcessing();
      }, 1000);
    }
  }

  // マイクボタン
  els.mic.addEventListener("click", async () => {
    if (state.isProcessing) return;
    els.micLabel.textContent = m.input.listening;
    els.interim.textContent = "";
    try {
      const text = await speech.recognizeOnce(speechLang(lang), (t) => {
        els.interim.textContent = t;
      });
      els.micLabel.textContent = m.input.micLabel;
      if (text) {
        await submitInput(text);
      } else {
        showJudge(m.judge.noVoice);
      }
    } catch (err) {
      els.micLabel.textContent = m.input.micLabel;
      // 診断のため実際のエラーコードを必ずコンソールへ(本番でもスタッフが確認可能)
      console.error("[speech] error:", err);
      const msg = err instanceof Error ? err.message : "";
      if (msg.includes("not-allowed") || msg.includes("service-not-allowed")) {
        // OS/ブラウザのマイク拒否。macOS では システム設定›マイク の Edge 権限を確認
        showJudge(m.judge.micDenied);
        els.manualDetails.open = true;
      } else if (msg.includes("not-supported")) {
        showJudge(m.judge.unsupported);
        els.manualDetails.open = true;
      } else if (msg.includes("no-speech")) {
        showJudge(m.judge.noSpeech);
      } else if (msg.includes("network")) {
        // クラウド音声サービスへ到達できない。ネットワーク/プロキシ/ファイアウォール確認
        showJudge(m.judge.netError);
      } else if (msg.includes("audio-capture")) {
        // マイクデバイスが取れない。別アプリが占有していないか確認
        showJudge(m.judge.micDenied);
        els.manualDetails.open = true;
      } else {
        // それ以外(aborted/unknown 等)。コードを表示して調査を容易にする
        showJudge(`${m.judge.generic} (${msg})`);
      }
    }
  });

  // 手動入力
  els.manualBtn.addEventListener("click", () => {
    void submitInput(els.manualInput.value);
  });
  els.manualInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void submitInput(els.manualInput.value);
    }
  });

  // --- エンディング ---
  async function goEnding(): Promise<void> {
    timer.stop();
    showPhase("ending");
    const kind = endingKind(state);
    try {
      const res: EndingResult = await endingApi.resolve({
        lives: kind.lives,
        finalAction: kind.finalAction,
        cleared: kind.cleared,
        sessionId: state.sessionId,
        lang: state.lang,
      });
      renderEnding(res);
      saveSession(state, res);
    } catch {
      // エラー時もフォールバック表示(通信エラー時は failed.png)
      els.endingTitle.textContent = m.ending.netErrorTitle;
      els.endingResult.textContent = m.ending.failedLabel;
      els.endingResult.className = "ending-result fail";
      els.endingImage.src = "/images/failed.png";
      els.endingImage.alt = m.ending.shortLabel.gameover;
      els.endingImage.onerror = () => {
        els.endingImage.src = fallbackImage(m.ending.fallbackEmoji.gameover, m.ending.shortLabel.gameover);
      };
    }
  }

  function renderEnding(res: EndingResult): void {
    els.endingTitle.textContent = m.ending.titles[res.endingType] ?? m.ending.fallbackTitle;

    // 固定のエンディング画像(public/images/{successful,failed}.png)を表示。
    // great/success → successful.png、gameover → failed.png
    const isClear = res.endingType === "great" || res.endingType === "success";
    els.endingImage.src = `/images/${isClear ? "successful" : "failed"}.png`;
    els.endingImage.alt = isClear ? m.ending.shortLabel.success : m.ending.shortLabel.gameover;
    els.endingResult.textContent = isClear ? m.ending.clearedLabel : m.ending.failedLabel;
    els.endingResult.className = `ending-result ${isClear ? "clear" : "fail"}`;
    els.endingImage.onerror = () => {
      els.endingImage.src = fallbackImage(
        m.ending.fallbackEmoji[res.endingType],
        m.ending.shortLabel[res.endingType],
      );
    };
  }

  // 冒険記録メールの本文を組み立てる
  function buildAdventureBody(endingType: EndingType): string[] {
    const lines: string[] = [];
    lines.push(m.ending.emailBody);
    lines.push("");
    lines.push(m.ending.adventureHeader);
    // ステージ毎の発言と判定を記録
    for (const h of state.history) {
      const stageNo = h.stageIndex + 1;
      const prefix = m.stage.prefix(stageNo).trim();
      const quote = h.spoken.length > 40 ? h.spoken.slice(0, 40) + "…" : h.spoken;
      lines.push(`[${prefix}] "${quote}" → ${h.verdict}`);
    }
    lines.push("");
    // エンディング結果ラベル
    const isClear = endingType === "great" || endingType === "success";
    lines.push(isClear ? m.ending.clearedLabel : m.ending.failedLabel);
    lines.push("");
    // 画像URL(ステージ1-4 + エンディング)
    lines.push(`${IMAGE_BASE}/s1.png`);
    lines.push(`${IMAGE_BASE}/s2.png`);
    lines.push(`${IMAGE_BASE}/s3.png`);
    lines.push(`${IMAGE_BASE}/s4.png`);
    lines.push(`${IMAGE_BASE}/${isClear ? "successful" : "failed"}.png`);
    lines.push("");
    // 署名
    lines.push(MAIL_SIGNATURE);
    return lines;
  }

  // メール送信
  els.emailInput.addEventListener("input", () => {
    els.emailBtn.disabled = !isValidEmail(els.emailInput.value);
  });
  els.emailBtn.addEventListener("click", () => {
    const email = els.emailInput.value;
    if (!isValidEmail(email)) return;
    const endingType = endingKind(state).cleared
      ? (state.dragonRoute === "befriend" ? "great" : "success")
      : "gameover";
    const bodyLines = buildAdventureBody(endingType as EndingType);
    location.href = buildAdventureMailto(email, m.ending.emailSubject, bodyLines);
  });

  // もういちど
  els.restart.addEventListener("click", () => {
    clearSession();
    state = initialState(newSessionId(), lang);
    els.lives.textContent = hearts(state.lives);
    els.timer.textContent = formatTime(300);
    els.emailInput.value = "";
    els.emailBtn.disabled = true;
    // 履歴・トースト・判定メッセージをクリア
    els.history.textContent = m.judge.historyEmpty;
    els.judgeReason.hidden = true;
    // エンディングで無効化したマイクを再有効化(intro → btn-start でも再設定される)
    if (speech.isSupported()) {
      els.mic.disabled = false;
    }
    showPhase("intro");
  });
}

// /r/{id} 用の結果復元ビュー
async function renderResultPage(
  resultApi: FetchResultApi,
  id: string,
  m: Messages,
): Promise<void> {
  const app = document.getElementById("app");
  if (!app) return;
  app.innerHTML = `
    <section class="screen result-view">
      <div class="story-card">
        <h2 id="rv-title" class="ending-title">${m.ending.loading}</h2>
        <img id="rv-image" class="ending-image" alt="" />
        <p id="rv-result" class="ending-result"></p>
      </div>
      <div class="share-area" style="margin-top: 2rem;">
        <div class="email-box">
          <label for="email-input" id="email-label">${m.ending.emailLabel}</label>
          <input id="email-input" type="email" placeholder="${m.ending.emailPlaceholder}" autocomplete="off" />
          <button id="btn-email" class="small-btn" disabled>${m.ending.emailBtn}</button>
        </div>
      </div>
    </section>`;
  const title = document.getElementById("rv-title")!;
  const img = document.getElementById("rv-image") as HTMLImageElement;
  const result = document.getElementById("rv-result")!;

  const emailInput = document.getElementById("email-input") as HTMLInputElement;
  const emailBtn = document.getElementById("btn-email") as HTMLButtonElement;

  try {
    const r = await resultApi.load(id);
    const t: EndingType = r.endingType;
    title.textContent = m.ending.titles[t] ?? m.ending.fallbackTitle;

    // 固定のエンディング画像を表示(great/success → successful、gameover → failed)
    const isClear = t === "great" || t === "success";
    img.src = `/images/${isClear ? "successful" : "failed"}.png`;
    img.alt = isClear ? m.ending.shortLabel.success : m.ending.shortLabel.gameover;
    result.textContent = isClear ? m.ending.clearedLabel : m.ending.failedLabel;
    result.className = `ending-result ${isClear ? "clear" : "fail"}`;
    img.onerror = () => {
      img.src = fallbackImage(m.ending.fallbackEmoji[t], m.ending.shortLabel[t]);
    };

    emailInput.addEventListener("input", () => {
      emailBtn.disabled = !isValidEmail(emailInput.value);
    });
    emailBtn.addEventListener("click", () => {
      const email = emailInput.value;
      if (!isValidEmail(email)) return;
      // 復元ビューでは冒険履歴が不明なため、画像URL+署名のみ
      const bodyLines = [
        m.ending.emailBody,
        "",
        isClear ? m.ending.clearedLabel : m.ending.failedLabel,
        "",
        `${IMAGE_BASE}/s1.png`,
        `${IMAGE_BASE}/s2.png`,
        `${IMAGE_BASE}/s3.png`,
        `${IMAGE_BASE}/s4.png`,
        `${IMAGE_BASE}/${isClear ? "successful" : "failed"}.png`,
        "",
        MAIL_SIGNATURE,
      ];
      location.href = buildAdventureMailto(email, m.ending.emailSubject, bodyLines);
    });
  } catch {
    title.textContent = m.ending.notFound;
    const shareArea = app.querySelector(".share-area");
    if (shareArea) (shareArea as HTMLElement).style.display = "none";
  }
}

main().catch((err) => {
  console.error(err);
});
