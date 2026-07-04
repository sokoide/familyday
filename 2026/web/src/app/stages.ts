// app: ステージ識別子と順序。表示テキスト(多言語)は messages.ts 側が持つ。
// サーバ側の成功条件は domain/catalog.go が別途持つ。
export type StageId = "stage1" | "stage2" | "stage3" | "stage4";

export interface StageRef {
  id: StageId;
  number: number; // 表示番号(1始まり)
}

export const STAGES: StageRef[] = [
  { id: "stage1", number: 1 },
  { id: "stage2", number: 2 },
  { id: "stage3", number: 3 },
  { id: "stage4", number: 4 },
];
