$contentType = "application/json"

$nav_uri = "https://localhost:8446/client/ba14nc21096/navigate"


$nav_json = @{
    'url' = "https://www.unibas.ch";
    'nextstatus' = "web";
    'waitfor' = "a.button.wide.cookie-banner-ok";
    'waittimeout' = "3s";
    'element' = "a.button.wide.cookie-banner-ok"; # "button#uc-btn-accept-banner";
};


[System.Net.ServicePointManager]::ServerCertificateValidationCallback = { $true }

Invoke-RestMethod -Method POST -Uri $nav_uri -ContentType $contentType -Body (ConvertTo-Json $nav_json)