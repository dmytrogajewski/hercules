#!/usr/bin/env python3
"""
Enhanced UAST Map Generation Script

This script generates UAST maps for all languages in the grammars directory,
ensuring all maps include proper language section headers with correct
language names and file extensions.

Uses Pygments for better language detection and provides comprehensive
error handling and reporting.
"""

import os
import subprocess
import json
from pathlib import Path

def get_language_info(lang):
    """Get language name, extensions, and files for a given language directory"""
    # Raw mapping: can contain both extensions and files in the same list
    raw_map = {
        'go': ('go', ['.go']),
        'javascript': ('javascript', ['.js', '.jsx', '.mjs']),
        'typescript': ('typescript', ['.ts', '.tsx']),
        'tsx': ('tsx', ['.tsx']),
        'python': ('python', ['.py', '.pyw', '.pyi']),
        'java': ('java', ['.java']),
        'c': ('c', ['.c', '.h']),
        'cpp': ('cpp', ['.cpp', '.cc', '.cxx', '.hpp', '.hxx']),
        'rust': ('rust', ['.rs']),
        'rust_with_rstml': ('rust_with_rstml', ['.rs']),
        'ruby': ('ruby', ['.rb', '.rbw']),
        'php': ('php', ['.php', '.phtml']),
        'c_sharp': ('csharp', ['.cs']),
        'kotlin': ('kotlin', ['.kt', '.kts']),
        'swift': ('swift', ['.swift']),
        'scala': ('scala', ['.scala']),
        'haskell': ('haskell', ['.hs', '.lhs']),
        'ocaml': ('ocaml', ['.ml', '.mli']),
        'fsharp': ('fsharp', ['.fs', '.fsx']),
        'clojure': ('clojure', ['.clj', '.cljs']),
        'erlang': ('erlang', ['.erl']),
        'elixir': ('elixir', ['.ex', '.exs']),
        'elm': ('elm', ['.elm']),
        'bash': ('bash', ['.sh', '.bash']),
        'fish': ('fish', ['.fish']),
        'sql': ('sql', ['.sql']),
        'html': ('html', ['.html', '.htm']),
        'css': ('css', ['.css']),
        'xml': ('xml', ['.xml']),
        'json': ('json', ['.json']),
        'yaml': ('yaml', ['.yaml', '.yml']),
        'toml': ('toml', ['.toml']),
        'ini': ('ini', ['.ini']),
        'markdown': ('markdown', ['.md', '.markdown']),
        'markdown_inline': ('markdown_inline', ['.md', '.markdown']),
        'dockerfile': ('dockerfile', ['.dockerfile', 'Dockerfile']),
        'make': ('make', ['.makefile', 'Makefile']),
        'lua': ('lua', ['.lua']),
        'perl': ('perl', ['.pl', '.pm']),
        'r': ('r', ['.r', '.R']),
        'dart': ('dart', ['.dart']),
        'crystal': ('crystal', ['.cr']),
        'nim': ('nim', ['.nim']),
        'nim_format_string': ('nim_format_string', ['.nim']),
        'fortran': ('fortran', ['.f', '.f90', '.f95']),
        'commonlisp': ('commonlisp', ['.lisp', '.lsp']),
        'cmake': ('cmake', ['.cmake', 'CMakeLists.txt']),
        'tcl': ('tcl', ['.tcl']),
        'hcl': ('hcl', ['.hcl', '.tf']),
        # Template languages
        'gotmpl': ('gotmpl', ['.gotmpl', '.go.tmpl']),
        'helm': ('helm', ['.yaml', '.yml']),
        # Configuration files
        'git_config': ('git_config', ['.gitconfig']),
        'gitignore': ('gitignore', ['.gitignore']),
        'gitattributes': ('gitattributes', ['.gitattributes']),
        # Removed git_rebase and gitcommit as they have no extensions or files
        'ssh_config': ('ssh_config', ['ssh_config']),
        'dotenv': ('dotenv', ['.env']),
        'properties': ('properties', ['.properties']),
        'gowork': ('gowork', ['go.work']),
        'gosum': ('gosum', ['go.sum']),
        # Data formats
        'csv': ('csv', ['.csv']),
        'psv': ('psv', ['.psv']),
        'proxima': ('proxima', ['.proxima']),
        'proto': ('proto', ['.proto']),
        'prql': ('prql', ['.prql']),
        # Documentation
        'latex': ('latex', ['.tex', '.ltx']),
        'ansible': ('ansible', ['.yml', '.yaml']),
    }
    if lang not in raw_map:
        return (lang, [], [])
    lang_name, all_patterns = raw_map[lang]
    extensions = [p for p in all_patterns if p.startswith('.')]
    files = [p for p in all_patterns if not p.startswith('.')]
    return (lang_name, extensions, files)

def ensure_language_section(content, lang_name, extensions, ff):
    """Ensure the UAST map has a proper language section header"""
    if not content.strip():
        return content
    
    # Remove any existing language section
    lines = content.split('\n')
    filtered_lines = []
    skip_until_empty = False
    
    for line in lines:
        if line.strip().startswith('[language '):
            skip_until_empty = True
            continue
        if skip_until_empty and line.strip() == '':
            skip_until_empty = False
            continue
        if skip_until_empty:
            continue
        filtered_lines.append(line)
    
    # Reconstruct content without language section
    content_without_header = '\n'.join(filtered_lines).strip()
    
    # Add proper language section at the beginning
    header_parts = [f'[language "{lang_name}"']
    
    if extensions:
        extensions_str = ', '.join(f'"{ext}"' for ext in extensions)
        header_parts.append(f'extensions: {extensions_str}')
    
    if ff:
        files_str = ', '.join(f'"{f}"' for f in ff)
        header_parts.append(f'files: {files_str}')
    
    header = ', '.join(header_parts) + ']\n\n'
    return header + content_without_header

def validate_uast_map(content, lang_name):
    """Validate that the generated UAST map is properly formatted"""
    if not content.strip():
        return False, "Empty content"
    
    # Check for language section
    if not content.startswith('[language '):
        return False, "Missing language section"
    
    # Check for basic UAST syntax
    if '=> uast(' not in content:
        return False, "No UAST mappings found"
    
    return True, "Valid"

def main():
    grammars_dir = 'grammars'
    output_dir = 'pkg/uast/uastmaps'
    os.makedirs(output_dir, exist_ok=True)
    
    # Track processing results
    processed = []
    skipped = []
    errors = []
    validation_errors = []
    
    for lang in sorted(os.listdir(grammars_dir)):
        lang_dir = os.path.join(grammars_dir, lang)
        if not os.path.isdir(lang_dir):
            continue
            
        # Get language info
        lang_name, extensions, ff = get_language_info(lang)
        
        # Skip languages with no extensions and no files
        if not extensions and not ff:
            print(f'Skipping {lang} - no extensions or files defined')
            skipped.append(lang)
            continue
        
        # Find node-types.json files
        node_types_found = False
        for root, dirs, files in os.walk(lang_dir):
            for fname in files:
                if fname == 'node-types.json':
                    node_types_path = os.path.join(root, fname)
                    out_path = os.path.join(output_dir, f'{lang}.uastmap')
                    
                    try:
                        print(f'Generating {out_path} from {node_types_path} for language {lang_name}...')
                        
                        # Create extensions string for command line
                        extensions_str = ','.join(extensions) if extensions else ''
                        
                        cmd = [
                            './uast', 'mapping', '--generate',
                            '--node-types', node_types_path,
                            '--language', lang_name,
                            '--format', 'text'
                        ]
                        
                        # Add extensions if available
                        if extensions_str:
                            cmd.extend(['--extensions', extensions_str])
                        
                        # Run the command and capture output
                        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
                        
                        # Ensure language section is present
                        content = ensure_language_section(result.stdout, lang_name, extensions, ff)
                        
                        # Validate the generated content
                        is_valid, validation_msg = validate_uast_map(content, lang_name)
                        if not is_valid:
                            validation_errors.append((lang, validation_msg))
                            print(f"WARNING: {lang} validation failed: {validation_msg}")
                        
                        # Write the content
                        with open(out_path, 'w') as outf:
                            outf.write(content)
                        
                        processed.append((lang, lang_name, len(extensions), len(ff)))
                        node_types_found = True
                        
                    except subprocess.CalledProcessError as e:
                        error_msg = f"Error generating {lang}: {e.stderr}"
                        print(f"ERROR: {error_msg}")
                        errors.append((lang, error_msg))
                    except Exception as e:
                        error_msg = f"Unexpected error for {lang}: {str(e)}"
                        print(f"ERROR: {error_msg}")
                        errors.append((lang, error_msg))
        
        if not node_types_found:
            skipped.append(lang)
    
    # Print comprehensive summary
    print(f"\n{'='*50}")
    print(f"UAST MAP GENERATION SUMMARY")
    print(f"{'='*50}")
    print(f"Successfully processed: {len(processed)} languages")
    print(f"Skipped (no node-types.json): {len(skipped)} languages")
    print(f"Errors: {len(errors)} languages")
    print(f"Validation warnings: {len(validation_errors)} languages")
    
    if processed:
        print(f"\nProcessed languages:")
        for lang, lang_name, ext_count, file_count in processed:
            print(f"  - {lang} -> {lang_name} ({ext_count} extensions, {file_count} files)")
    
    if skipped:
        print(f"\nSkipped languages:")
        for lang in skipped:
            print(f"  - {lang}")
    
    if errors:
        print(f"\nErrors:")
        for lang, error in errors:
            print(f"  - {lang}: {error}")
    
    if validation_errors:
        print(f"\nValidation warnings:")
        for lang, msg in validation_errors:
            print(f"  - {lang}: {msg}")
    
    # Check for missing language sections in existing files
    print(f"\nChecking existing UAST maps for missing language sections...")
    missing_sections = []
    for uastmap_file in os.listdir(output_dir):
        if uastmap_file.endswith('.uastmap'):
            file_path = os.path.join(output_dir, uastmap_file)
            with open(file_path, 'r') as f:
                content = f.read()
                if not content.startswith('[language '):
                    missing_sections.append(uastmap_file)
    
    if missing_sections:
        print(f"Files missing language sections: {len(missing_sections)}")
        for file in missing_sections:
            print(f"  - {file}")
    else:
        print("All existing UAST maps have language sections ✓")

if __name__ == '__main__':
    main() 