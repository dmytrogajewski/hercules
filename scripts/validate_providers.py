#!/usr/bin/env python3
"""
Validate UAST provider YAML files against Tree-sitter grammar JSON files.
"""

import json
import yaml
import os
import sys
from pathlib import Path
import glob

def extract_grammar_nodes(grammar_file):
    """Extract all node names from a Tree-sitter grammar JSON file."""
    try:
        with open(grammar_file, 'r') as f:
            grammar = json.load(f)
        
        if 'rules' not in grammar:
            return set()
        
        return set(grammar['rules'].keys())
    except Exception as e:
        print(f"Error reading {grammar_file}: {e}")
        return set()

def extract_provider_nodes(provider_file):
    """Extract all node names from a UAST provider YAML file."""
    try:
        with open(provider_file, 'r') as f:
            provider = yaml.safe_load(f)
        
        if not provider or 'mapping' not in provider:
            return set()
        
        return set(provider['mapping'].keys())
    except Exception as e:
        print(f"Error reading {provider_file}: {e}")
        return set()

def validate_provider(provider_file, grammar_file):
    """Validate a provider YAML file against its grammar JSON file."""
    provider_name = os.path.basename(provider_file).replace('.yaml', '')
    
    if not os.path.exists(grammar_file):
        print(f"❌ {provider_name}: Grammar file not found: {grammar_file}")
        return False
    
    grammar_nodes = extract_grammar_nodes(grammar_file)
    provider_nodes = extract_provider_nodes(provider_file)
    
    if not provider_nodes:
        print(f"❌ {provider_name}: No nodes found in provider file")
        return False
    
    # Check for nodes in provider that don't exist in grammar
    invalid_nodes = provider_nodes - grammar_nodes
    if invalid_nodes:
        # Filter out None values and sort
        valid_invalid_nodes = [node for node in invalid_nodes if node is not None]
        if valid_invalid_nodes:
            print(f"❌ {provider_name}: Invalid nodes in provider: {sorted(valid_invalid_nodes)}")
            return False
    
    # Check for common nodes that should be mapped
    common_nodes = provider_nodes & grammar_nodes
    if not common_nodes:
        print(f"⚠️  {provider_name}: No common nodes found between provider and grammar")
        return False
    
    print(f"✅ {provider_name}: Valid ({len(common_nodes)} nodes mapped)")
    return True

def find_grammar_file(provider_name):
    """Find the grammar JSON file for a given provider name."""
    # 1. Try grammars_json/{name}_grammar.json
    candidate = Path("grammars_json") / f"{provider_name}_grammar.json"
    if candidate.exists():
        return candidate
    # 2. Try grammars/{name}/src/grammar.json
    candidate = Path(f"grammars/{provider_name}/src/grammar.json")
    if candidate.exists():
        return candidate
    # 3. Try grammars/{name}/grammar.json
    candidate = Path(f"grammars/{provider_name}/grammar.json")
    if candidate.exists():
        return candidate
    # 4. Try grammars/{name}/tree-sitter-{name}/grammar.json
    candidate = Path(f"grammars/{provider_name}/tree-sitter-{provider_name}/grammar.json")
    if candidate.exists():
        return candidate
    # 5. Try grammars/{name}/tree-sitter-{name}/src/grammar.json
    candidate = Path(f"grammars/{provider_name}/tree-sitter-{provider_name}/src/grammar.json")
    if candidate.exists():
        return candidate
    # 6. Try grammars/{name}/grammars/{name}/src/grammar.json (for ocaml)
    candidate = Path(f"grammars/{provider_name}/grammars/{provider_name}/src/grammar.json")
    if candidate.exists():
        return candidate
    # 7. Try glob for grammar.json in grammars/{name}/**/
    matches = glob.glob(f"grammars/{provider_name}/**/grammar.json", recursive=True)
    if matches:
        return Path(matches[0])
    return None

def main():
    """Main validation function."""
    providers_dir = Path("pkg/uast/providers")
    
    if not providers_dir.exists():
        print(f"Error: Providers directory not found: {providers_dir}")
        sys.exit(1)
    
    # Get all provider YAMLs
    providers_to_validate = [p.stem for p in providers_dir.glob("*.yaml")]
    
    print("Validating UAST providers against Tree-sitter grammars...")
    print("=" * 60)
    
    valid_count = 0
    total_count = len(providers_to_validate)
    
    for provider_name in providers_to_validate:
        provider_file = providers_dir / f"{provider_name}.yaml"
        grammar_file = find_grammar_file(provider_name)
        if not grammar_file:
            print(f"❌ {provider_name}: Grammar file not found for provider.")
            continue
        if validate_provider(provider_file, grammar_file):
            valid_count += 1
    
    print("=" * 60)
    print(f"Validation complete: {valid_count}/{total_count} providers valid")
    
    if valid_count == total_count:
        print("🎉 All providers are valid!")
        sys.exit(0)
    else:
        print("❌ Some providers have issues that need to be fixed.")
        sys.exit(1)

if __name__ == "__main__":
    main() 