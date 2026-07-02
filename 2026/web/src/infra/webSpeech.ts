// infra: Web Speech API による SpeechRecognizer 実装。
import type { SpeechRecognizer } from "../ports/speech";

// 標準化されていない webkit 型の最小定義
interface SpeechRecognitionLike {
  lang: string;
  continuous: boolean;
  interimResults: boolean;
  maxAlternatives: number;
  start(): void;
  stop(): void;
  abort(): void;
  onresult: ((e: SpeechRecognitionEventLike) => void) | null;
  onerror: ((e: SpeechRecognitionErrorEventLike) => void) | null;
  onend: (() => void) | null;
}
interface SpeechRecognitionEventLike {
  resultIndex: number;
  results: ArrayLike<{
    0: { transcript: string };
    isFinal: boolean;
  }> & { length: number };
}
interface SpeechRecognitionErrorEventLike {
  error: string;
}

function getCtor():
  | (new () => SpeechRecognitionLike)
  | null {
  const w = window as unknown as {
    SpeechRecognition?: new () => SpeechRecognitionLike;
    webkitSpeechRecognition?: new () => SpeechRecognitionLike;
  };
  return w.SpeechRecognition ?? w.webkitSpeechRecognition ?? null;
}

export class WebSpeechRecognizer implements SpeechRecognizer {
  private rec: SpeechRecognitionLike | null = null;

  isSupported(): boolean {
    return getCtor() !== null;
  }

  recognizeOnce(lang: string, onInterim: (text: string) => void): Promise<string> {
    return new Promise((resolve, reject) => {
      const Ctor = getCtor();
      if (!Ctor) {
        reject(new Error("speech-not-supported"));
        return;
      }
      const rec = new Ctor();
      rec.lang = lang; // "ja-JP" / "en-US"
      rec.continuous = false;
      rec.interimResults = true;
      rec.maxAlternatives = 1;

      let finalText = "";
      let interimText = ""; // 直近の interim を保持(onend 時のフォールバック用)
      let settled = false;
      const settle = (fn: () => void) => {
        if (settled) return;
        settled = true;
        clearTimeout(to);
        fn();
      };

      rec.onresult = (e) => {
        let interim = "";
        for (let i = e.resultIndex; i < e.results.length; i++) {
          const r = e.results[i];
          if (r.isFinal) {
            finalText += r[0].transcript;
          } else {
            interim += r[0].transcript;
          }
        }
        if (interim) {
          interimText = interim; // 末確定でも保持しておく
          onInterim(interim);
        }
      };
      rec.onerror = (e) => settle(() => reject(new Error(`speech-${e.error}`)));
      // onend で isFinal が来ていないことがある(短い発話・間が空いた等)。
      // その場合は直近の interim を確定結果として採用し、文字を捨てない。
      rec.onend = () => settle(() => resolve((finalText || interimText).trim()));

      // 8秒で強制停止(暴走防止)。停止で onend が発火し settle される。
      const to = setTimeout(() => {
        try {
          rec.stop();
        } catch {
          /* ignore */
        }
      }, 8000);

      this.rec = rec;
      try {
        rec.start();
      } catch (err) {
        settle(() => reject(err));
      }
    });
  }

  stop(): void {
    try {
      this.rec?.stop();
    } catch {
      /* ignore */
    }
  }
}
