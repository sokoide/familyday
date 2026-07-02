// app: エントリ。状態機械 + infra ワイヤリング + UI バインディング + i18n。
import { FetchEndingApi, FetchJudgeApi, FetchResultApi } from "../infra/fetchApi";
import { WebSpeechRecognizer } from "../infra/webSpeech";
import type { EndingResult, EndingType, JudgeResult } from "../ports/api";
import { STAGES } from "./stages";
import {
  MAX_LIVES,
  applyJudge,
  endingKind,
  forceEnding,
  initialState,
  start,
  type GameState,
} from "./state";
import { createTimer, formatTime } from "./timer";
import { getMessages, isLang, type Lang, type Messages } from "./messages";
import { buildMailto, fallbackImage, isValidEmail } from "../ui/share";
import { renderQR } from "../ui/qr";

const LANG_KEY = "fd-lang";

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
  setText("manual-summary", m.input.manualSummary);
  setText("qr-note", m.ending.qrNote);
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
  let lastResultUrl = "";

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
    endingStory: $("ending-story"),
    qrCanvas: $("qr-canvas") as HTMLCanvasElement,
    emailInput: $("email-input") as HTMLInputElement,
    emailBtn: $("btn-email") as HTMLButtonElement,
    resultUrl: $("result-url"),
    restart: $("btn-restart") as HTMLButtonElement,
  };

  els.lives.textContent = hearts(state.lives);
  els.micLabel.textContent = m.input.micLabel;
  els.historyTitle.textContent = m.judge.historyTitle;
  els.history.textContent = m.judge.historyEmpty;

  if (!speech.isSupported()) {
    els.mic.disabled = true;
    els.micLabel.textContent = m.input.micUnsupported;
    els.manualDetails.open = true;
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
    state = start({ ...state, lang });
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

  function renderStage(): void {
    const st = m.stage.stages[state.stageIndex];
    const ref = STAGES[state.stageIndex];
    els.stageTitle.textContent = m.stage.prefix(ref.number) + st.title;
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
    state = applyJudge(state, res);
    els.lives.textContent = hearts(state.lives);

    // 判定理由をトースト表示(5秒) + 履歴へ蓄積
    showReason(res, spoken);
    addHistory(state.stageIndex, spoken, res);

    if (state.phase === "ending") {
      releaseProcessing();
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
    } catch {
      // エラー時もフォールバック表示
      els.endingTitle.textContent = m.ending.netErrorTitle;
      els.endingStory.textContent = m.ending.netErrorStory;
      els.endingImage.onerror = () => {
        els.endingImage.src = fallbackImage(m.ending.fallbackEmoji.gameover, m.ending.shortLabel.gameover);
      };
    }
  }

  function renderEnding(res: EndingResult): void {
    lastResultUrl = res.resultUrl;
    els.endingTitle.textContent = m.ending.titles[res.endingType] ?? m.ending.fallbackTitle;
    els.endingStory.textContent = res.story;
    els.endingImage.src = res.imageUrl;
    els.endingImage.onerror = () => {
      els.endingImage.src = fallbackImage(
        m.ending.fallbackEmoji[res.endingType],
        m.ending.shortLabel[res.endingType],
      );
    };
    els.resultUrl.textContent = res.resultUrl;

    void renderQR(els.qrCanvas, res.resultUrl).catch(() => {
      /* QR 失敗時も URL テキストで救済済み */
    });
  }

  // メール送信
  els.emailInput.addEventListener("input", () => {
    els.emailBtn.disabled = !isValidEmail(els.emailInput.value);
  });
  els.emailBtn.addEventListener("click", () => {
    const email = els.emailInput.value;
    if (!isValidEmail(email) || !lastResultUrl) return;
    location.href = buildMailto(email, lastResultUrl, m.ending.emailSubject, m.ending.emailBody);
  });

  // もういちど
  els.restart.addEventListener("click", () => {
    state = initialState(newSessionId(), lang);
    els.lives.textContent = hearts(state.lives);
    els.timer.textContent = formatTime(300);
    els.emailInput.value = "";
    els.emailBtn.disabled = true;
    // 履歴・トースト・判定メッセージをクリア
    els.history.textContent = m.judge.historyEmpty;
    els.judgeReason.hidden = true;
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
      <h2 id="rv-title" class="ending-title">${m.ending.loading}</h2>
      <img id="rv-image" class="ending-image" alt="" />
      <p id="rv-story" class="ending-story"></p>
      <p class="qr-note" id="rv-url"></p>
    </section>`;
  const title = document.getElementById("rv-title")!;
  const img = document.getElementById("rv-image") as HTMLImageElement;
  const story = document.getElementById("rv-story")!;
  const url = document.getElementById("rv-url")!;
  try {
    const r = await resultApi.load(id);
    const t: EndingType = r.endingType;
    title.textContent = m.ending.titles[t] ?? m.ending.fallbackTitle;
    story.textContent = r.story;
    img.src = r.imageUrl;
    img.onerror = () => {
      img.src = fallbackImage(m.ending.fallbackEmoji[t], m.ending.shortLabel[t]);
    };
    url.textContent = r.resultUrl;
  } catch {
    title.textContent = m.ending.notFound;
  }
}

main().catch((err) => {
  console.error(err);
});
