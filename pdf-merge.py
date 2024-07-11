import os
import sys
from pypdf import PdfMerger

def merge_pdfs(max, output_path, directory):
    pdf_merger = PdfMerger()

    for i in range(1, max + 1):
        pdf_path = os.path.join(directory, f"{i}.pdf")
        if os.path.exists(pdf_path):
            pdf_merger.append(pdf_path)
        else:
            print(f"File {pdf_path} not found. Skipping...")

    with open(output_path, 'wb') as output_file:
        pdf_merger.write(output_file)

    print(f"Merged file saved as {output_path}")

max = int(sys.argv[1])

merge_pdfs(max, "merge/merge.pdf", "cut")
