import os
import json
import subprocess
import shutil

GO_SITTER_FOREST_REPO = "https://github.com/alexaandru/go-sitter-forest.git"
GO_SITTER_FOREST_DIR = "go-sitter-forest"
GRAMMARS_DIR = "grammars"
GRAMMARS_JSON_DIR = "grammars_json"

def run(cmd, cwd=None):
    print(f"Running: {cmd}")
    subprocess.run(cmd, shell=True, check=True, cwd=cwd)

def ensure_repo_cloned(repo_url, target_dir):
    if not os.path.exists(target_dir):
        run(f"git clone {repo_url} {target_dir}")
    else:
        print(f"Repo already cloned: {target_dir}, skipping clone.")
        # Optionally, could pull here if you want updates
        # run("git pull", cwd=target_dir)

def main():
    os.makedirs(GRAMMARS_JSON_DIR, exist_ok=True)
    # 1. Clone or update go-sitter-forest
    ensure_repo_cloned(GO_SITTER_FOREST_REPO, GO_SITTER_FOREST_DIR)

    # 2. For each folder, read grammar.json and extract "url"
    for entry in os.listdir(GO_SITTER_FOREST_DIR):
        try:
            lang_dir = os.path.join(GO_SITTER_FOREST_DIR, entry)
            grammar_json_path = os.path.join(lang_dir, "grammar.json")
            if not os.path.isdir(lang_dir) or not os.path.isfile(grammar_json_path):
                continue
            with open(grammar_json_path) as f:
                grammar = json.load(f)
            url = grammar.get("url")
            if not url:
                print(f"No url in {grammar_json_path}")
                continue
            # 3. Clone the grammar repo into grammars/ if not already present
            target_dir = os.path.join(GRAMMARS_DIR, entry)
            try:
                ensure_repo_cloned(url, target_dir)
            except Exception as e:
                print(f"ERROR: Failed to clone {url} for {entry}: {e}")
                continue
            print(f"Pulled grammar for {entry} into {target_dir}")
            # 4. Copy grammar.json from /src in the grammar repo, name as <language>_grammar.json in grammars_json/
            src_grammar_json = os.path.join(target_dir, "src", "grammar.json")
            dest_grammar_json = os.path.join(GRAMMARS_JSON_DIR, f"{entry}_grammar.json")
            if os.path.isfile(src_grammar_json):
                shutil.copyfile(src_grammar_json, dest_grammar_json)
                print(f"Copied src/grammar.json for {entry} to {dest_grammar_json}")
            else:
                print(f"WARNING: src/grammar.json not found in {target_dir}")
        except Exception as e:
            print(f"ERROR: Unexpected error for {entry}: {e}")
            continue

if __name__ == "__main__":
    os.makedirs(GRAMMARS_DIR, exist_ok=True)
    main() 