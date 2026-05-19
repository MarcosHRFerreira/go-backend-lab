param(
    [string]$InputMarkdown = "GO-TWEETS-APOSTILA-PROFISSIONAL-A4.md",
    [string]$OutputPdf = "GO-TWEETS-APOSTILA-PROFISSIONAL-A4.pdf",
    [string]$OutputHtml = "GO-TWEETS-APOSTILA-PROFISSIONAL-A4.html",
    [string]$DocumentTitle = "Go Tweets Apostila Profissional",
    [string]$FooterTitle = "Go Tweets | Apostila Profissional"
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "[apostila] $Message"
}

function Get-ToolPath {
    param([string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($null -eq $command) {
        return $null
    }
    return $command.Source
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$inputPath = Join-Path $scriptDir $InputMarkdown
$pdfPath = Join-Path $scriptDir $OutputPdf
$htmlPath = Join-Path $scriptDir $OutputHtml

if (-not (Test-Path $inputPath)) {
    throw "Arquivo de entrada nao encontrado: $inputPath"
}

$pandocPath = Get-ToolPath -Name "pandoc"
$wkhtmltopdfPath = Get-ToolPath -Name "wkhtmltopdf"
$weasyprintPath = Get-ToolPath -Name "weasyprint"

Write-Step "Entrada: $inputPath"

if ($null -eq $pandocPath) {
    Write-Step "Pandoc nao encontrado no PATH."
    Write-Step "Instale o Pandoc para converter Markdown em HTML/PDF."
    Write-Step "Download: https://pandoc.org/installing.html"
    Write-Step "Depois execute novamente este script."
    exit 1
}

Write-Step "Pandoc encontrado em: $pandocPath"
Write-Step "Gerando HTML intermediario..."

& $pandocPath `
    --standalone `
    --from markdown+raw_html+markdown_in_html_blocks `
    --to html5 `
    --metadata title=$DocumentTitle `
    --output $htmlPath `
    $inputPath

Write-Step "HTML gerado em: $htmlPath"

if ($null -ne $wkhtmltopdfPath) {
    Write-Step "wkhtmltopdf encontrado em: $wkhtmltopdfPath"
    Write-Step "Gerando PDF com wkhtmltopdf e numeracao de paginas..."
    & $wkhtmltopdfPath `
        --enable-local-file-access `
        --print-media-type `
        --margin-top 11mm `
        --margin-right 11mm `
        --margin-bottom 17mm `
        --margin-left 11mm `
        --footer-left $FooterTitle `
        --footer-right "[page] / [topage]" `
        --footer-font-size 9 `
        --footer-spacing 6 `
        --footer-line `
        $htmlPath `
        $pdfPath
    Write-Step "PDF gerado com sucesso em: $pdfPath"
    exit 0
}

if ($null -ne $weasyprintPath) {
    Write-Step "WeasyPrint encontrado em: $weasyprintPath"
    Write-Step "Gerando PDF com WeasyPrint..."
    & $weasyprintPath `
        $htmlPath `
        $pdfPath
    Write-Step "PDF gerado com sucesso em: $pdfPath"
    exit 0
}

Write-Step "Nenhum conversor PDF compativel foi encontrado no PATH."
Write-Step "Opcoes suportadas por este script:"
Write-Step "- wkhtmltopdf: https://wkhtmltopdf.org/downloads.html"
Write-Step "- WeasyPrint: https://weasyprint.org/"
Write-Step "O HTML intermediario ja foi gerado e pode ser aberto no navegador para impressao manual em PDF."
exit 0
