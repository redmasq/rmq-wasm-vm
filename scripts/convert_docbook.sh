#!/bin/bash
# zsh compatible

DOCBOOK_INPUT="docs/specs/wasm_vm_srs.docbook"
OUTPUT_DIR="docs/output"
OUTPUT_HTML="$OUTPUT_DIR/wasm_vm_srs.html"

# Create output directory if needed
mkdir -p "$OUTPUT_DIR"

convert_with_tool() {
    local tool=$1
    case "$tool" in
        xsltproc)
            XSL_PATHS=(
            "/usr/share/xml/docbook/stylesheet/docbook-xsl/html/docbook.xsl"
            "/usr/share/sgml/docbook/xsl-stylesheets/html/docbook.xsl"
            "/usr/local/opt/docbook-xsl/docbook-xsl/html/docbook.xsl"
            )
            for xsl in "${XSL_PATHS[@]}"; do
            if [ -f "$xsl" ]; then
                XSL_PATH="$xsl"
                break
            fi
            done
            if [ ! -f "$XSL_PATH" ]; then
                echo "DocBook XSL stylesheets not found. Please install them (see your distro's docs)."
                exit 1
            fi
            xsltproc -o "$OUTPUT_HTML" "$XSL_PATH" "$DOCBOOK_INPUT"
            ;;
        dbtohtml)
            dbtohtml -o "$OUTPUT_HTML" "$DOCBOOK_INPUT"
            ;;
        docbook2html)
            docbook2html -o "$OUTPUT_HTML" "$DOCBOOK_INPUT"
            ;;
        pandoc)
            pandoc -f docbook -t html -o "$OUTPUT_HTML" "$DOCBOOK_INPUT"
            ;;
    esac
}

# 1. Try to find a conversion tool
for tool in xsltproc dbtohtml docbook2html pandoc; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo "Using $tool to convert DocBook to HTML..."
        convert_with_tool "$tool"
        if command -v "tidy" >/dev/null 2>&1; then
            mv "$OUTPUT_HTML" "$OUTPUT_HTML.bak"
            tidy -i -q -wrap 0 -o "$OUTPUT_HTML" "$OUTPUT_HTML.bak"
        fi
        echo "Conversion complete: $OUTPUT_HTML"
        exit 0
    fi
done

# 2. Not found, suggest how to install
echo "No DocBook to HTML conversion tool found on your system."
# Detect package manager
if command -v brew >/dev/null 2>&1; then
    echo "Try: brew install docbook docbook-xsl xsltproc pandoc"
elif command -v apt >/dev/null 2>&1; then
    echo "Try: sudo apt update && sudo apt install docbook-xsl xsltproc pandoc"
elif command -v pacman >/dev/null 2>&1; then
    echo "Try: sudo pacman -S docbook-xsl xsltproc pandoc"
elif command -v yum >/dev/null 2>&1; then
    echo "Try: sudo yum install docbook-style-xsl xsltproc pandoc"
else
    echo "Please install xsltproc, docbook-xsl, or pandoc, and add them to your PATH. See your OS documentation."
fi


(return 0 2>/dev/null) && SOURCED=1 || SOURCED=0
if [ "${SOURCED:-0}" -eq 0 ]; then
  exit 1
else
  return 1
fi