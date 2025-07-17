import os
import subprocess

def main():
    grammars_dir = 'grammars'
    output_dir = 'pkg/uast/uastmaps'
    os.makedirs(output_dir, exist_ok=True)
    for lang in os.listdir(grammars_dir):
        lang_dir = os.path.join(grammars_dir, lang)
        if not os.path.isdir(lang_dir):
            continue
        for root, dirs, files in os.walk(lang_dir):
            for fname in files:
                if fname == 'node-types.json':
                    node_types_path = os.path.join(root, fname)
                    out_path = os.path.join(output_dir, f'{lang}.uastmap')
                    print(f'Generating {out_path} from {node_types_path}...')
                    cmd = [
                        'uast', 'mapping', '--generate',
                        '--node-types', node_types_path,
                        '--format', 'text'
                    ]
                    with open(out_path, 'w') as outf:
                        subprocess.run(cmd, stdout=outf, check=True)

if __name__ == '__main__':
    main() 