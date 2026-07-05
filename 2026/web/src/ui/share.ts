// ui: メール送信(mailto:) と fallback 画像のヘルパー。副作用なしでテスト可能。
// 表示文言は呼び出し側(messages)から渡す — ここにハードコードしない。

// メールアドレスの簡易バリデーション
export function isValidEmail(s: string): boolean {
  const v = s.trim();
  if (v.length === 0 || v.length > 254) return false;
  // 実用的な簡易チェック
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v);
}

// mailto: URL を構築。本文は行配列を改行で結合。subject は多言語文字列。
export function buildAdventureMailto(email: string, subject: string, bodyLines: string[]): string {
  const to = email.trim();
  const body = bodyLines.join("\n");
  return `mailto:${encodeURIComponent(to)}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
}

// fallback 画像 data URI(SVG)。画像生成失敗時に <img onerror> で差し替え。
// emoji は共通、label は呼び出し側から渡す(多言語)。
export function fallbackImage(emoji: string, label: string): string {
  const svg = `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="400" height="400" viewBox="0 0 400 400">
  <rect width="400" height="400" fill="#fff7e6"/>
  <text x="200" y="200" font-size="140" text-anchor="middle" dominant-baseline="central">${emoji}</text>
  <text x="200" y="320" font-size="32" text-anchor="middle" fill="#7a5230" font-family="sans-serif">${escapeXml(label)}</text>
</svg>`;
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

function escapeXml(s: string): string {
  return s.replace(/[<>&'"]/g, (c) => {
    switch (c) {
      case "<": return "&lt;";
      case ">": return "&gt;";
      case "&": return "&amp;";
      case "'": return "&apos;";
      case '"': return "&quot;";
      default: return c;
    }
  });
}
