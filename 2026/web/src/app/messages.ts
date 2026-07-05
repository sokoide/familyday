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
    practice: string; // れんしゅうモード
    pickLang: string;
  };

  practice: {
    title: string;
    situation: string;
    goal: string;
    micLabel: string;
    manualSummary: string;
    manualPlaceholder: string;
    manualBtn: string;
    judging: string;
    success: string;
    failure: string;
    back: string;
  };

  hud: {
    livesEmpty: string; // ライフ0表示(念のため)
  };

  stage: {
    prefix: (n: number) => string; // "ステージ1: " 等
    stages: StageText[]; // 4ステージ分の表示テキスト
  };

  input: {
    micLabel: string;
    listening: string;
    judging: string; // 判定送信中のラベル
    ended: string; // エンディング中のラベル（押せない）
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
    emailLabel: string;
    emailPlaceholder: string;
    emailBtn: string;
    emailSubject: string;
    emailBody: string;
    adventureHeader: string; // メール本文の冒険記録ヘッダ
    clearedLabel: string; // エンディング成功ラベル
    failedLabel: string; // エンディング失敗ラベル
    timeoutLabel: string; // タイムアップ失敗ラベル
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
    subtitle: "〜 AI ゲームブック 〜",
    lines: [
      "あなたは ゆうしゃ。おひめさまを たすけるため、トムハラゴン城へ むかいます。",
      "あなたは さまざまな まほうを 使えます。まほうの なまえは じゆうに決めてください。",
      "こえで ゆうしゃの「こうどう」を いってね!（例: どいて! / ねがいを かなえて! / すりぬける!）"
    ],
    hint: "ハートが 3つ あります。なくならないように きをつけてね!",
    start: "ぼうけんを はじめる",
    practice: "れんしゅう",
    pickLang: "ことばを えらぶ"
  },
  practice: {
    title: "れんしゅう",
    situation:
      "あるいていると、Aliceのものと おもわれる かばんの おとしものが ある。とおくに こうばん が みえる。どうする!",
    goal: "ゆうしゃは どう する?  ことばで いってね!",
    micLabel: "押して はなす",
    manualSummary: "こえが とおらないとき：文字で にゅうりょく",
    manualPlaceholder: "ゆうしゃは どう する? (ことばで いれてね)",
    manualBtn: "おくる",
    judging: "しんぱんちゅう…",
    success: "せいこう! かばんを こうばんに とどけたよ!",
    failure: "ざんねん! そのまま あるきつづけた ね…",
    back: "もどる"
  },
  hud: { livesEmpty: "ハート なし" },
  stage: {
    prefix: (n) => `ステージ${n}: `,
    stages: [
      {
        title: "遠くのお城",
        situation:
          "とおくにトムハラゴン城がみえる。おひめさまは そこにいるます。でも、目の前に 川が流れていて むこうにわたれない!",
        goal: "どうやって むこうにわたる?  ことばで いってね!"
      },
      {
        title: "城門のゴーレム",
        situation:
          "トムハラゴン城の まえに とうちゃくした。でも、おおきくて こわい「もんばんのゴーレム」が とおせんぼしている!",
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
    ended: "しゅうりょう",
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
      great: "🏆 でんせつの...",
      success: "✨ がんばった...",
      gameover: "💀 バッドエンド"
    },
    fallbackEmoji: { great: "🏆", success: "✨", gameover: "💀" },
    shortLabel: {
      great: "でんせつの...",
      success: "がんばった...",
      gameover: "バッドエンド"
    },
    fallbackTitle: "ぼうけん おわり!",
    fallbackStory:
      "つうしんが ちょっと つながりにくいようです。スタッフに きいてね。",
    netErrorTitle: "ぼうけん おわり!",
    netErrorStory:
      "つうしんが ちょっと つながりにくいようです。スタッフに きいてね。",
    emailLabel: "メールで 送る",
    emailPlaceholder: "your@email.com",
    emailBtn: "メールで 送る",
    emailSubject: "トムハラゴン城の秘宝・あなたのぼうけんのきろく",
    emailBody: "ゆうしゃの ぼうけんきろくを おくるよ!",
    adventureHeader: "— これまでの ぼうけん —",
    clearedLabel: "けっか: クリア!",
    failedLabel: "けっか: ざんねん!",
    timeoutLabel: "じかんが きたよ!",
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
    subtitle: "— AI Gamebook —",
    lines: [
      "You are a hero, heading to the Tom Harragon Castle to save the princess.",
      "You can use various spells. Name them as you wish.",
      "Say what your hero does! (e.g.: Move! / Make a wish! / Slip past it!)"
    ],
    hint: "You have 3 hearts. Don't lose them all!",
    start: "Start the adventure",
    practice: "Practice",
    pickLang: "Choose language"
  },
  practice: {
    title: "Practice",
    situation:
      "While walking, you find a lost bag that seems to belong to Alice. You can see a police box in the distance. What do you do!",
    goal: "What does your hero do? Say it!",
    micLabel: "Tap & speak",
    manualSummary: "If voice doesn't work: type it",
    manualPlaceholder: "What does your hero do? (type it)",
    manualBtn: "Send",
    judging: "Judging…",
    success: "Success! You delivered the bag to the police box!",
    failure: "Too bad! You just kept walking…",
    back: "Back"
  },
  hud: { livesEmpty: "No hearts" },
  stage: {
    prefix: (n) => `Stage ${n}: `,
    stages: [
      {
        title: "The Castle behind a river",
        situation:
          "The Tom Harragon Castle is visible in the distance. But there's a river in front of you, and you can't cross it!",
        goal: "How do you cross the river? Say it!"
      },
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
    ended: "End",
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
      gameover: "💀 Bad End"
    },
    fallbackEmoji: { great: "🏆", success: "✨", gameover: "💀" },
    shortLabel: {
      great: "Legendary Hero",
      success: "Brave Hero",
      gameover: "Bad End"
    },
    fallbackTitle: "Adventure Over!",
    fallbackStory: "The connection seems weak. Please ask the staff.",
    netErrorTitle: "Adventure Over!",
    netErrorStory: "The connection seems weak. Please ask the staff.",
    emailLabel: "Send by email",
    emailPlaceholder: "your@email.com",
    emailBtn: "Send by email",
    emailSubject: "Tom Harragon Castle — your adventure record",
    emailBody: "Here is your hero's adventure record!",
    adventureHeader: "— Your adventure —",
    clearedLabel: "Result: Cleared!",
    failedLabel: "Result: Try again!",
    timeoutLabel: "Time's up!",
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
