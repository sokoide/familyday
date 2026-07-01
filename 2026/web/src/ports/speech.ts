// SpeechRecognizer: 音声入力ポート。
// infra で Web Speech API 実装、テスト時はダミーに差し替え。
// lang は BCP47 ("ja-JP"/"en-US") で認識言語を切替。
export interface SpeechRecognizer {
  isSupported(): boolean;
  // 1発話を認識して最終文字列を返す。onInterim で途中経過を通知。
  recognizeOnce(lang: string, onInterim: (text: string) => void): Promise<string>;
  stop(): void;
}

export interface Clock {
  now(): number; // Unix ms
}

export interface Logger {
  info(msg: string): void;
  warn(msg: string): void;
}
