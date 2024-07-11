import pypdf
import sys

def split_pdf_pages(src_path, dst_basepath, ):
    src_pdf = pypdf.PdfReader(src_path)
    for i, page in enumerate(src_pdf.pages):
        dst_pdf = pypdf.PdfWriter()
        dst_pdf.add_page(page)
        dst_pdf.write(f'{dst_basepath}{i+1}.pdf')

args = sys.argv

split_pdf_pages(args[1], 'cut/')
