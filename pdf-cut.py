import sys
import fitz  # PyMuPDF
from PIL import Image
import os

# PDFファイルを1ページごとに分割してjpgに変換
def pdf_to_jpeg(pdf_path, output_dir):
    # 分割すファイルを開く
    pdf_document = fitz.open(pdf_path)

    # 分割した資料をアップロードするディレクトリが存在する確認
    if not os.path.exists(output_dir):
        # なかったら作成
        os.makedirs(output_dir)

    # ページごとに処理
    for page_num in range(pdf_document.page_count):
        # ページ読み込み
        page = pdf_document.load_page(page_num)
        pix = page.get_pixmap()

        # 画像に変換
        img = Image.frombytes("RGB", [pix.width, pix.height], pix.samples)

        # JPEGとして保存
        output_path = os.path.join(output_dir, f"{page_num + 1}.jpg")
        img.save(output_path, "JPEG")

# 引数取得
args = sys.argv

pdf_path = args[1]  # 分割する資料のパス
output_dir = args[2]  # 分割した資料をアップロードするディレクトリのパス
pdf_to_jpeg(pdf_path, output_dir)

# 正常に完了したらプリント
print("Done")
