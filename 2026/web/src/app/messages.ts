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
    restart: string;
    restartConfirm: string;
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
    title: "🐉 ドラゴン城の ひほう",
    subtitle: "〜 AI ゲームブック 〜",
    lines: [
      "あなたは ゆうしゃ。おひめさまを たすけるため、ドラゴン城へ むかいます。",
      "どんな こうどうを しても だいじょうぶ。わざや まほうの なまえも、じゆうに きめてね。",
      "ゆうしゃが どうするか、こえで いってね！（れい：どいて！／ねがいを かなえて！／こっそり すりぬける！）"
    ],
    hint: "ハートは 3つ。ぜんぶ なくならないように、きをつけてね！",
    start: "ぼうけんを はじめる",
    practice: "れんしゅうする",
    pickLang: "ことばを えらぶ"
  },

  practice: {
    title: "れんしゅう",
    situation:
      "みちを あるいていると、アリスのものらしい かばんを みつけました。とおくに こうばんが みえます。どうする？",
    goal: "ゆうしゃは どうする？ こえで いってね！",
    micLabel: "おして 話す",
    manualSummary: "こえが つかえないときは、文字で こたえてね",
    manualPlaceholder: "ゆうしゃは どうする？",
    manualBtn: "おくる",
    judging: "こたえを みています…",
    success: "せいこう！ かばんを こうばんに とどけたよ！",
    failure: "ざんねん！ かばんを おいたまま、あるいていってしまったよ…",
    back: "はじめに もどる"
  },

  hud: {
    livesEmpty: "ハートが ないよ"
  },

  stage: {
    prefix: (n) => `ステージ ${n}: `,
    restart: "はじめから やりなおす",
    restartConfirm: "ほんとうに はじめから やりなおす？",

    stages: [
      {
        title: "川の むこうの お城",
        situation:
          "とおくに ドラゴン城が みえます。おひめさまは、きっと そこにいます。でも、めのまえに 川が ながれていて、むこうへ わたれません！",
        goal: "どうやって 川を わたる？ こえで いってね！"
      },
      {
        title: "お城の もんの ゴーレム",
        situation:
          "ドラゴン城に つきました。でも、おおきくて こわい ゴーレムが、もんの まえで とおせんぼしています！",
        goal: "ゴーレムが みちを ふさいでいる！ ゆうしゃは どうする？ こえで いってね！"
      },
      {
        title: "ほのおの わな",
        situation:
          "お城の なかへ はいると、はげしい ほのおが ろうかを ふさいでいます。このままでは、さきへ すすめません！",
        goal: "ほのおが みちを ふさいでいる！ ゆうしゃは どうする？ こえで いってね！"
      },
      {
        title: "ドラゴンとの さいごの たたかい",
        situation:
          "いちばん うえの かいで、おひめさまを みつけました！ でも、おこった ドラゴンが おそいかかってきます！",
        goal: "ドラゴンが きた！ たたかう？ なかよくする？ それとも、ほかの ほうほうを ためす？ こえで いってね！"
      }
    ]
  },

  input: {
    micLabel: "おして 話す",
    listening: "きいています…",
    judging: "はんていちゅう…",
    ended: "おわり",
    micUnsupported: "文字で こたえてね",
    manualSummary: "こえが つかえないときは、文字で こたえてね",
    manualPlaceholder: "ゆうしゃは どうする？",
    manualBtn: "おくる"
  },

  judge: {
    bad: "もう すこし！ もういちど かんがえてみよう！ 💡",
    goodSuffix: "(ハート -1) つぎの ステージへ！",
    greatSuffix: "✨ つぎの ステージへ！",
    noVoice: "こえが きこえなかったよ。もういちど 話してね！ 🎤",
    netError: "つうしんが うまく いかなかったよ。もういちど ためしてね！",
    micDenied:
      "マイクを つかうことが できません。したの らんに、文字で こたえてね。",
    noSpeech: "なにも きこえなかったよ。もういちど 話してね！",
    unsupported:
      "このブラウザでは マイクを つかえません。文字で こたえてね。",
    generic: "うまく いかなかったよ。もういちど ためしてね！",
    noReason: "ちがう ほうほうも かんがえてみよう！",
    lifeDown: "ハート -1",
    lifeNone: "ハートは へらないよ！",
    historyTitle: "📜 これまでの ぼうけん",
    historyEmpty: "ぼうけんは まだ はじまっていないよ！"
  },

  ending: {
    titles: {
      great: "🏆 でんせつの ゆうしゃ エンド！",
      success: "✨ ゆうかんな ゆうしゃ エンド！",
      gameover: "💀 バッドエンド"
    },

    fallbackEmoji: {
      great: "🏆",
      success: "✨",
      gameover: "💀"
    },

    shortLabel: {
      great: "でんせつの ゆうしゃ",
      success: "ゆうかんな ゆうしゃ",
      gameover: "バッドエンド"
    },

    fallbackTitle: "ぼうけんは おしまい！",
    fallbackStory:
      "つうしんに もんだいが あるようです。ちかくの スタッフに きいてね。",
    netErrorTitle: "ぼうけんは おしまい！",
    netErrorStory:
      "つうしんに もんだいが あるようです。ちかくの スタッフに きいてね。",

    emailLabel: "メールで おくる",
    emailPlaceholder: "your@email.com",
    emailBtn: "メールで おくる",
    emailSubject: "ドラゴン城の ひほう：あなたの ぼうけんの きろく",
    emailBody: "ゆうしゃの ぼうけんの きろくを おくるよ！",
    adventureHeader: "— これまでの ぼうけん —",

    clearedLabel: "けっか：クリア！",
    failedLabel: "けっか：しっぱい",
    timeoutLabel: "じかんぎれ！",
    restart: "もういちど あそぶ",
    loading: "よみこんでいます…",
    notFound: "ぼうけんの きろくが みつかりません"
  }
};

const en: Messages = {
  lang: "en",
  langName: "English",
  intro: {
    title: "🐉 Dragon Castle's Secret",
    subtitle: "— AI Gamebook —",
    lines: [
      "You are a hero on your way to Dragon Castle to save the princess.",
      "You can do anything you can imagine!",
      "Say what your hero does! (For example: Move! / Make a wish! / Sneak past it!)"],
    hint: "You have 3 hearts. Don't lose them all!",
    start: "Start the adventure",
    practice: "Practice",
    pickLang: "Choose language"
  },
  practice: {
    title: "Practice",
    situation:
      "While walking, you find a lost bag that seems to belong to Alice. You can see a police box in the distance. What do you do?",
    goal: "What should your hero do? Say it!",
    micLabel: "Click to speak",
    manualSummary: "If voice input doesn't work, type your answer.",
    manualPlaceholder: "What should your hero do?",
    manualBtn: "Send",
    judging: "Checking your answer…",
    success: "Success! You turned in the bag at the police box!",
    failure: "Oh no! You walked past the bag without helping…",
    back: "Back to the beginning"
  },
  hud: {
    livesEmpty: "No hearts left"
  },
  stage: {
    prefix: (n) => `Stage ${n}: `,
    restart: "Restart from the beginning",
    restartConfirm: "Are you sure you want to start over?",
    stages: [
      {
        title: "The Castle Across the River",
        situation:
          "Dragon Castle is visible in the distance, but a river blocks your path. You can't cross it!",
        goal: "How should your hero cross the river? Say it!"
      },
      {
        title: "The Golem at the Castle Gate",
        situation:
          "You reach Dragon Castle, but a huge, terrifying golem blocks the entrance!",
        goal: "The golem blocks the way! What should your hero do? Say it!"
      },
      {
        title: "The Fire Trap",
        situation:
          "Inside the castle, fierce flames block the hallway. You can't move forward!",
        goal: "Flames block the path! What should your hero do? Say it!"
      },
      {
        title: "The Final Battle with the Dragon",
        situation:
          "On the top floor, you find the princess—but an angry dragon attacks!",
        goal:
          "The dragon attacks! Will your hero fight it, befriend it, or try something else? Say it!"
      }
    ]
  },

  input: {
    micLabel: "Click to speak",
    listening: "Listening…",
    judging: "Judging…",
    ended: "Finished",
    micUnsupported: "Please type your answer instead.",
    manualSummary: "If voice input doesn't work, type your answer.",
    manualPlaceholder: "What should your hero do?",
    manualBtn: "Send"
  },

  judge: {
    bad: "Not quite! Think again and try once more. 💡",
    goodSuffix: "(-1 heart) On to the next stage!",
    greatSuffix: "✨ On to the next stage!",
    noVoice: "I couldn't hear you. Please try speaking again. 🎤",
    netError: "There was a connection problem. Please try again.",
    micDenied:
      "Microphone access is turned off. Please use the text box below.",
    noSpeech: "I didn't hear anything. Please try again.",
    unsupported:
      "Voice input isn't available in this browser. Please type your answer.",
    generic: "Something went wrong. Please try again.",
    noReason: "You gave it a try!",
    lifeDown: "-1 heart",
    lifeNone: "No hearts lost!",
    historyTitle: "📜 Your adventure so far",
    historyEmpty: "Your adventure hasn't started yet!"
  },

  ending: {
    titles: {
      great: "🏆 Legendary Hero Ending!",
      success: "✨ Brave Hero Ending!",
      gameover: "💀 Bad Ending"
    },

    fallbackEmoji: {
      great: "🏆",
      success: "✨",
      gameover: "💀"
    },

    shortLabel: {
      great: "Legendary Hero",
      success: "Brave Hero",
      gameover: "Bad Ending"
    },

    fallbackTitle: "Adventure Over!",
    fallbackStory:
      "There seems to be a connection problem. Please ask a staff member for help.",
    netErrorTitle: "Adventure Over!",
    netErrorStory:
      "There seems to be a connection problem. Please ask a staff member for help.",

    emailLabel: "Send by email",
    emailPlaceholder: "your@email.com",
    emailBtn: "Send by email",
    emailSubject: "Dragon Castle: Your Adventure Record",
    emailBody: "Here is your hero's adventure record!",
    adventureHeader: "— Your Adventure —",

    clearedLabel: "Result: Cleared!",
    failedLabel: "Result: Failed",
    timeoutLabel: "Time's up!",
    restart: "Play again",
    loading: "Loading…",
    notFound: "Adventure record not found"
  }
};

export const dictionaries: Record<Lang, Messages> = { ja, en };

export function getMessages(lang: Lang): Messages {
  return dictionaries[lang] ?? dictionaries.ja;
}

export function isLang(v: string): v is Lang {
  return v === "ja" || v === "en";
}
