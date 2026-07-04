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
    judging: string; // 判定送信中のラベル
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
    noReason: string; // 理由が空の時のフォールバック
    lifeDown: string; // "ハート -1"
    lifeNone: string; // "ハート ±0"
    historyTitle: string; // 履歴エリアのタイトル
    historyEmpty: string; // 履歴がない時の表示
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
    title: "🐉 トムハラゴン城の秘宝",
    subtitle: "〜 AI魔法のゲームブック 〜",
    lines: [
      "あなたは ゆうしゃ。おひめさまを たすけるため、トムハラゴン城へ むかいます。",
      "こえで ゆうしゃの「こうどう」を いってね!（例: どいて! / ねがいを かなえて! / すりぬける!）"
    ],
    hint: "ハートが 3つ あります。なくならないように きをつけてね!",
    start: "ぼうけんを はじめる",
    pickLang: "ことばを えらぶ"
  },
  hud: { livesEmpty: "ハート なし" },
  stage: {
    prefix: (n) => `ステージ${n}: `,
    stages: [
      {
        title: "城門のゴーレム",
        situation:
          "ドラゴン城の まえに とうちゃくした。でも、おおきくて こわい「もんばんのゴーレム」が とおせんぼしている!",
        goal: "ゴーレムが とおせんぼ! ゆうしゃは どう する?  ことばで いってね!"
      },
      {
        title: "炎のトラップ",
        situation:
          "城の なかに はいると、つうろが がれきの ほのおで つつまれていて すすめない!",
        goal: "ほのおが みちを ふさいでる! ゆうしゃは どう する?  ことばで いってね!"
      },
      {
        title: "トムハラゴンとの さいしゅうけっせん",
        situation:
          "さいじょうかいで おひめさまを はっけん!でも、おこった ドラゴンが おそいかかってきた!",
        goal: "ドラゴンが おそいかかってきた! たおす? それとも なかよくする?  ことばで いってね!"
      }
    ]
  },
  input: {
    micLabel: "押して はなす",
    listening: "ききとりちゅう…",
    judging: "しんぱんちゅう…",
    micUnsupported: "文字で にゅうりょく",
    manualSummary: "こえが とおらないとき：文字で にゅうりょく",
    manualPlaceholder: "ゆうしゃは どう する? (ことばで いれてね)",
    manualBtn: "おくる"
  },
  judge: {
    bad: "おしい!もういちど かんがえて みよう! 💡",
    goodSuffix: "(ライフ -1) つぎのへやへ!",
    greatSuffix: "✨ つぎのへやへ!",
    noVoice: "きこえなかったよ!もういちど はなしてね 🎤",
    netError: "ちょっと つうしんエラー。もういちど ためしてね!",
    micDenied:
      "マイクの きょかが いません。したの じでにゅうりょくを つかってね。",
    noSpeech: "こえが きこえなかったよ!もういちど!",
    unsupported:
      "このブラウザでは こえが つかえません。じでにゅうりょくしてね。",
    generic: "えらーが おきました。もういちど!",
    noReason: "がんばったね!",
    lifeDown: "ハート -1",
    lifeNone: "ハート へらないよ!",
    historyTitle: "📜 これまでの ぼうけん",
    historyEmpty: "まだ ぼうけんを はじめてないよ!"
  },
  ending: {
    titles: {
      great: "🏆 でんせつの ゆうしゃ エンド!",
      success: "✨ がんばった ゆうしゃ エンド!",
      gameover: "😢 また ちょうせんしてね エンド!"
    },
    fallbackEmoji: { great: "🏆", success: "✨", gameover: "😢" },
    shortLabel: {
      great: "でんせつの ゆうしゃ",
      success: "がんばった ゆうしゃ",
      gameover: "また ちょうせんしてね"
    },
    fallbackTitle: "ぼうけん おわり!",
    fallbackStory:
      "つうしんが ちょっと つながりにくいようです。スタッフに きいてね。",
    netErrorTitle: "ぼうけん おわり!",
    netErrorStory:
      "つうしんが ちょっと つながりにくいようです。スタッフに きいてね。",
    qrNote: "スマホの カメラで よみとってね!",
    emailLabel: "メールで 送る",
    emailPlaceholder: "your@email.com",
    emailBtn: "メールで 送る",
    emailSubject: "ドラゴン城の秘宝・あなたのぼうけんのきろく",
    emailBody:
      "ゆうしゃの きろくは こちら!\n(このURLは1しゅうかんだけ みられます)",
    resultUrlLabel: "",
    restart: "もういちど",
    loading: "よみこみちゅう…",
    notFound: "きろくが みつかりませんでした"
  }
};

const en: Messages = {
  lang: "en",
  langName: "English",
  intro: {
    title: "🐉 Tom Harragon Castle's Secret",
    subtitle: "— AI Magic Gamebook —",
    lines: [
      "You are a hero, heading to the Dragon Castle to save the princess.",
      "Say what your hero does! (e.g.: Move! / Make a wish! / Slip past it!)"
    ],
    hint: "You have 3 hearts. Don't lose them all!",
    start: "Start the adventure",
    pickLang: "Choose language"
  },
  hud: { livesEmpty: "No hearts" },
  stage: {
    prefix: (n) => `Stage ${n}: `,
    stages: [
      {
        title: "The Castle Gate Golem",
        situation:
          "You reach the Tom Harragon Castle. But a huge, scary Gate Golem blocks the way!",
        goal: "The Golem blocks the way! What does your hero do? Say it!"
      },
      {
        title: "The Fire Trap",
        situation:
          "Inside the castle, the hallway is wrapped in fierce flames. You can't go forward!",
        goal: "Flames block the path! What does your hero do? Say it!"
      },
      {
        title: "The Final Tom Harragon Battle",
        situation:
          "On the top floor you find the princess! But an angry dragon attacks!",
        goal: "The dragon attacks! Fight it or befriend it? What does your hero do? Say it!"
      }
    ]
  },
  input: {
    micLabel: "Tap & speak",
    listening: "Listening…",
    judging: "Judging…",
    micUnsupported: "Type it instead",
    manualSummary: "If voice doesn't work: type it",
    manualPlaceholder: "What does your hero do? (type it)",
    manualBtn: "Send"
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
    noReason: "You gave it a try!",
    lifeDown: "Heart -1",
    lifeNone: "no damage!",
    historyTitle: "📜 Your adventure so far",
    historyEmpty: "Start your adventure!"
  },
  ending: {
    titles: {
      great: "🏆 Legendary Hero Ending!",
      success: "✨ Brave Hero Ending!",
      gameover: "😢 Try Again Ending!"
    },
    fallbackEmoji: { great: "🏆", success: "✨", gameover: "😢" },
    shortLabel: {
      great: "Legendary Hero",
      success: "Brave Hero",
      gameover: "Try Again"
    },
    fallbackTitle: "Adventure Over!",
    fallbackStory: "The connection seems weak. Please ask the staff.",
    netErrorTitle: "Adventure Over!",
    netErrorStory: "The connection seems weak. Please ask the staff.",
    qrNote: "Scan with your phone camera!",
    emailLabel: "Send by email",
    emailPlaceholder: "your@email.com",
    emailBtn: "Send by email",
    emailSubject: "Dragon Castle — your adventure record",
    emailBody:
      "Here is your hero's record!\n(This link is available for one week.)",
    resultUrlLabel: "",
    restart: "Play again",
    loading: "Loading…",
    notFound: "Record not found"
  }
};

export const dictionaries: Record<Lang, Messages> = { ja, en };

export function getMessages(lang: Lang): Messages {
  return dictionaries[lang] ?? dictionaries.ja;
}

export function isLang(v: string): v is Lang {
  return v === "ja" || v === "en";
}
