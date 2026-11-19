param(
    [ValidateSet('up', 'down', 'ps', 'logs', 'stop-payment', 'start-payment')]
    [string]$Action = 'up',
    [switch]$LocalGo
)

& (Join-Path (Split-Path $PSScriptRoot -Parent) 'compose-lab.ps1') -Lab part1 -Action $Action -LocalGo:$LocalGo
