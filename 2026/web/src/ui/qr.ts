// ui: QRコード描画(qrcode ライブラリ)
import QRCode from "qrcode";

export async function renderQR(canvas: HTMLCanvasElement, text: string): Promise<void> {
  await QRCode.toCanvas(canvas, text, { width: 220, margin: 1 });
}
