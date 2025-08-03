# Security Policy

## Supported Versions

Right now, the project is in its early stages, so only the head of main is actually supported.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| previous| :x:                |

## Reporting a Vulnerability

For now, please report issues to the [GitHub Issues Page](https://github.com/redmasq/rmq-wasm-vm/issues).

1. Describe the issue, including reproduction sets, what was expected, and what actually happened. Also mention if it happens every time.
2. Make sure to include the commit id, which is a value that looks like 7f43f833840f341545b003d4ca92909e33d187e6, or in its short form 7f43f83.

At this time, I do not have a specific service level for triage or resolving issues. I will comment and close any issues that are resolved.

If you do not know how to find this information, you can do one of the following
1. From the GitHub UI [rmq-wasm-vm main](https://github.com/redmasq/rmq-wasm-vm/tree/main), look for a banner that says the number of commits, the aforementioned value should be a little to the left of it.
2. From you Git Bash, WSL, Powershell, zsh, terminal, or similar prompt, run the following

```
    git show
```
It should reveal something that looks like the following (only need the first line that starts with commit)
```
commit c668705d0b71595f00df2e84331d3a0ba7e4b3ab (HEAD -> main, origin/main, origin/HEAD)
Author: redmasq <redmasq@users.noreply.github.com>
Date:   Sat Aug 2 17:20:05 2025 -0500

    Adjusted tooling again for troubleshooting. Didn't bother with pull request.

diff --git a/.github/workflows/go.yml b/.github/workflows/go.yml
index 528b2b4..2ffe0b2 100644
--- a/.github/workflows/go.yml
+++ b/.github/workflows/go.yml
@@ -45,9 +45,8 @@ jobs:
         awk -v coverage="$coverage" -v threshold="$threshold" 'BEGIN {if (coverage+0 < threshold) exit 1}'
     - name: Convert coverage.out to lcov
       run: |
-        go install github.com/AlekSi/gocov-xml@latest
-        go install github.com/axw/gocov/gocov@latest
-        gocov test ./... | gocov-xml > coverage.lcov
+        go install github.com/jandelgado/gcov2lcov@latest
+        gcov2lcov -infile coverage.out -outfile coverage.lcov

     - name: Upload coverage to Coveralls
       uses: coverallsapp/github-action@v2
```
