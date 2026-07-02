// app: 多言語メッセージ定義。UI 文字列をハードコードせずここに集約。
// 新言語追加は dictionaries に1エントリ足すだけ。

export type Lang = "ja" | "en";

export type EndingType = "great" | "success" | "gameover";

export interface StageText {
  title: string;
  situation: string;
  goal: string;
}

export interface Messages {
  lang: Lang;
  langName: string; // セレクタ表示用

  intro: {
    title: string;
    subtitle: string;
    lines: string[]; // ストーリーカードの行
    hint: string;
    start: string;
    pickLang: string;
  };

  hud: {
    livesEmpty: string; // ライフ0表示(念のため)
  };

  stage: {
    prefix: (n: number) => string; // "ステージ1: " 等
    stages: StageText[]; // 3ステージ分の表示テキスト
  };

  input: {
    micLabel: string;
    listening: string;
    micUnsupported: string;
    manualSummary: string;
    manualPlaceholder: string;
    manualBtn: string;
  };

  judge: {
    bad: string;
    goodSuffix: string; // "{message} (-1 life) Next!"
    greatSuffix: string;
    noVoice: string;
    netError: string;
    micDenied: string;
    noSpeech: string;
    unsupported: string;
    generic: string;
  };

  ending: {
    titles: Record<EndingType, string>;
    fallbackEmoji: Record<EndingType, string>;
    shortLabel: Record<EndingType, string>;
    fallbackTitle: string;
    fallbackStory: string;
    netErrorTitle: string;
    netErrorStory: string;
    qrNote: string;
    emailLabel: string;
    emailPlaceholder: string;
    emailBtn: string;
    emailSubject: string;
    emailBody: string;
    resultUrlLabel: string;
    restart: string;
    loading: string;
    notFound: string;
  };
}

// 言語切替で並び順が狂わないよう、ステージテキストは固定長3。
const ja: Messages = {
  lang: "ja",
  langName: "日本語",
  intro: {
    title: "🐉 ドラゴン城の秘宝",
    subtitle: "〜 AI魔法のゲームブック 〜",
    lines: [
      "あなたは ゆうしゃ。おひめさまを たすけるため、ドラゴン城へ むかいます。",
      "こえで「じゅもん」を となえて、3つの しれんを つっぱしれよう!",
    ],
    hint: "ハートが 3つ あります。なくならないように きをつけてね!",
    start: "ぼうけんを はじめる",
    pickLang: "ことばを えらぶ",
  },
  hud: { livesEmpty: "ハート なし" },
  stage: {
    prefix: (n) => `ステージ${n}: `,
    stages: [
      {
        title: "城門のゴーレム",
        situation: "ドラゴン城の まえに とうちゃくした。でも、おおきくて こわい「もんばんのゴーレム」が とおせんぼしている!",
        goal: "ゴーレムを どかす、または とおりぬける じゅもんを となえよう!",
      },
      {
        title: "炎のトラップ",
        situation: "城の なかに はいると、つうろが げれきの ほのおで つつまれていて すすめない!",
        goal: "ほのおを けす、または あんぜんに とびこえる じゅもんを となえよう!",
      },
      {
        title: "ドラゴンとの さいしゅうけっせん",
        situation: "さいじょうかいで おひめさまを はっけん!でも、おこった ドラゴンが おそいかかってきた!",
        goal: "ドラゴンを たおす、または なかよくする じゅもんを となえよう!",
      },
    ],
  },
  input: {
    micLabel: "押して はなす",
    listening: "ききとりちゅう…",
    micUnsupported: "文字で にゅうりょく",
    manualSummary: "こえが とおらないとき：文字で にゅうりょく",
    manualPlaceholder: "じゅもんを かいてね",
    manualBtn: "となえる",
  },
  judge: {
    bad: "おしい!もういちど かんがえて みよう! 💡",
    goodSuffix: "(ライフ -1) つぎのへやへ!",
    greatSuffix: "✨ つぎのへやへ!",
    noVoice: "きこえなかったよ!もういちど はなしてね 🎤",
    netError: "ちょっと つうしんエラー。もういちど ためしてね!",
    micDenied: "マイクの きょかが いません。したの じでにゅうりょくを つかってね。",
    noSpeech: "こえが きこえなかったよ!もういちど!",
    unsupported: "このブラウザでは こえが つかえません。じでにゅうりょくしてね。",
    generic: "えらーが おきました。もういちど!",
  },
  ending: {
    titles: {
      great: "🏆 でんせつの ゆうしゃ エンド!",
      success: "✨ がんばった ゆうしゃ エンド!",
      gameover: "😢 また ちょうせんしてね エンド",
    },
    fallbackEmoji: { great: "🏆", success: "✨", gameover: "😢" },
    shortLabel: {
      great: "でんせつの ゆうしゃ",
      success: "がんばった ゆうしゃ",
      gameover: "また ちょうせんしてね",
    },
    fallbackTitle: "ぼうけん おわり!",
    fallbackStory: "でんわが ちょっと つながりにくいようです。スタッフに きいてね。",
    netErrorTitle: "ぼうけん おわり!",
    netErrorStory: "でんわが ちょっと つながりにくいようです。スタッフに きいてね。",
    qrNote: "スマホの カメラで よみとってね!",
    emailLabel: "メールで 送る",
    emailPlaceholder: "your@email.com",
    emailBtn: "メールで 送る",
    emailSubject: "ドラゴン城の秘宝・あなたのぼうけんのきろく",
    emailBody: "ゆうしゃの きろくは こちら!\n(このURLは1しゅうかんだけ みられます)",
    resultUrlLabel: "",
    restart: "もういちど",
    loading: "読み込み中…",
    notFound: "きろくが みつかりませんでした",
  },
};

const en: Messages = {
  lang: "en",
  langName: "English",
  intro: {
    title: "🐉 Dragon Castle's Secret",
    subtitle: "— AI Magic Gamebook —",
    lines: [
      "You are a hero, heading to the Dragon Castle to save the princess.",
      "Cast a spell with your voice and brave 3 challenges!",
    ],
    hint: "You have 3 hearts. Don't lose them all!",
    start: "Start the adventure",
    pickLang: "Choose language",
  },
  hud: { livesEmpty: "No hearts" },
  stage: {
    prefix: (n) => `Stage ${n}: `,
    stages: [
      {
        title: "The Castle Gate Golem",
        situation: "You reach the Dragon Castle. But a huge, scary Gate Golem blocks the way!",
        goal: "Cast a spell to move the golem or slip past it!",
      },
      {
        title: "The Fire Trap",
        situation: "Inside the castle, the hallway is wrapped in fierce flames. You can't go forward!",
        goal: "Cast a spell to put out the fire or leap over it safely!",
      },
      {
        title: "The Final Dragon Battle",
        situation: "On the top floor you find the princess! But an angry dragon attacks!",
        goal: "Cast a spell to defeat the dragon or befriend it!",
      },
    ],
  },
  input: {
    micLabel: "Tap & speak",
    listening: "Listening…",
    micUnsupported: "Type it instead",
    manualSummary: "If voice doesn't work: type it",
    manualPlaceholder: "Type your spell",
    manualBtn: "Cast",
  },
  judge: {
    bad: "So close! Think and try again! 💡",
    goodSuffix: "(-1 heart) Next room!",
    greatSuffix: "✨ Next room!",
    noVoice: "I couldn't hear you! Speak again 🎤",
    netError: "Connection trouble. Please try again!",
    micDenied: "Mic permission is off. Please use the text box below.",
    noSpeech: "I didn't hear anything. Try again!",
    unsupported: "Voice isn't available in this browser. Please type.",
    generic: "Something went wrong. Try again!",
  },
  ending: {
    titles: {
      great: "🏆 Legendary Hero Ending!",
      success: "✨ Brave Hero Ending!",
      gameover: "😢 Try Again Ending",
    },
    fallbackEmoji: { great: "🏆", success: "✨", gameover: "😢" },
    shortLabel: {
      great: "Legendary Hero",
      success: "Brave Hero",
      gameover: "Try Again",
    },
    fallbackTitle: "Adventure Over!",
    fallbackStory: "The connection seems weak. Please ask the staff.",
    netErrorTitle: "Adventure Over!",
    netErrorStory: "The connection seems weak. Please ask the staff.",
    qrNote: "Scan with your phone camera!",
    emailLabel: "Send by email",
    emailPlaceholder: "your@email.com",
    emailBtn: "Send email",
    emailSubject: "Dragon Castle — your adventure record",
    emailBody: "Here is your hero's record!\n(This link is available for one week.)",
    resultUrlLabel: "",
    restart: "Play again",
    loading: "Loading…",
    notFound: "Record not found",
  },
};

export const dictionaries: Record<Lang, Messages> = { ja, en };

export function getMessages(lang: Lang): Messages {
  return dictionaries[lang] ?? dictionaries.ja;
}

export function isLang(v: string): v is Lang {
  return v === "ja" || v === "en";
}
