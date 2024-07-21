import sys
import fitz  # PyMuPDF
from PIL import Image
import os

def pdf_to_jpeg(pdf_path, output_dir):
    pdf_document = fitz.open(pdf_path)

    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    for page_num in range(pdf_document.page_count):
        page = pdf_document.load_page(page_num)
        pix = page.get_pixmap()

        img = Image.frombytes("RGB", [pix.width, pix.height], pix.samples)

        output_path = os.path.join(output_dir, f"{page_num + 1}.jpg")
        img.save(output_path, "JPEG")


args = sys.argv

pdf_path = args[1]
output_dir = args[2]
pdf_to_jpeg(pdf_path, output_dir)

print("Done")
