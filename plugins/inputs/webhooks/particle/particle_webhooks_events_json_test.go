package particle

func NewItemJSON() string {
	return `
	{ 
	  "event": "temperature",
	  "data": "{ 
		  "tags": {
			  "id": "230035001147343438323536", 
			  "location\": \"TravelingWilbury"
		  }, 
		  "values": {
			  "temp_c": 26.680000, 
			  "temp_f": 80.024001, 
			  "humidity": 44.937500, 
			  "pressure": 998.998901, 
			  "altitude": 119.331436, 
			  "broadband": 1266, 
			  "infrared": 528, 
			  "lux": 0
		  }
	  }",
	  "ttl": 60,
	  "published_at": "2017-09-28T21:54:10.897Z",
	  "coreid": "123456789938323536",
	  "userid": "1234ee123ac8e5ec1231a123d",
	  "version": 10,
	  "public": false,
	  "productID": 1234,
	  "name": "sensor"
  }`
}
func UnknowJSON() string {
	return `
    {
      "event": "roger"
    }`
}
