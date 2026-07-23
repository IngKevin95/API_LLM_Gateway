import os
import glob

docs_dir = r"E:\Datos\Documentos\GitHub Personal\API_LLM_Gateway\docs\04-historias"
output_file = r"E:\Datos\Documentos\GitHub Personal\API_LLM_Gateway\docs\04-historias\scratch_all_stories.txt"

files = glob.glob(os.path.join(docs_dir, "HU-*.md"))
files.sort()

with open(output_file, "w", encoding="utf-8") as out:
    for f in files:
        out.write(f"\n--- {os.path.basename(f)} ---\n")
        with open(f, "r", encoding="utf-8") as inp:
            out.write(inp.read())
