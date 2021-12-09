package xtremio

const authenticateResponse = `
{
	"loginProviderName": "tmos",
	"token": {
		"token": "FUL2V33SRR2JBF4NKKABCDEFGH",
		"name": "FUL2V33SRR2JBF4NKKABCDEFGH"
	}
}
`

const sampleBBUResponseOne = `
{
    "content": {
        "is-low-battery-has-input": "false",
        "serial-number": "A123B45678",
        "guid": "987654321abcdef",
        "brick-name": "X1",
        "ups-battery-charge-in-percent": 100,
        "power": 244,
        "avg-daily-temp": 23,
        "fw-version": "01.02.0034",
        "sys-name": "ABCXIO001",
		"power-feed": "PWR-A",
        "ups-load-in-percent": 21,
        "name": "X1-BBU",
		"enabled-state": "enabled",
        "is-low-battery-no-input": "false",
        "ups-need-battery-replacement": "false",
        "model-name": "Eaton Model Name",
    }
}
`

const sampleGetBBUsResponse = `
{
    "bbus": [
        {
            "href": "https://127.0.0.1/api/json/v3/types/bbus/987654321abcdef", 
            "name": "X1-BBU", 
            "sys-name": "ABCXIO001"
        }
    ], 
    "links": [
        {
            "href": "https://127.0.0.1/api/json/v3/types/bbus/", 
            "rel": "self"
        }
    ]
}
`