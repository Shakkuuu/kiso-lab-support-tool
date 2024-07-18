import os
import sys
from pypdf import PdfWriter

def merge_pdfs(max, output_path, directory):
    writer = PdfWriter()

    for i in range(1, max + 1):
        pdf_path = os.path.join(directory, f"{i}.pdf")
        if os.path.exists(pdf_path):
            writer.append(pdf_path)
        else:
            print(f"File {pdf_path} not found. Skipping...")

    with open(output_path, 'wb') as output_file:
        writer.write(output_file)

args = sys.argv

max = int(args[1])

merge_pdfs(max, args[2], args[3])

print("Done")
