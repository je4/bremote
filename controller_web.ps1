param([string]$machine, [string]$jsonfile)

$contentType = "application/json"
$slideweb_uri = "https://localhost:8444/kvstore/$($machine)/slide_web.gohtml"
$pdf_uri = "https://localhost:8444/kvstore/$($machine)/pdf.gohtml"
$autonav_uri = "https://localhost:8444/kvstore/$($machine)/autonav"
$nav_uri = "https://localhost:8444/client/$($machine)/navigate"

$kv_json = (Get-Content $jsonfile | Out-String | ConvertFrom-Json)

$nav_json = @{
    'url' = "https://localhost:8443/controller01/templates/slide_web.gohtml";
    'nextstatus' = "slideweb";
};

$pdf_json = @{
    'BackgroundColor' = "black";
};


[System.Net.ServicePointManager]::ServerCertificateValidationCallback = { $true }

Invoke-RestMethod -Method PUT -Uri $slideweb_uri -ContentType $contentType -Body (ConvertTo-Json $kv_json)

Invoke-RestMethod -Method POST -Uri $autonav_uri -ContentType $contentType -Body (ConvertTo-Json $nav_json)
Invoke-RestMethod -Method POST -Uri $pdf_uri -ContentType $contentType -Body (ConvertTo-Json $pdf_json)

Invoke-RestMethod -Method POST -Uri $nav_uri -ContentType $contentType -Body (ConvertTo-Json $nav_json)