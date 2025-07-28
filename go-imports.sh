alias go-imports="uast parse \$1 | uast query \"rfilter(.type == \\\"Import\\\") |> rmap(.token)\" | jq -r '.results[] | select(. != \"\") | gsub(\"\\\"\"; \"\")' | sort | uniq"
