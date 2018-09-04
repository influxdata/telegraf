<#

.SYNOPSIS

Generate data contracts for Application Insights Go SDK from bond schema.

.DESCRIPTION

This is a convenience tool for generating the AI data contracts from the latest
bond schema.  It requires BondSchemaGenerator.exe in order to operate.  It also
requires the latest bond schema, but will check it out from github if it is not
present in the current directory.

.PARAMETER BondSchemaGenerator

The full path to BondSchemaGenerator.exe

.PARAMETER SchemasDir

The path to the directory that contains all of the input .bond files.

.LINK https://github.com/Microsoft/ApplicationInsights-Home

#>

[cmdletbinding()]
Param(
    [Parameter(Mandatory=$true)]
    [string] $BondSchemaGenerator,
    [string] $SchemasDir
)

function RunBondSchemaGenerator
{
    [cmdletbinding()]
    Param(
        [string] $Language,
        [string[]] $Files,
        [string] $Layout,
        [string[]] $Omissions
    )

    $args = @("-v")
    $args += @("-o", ".")
    $args += @("-e", $Language)
    $args += @("-t", $Layout)

    foreach ($file in $Files) {
        $args += @("-i", $file)
    }

    foreach ($omission in $Omissions) {
        $args += @("--omit", $omission)
    }

    & "$BondSchemaGenerator" $args 2>&1
}

$origpath = Get-Location

try {
    $scriptpath = $MyInvocation.MyCommand.Path
    $dir = Split-Path $scriptpath
    cd $dir

    if (-not (Test-Path $BondSchemaGenerator -PathType Leaf)) {
        Write-Host "Could not find BondSchemaGenerator at $BondSchemaGenerator"
        Write-Host "Please specify the full path"
        Exit 1
    }

    if (-not $schemasDir) {
        $schemasDir = ".\ApplicationInsights-Home\EndpointSpecs\Schemas\Bond"

        # Check locally.
        if (-not (Test-Path .\ApplicationInsights-Home -PathType Container)) {
            # Clone into it!
            git clone https://github.com/Microsoft/ApplicationInsights-Home.git
        }
    }

    $files = Get-ChildItem $schemasDir | % { "$schemasDir\$_" }
    $omissions = @("Microsoft.Telemetry.Domain", "Microsoft.Telemetry.Base", "AI.AjaxCallData", "AI.PageViewPerfData")

    RunBondSchemaGenerator -Files $files -Language GoBondTemplateLanguage -Layout GoTemplateLayout -Omissions $omissions
    RunBondSchemaGenerator -Files $files -Language GoContextTagsLanguage -Layout GoTemplateLayout -Omissions $omissions

    cd appinsights\contracts
    go fmt
} finally {
    cd $origpath
}
