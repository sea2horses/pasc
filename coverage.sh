
#!/usr/bin/env bash

pascc="pascc"
build_args=(build --temp)

mapfile -d '' -t files < <(find . -type f -name "*.psc" -print0)

total=${#files[@]}
successful=0

for file in "${files[@]}"; do
    if output=$("$pascc" "${build_args[@]}" "$file" 2>&1); then
        ((successful += 1))
        printf '✓ %s (%d/%d)\n' "$file" "$successful" "$total"
    else
        printf '! Failed to build file %s (%d/%d)\n' \
            "$file" "$successful" "$total"
        printf '%s\n' "$output"
    fi
done

if (( total > 0 )); then
    coverage=$((successful * 100 / total))
else
    coverage=0
fi

printf 'Total coverage: %d%% (%d/%d)\n' \
    "$coverage" "$successful" "$total"
