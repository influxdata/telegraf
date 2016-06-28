package filestack

func DialogOpenJSON() string {
	return `{
	  "action": "fp.dialog",
	  "timestamp": 1435584646,
	  "id": 102,
	  "text": {
	    "mimetypes": ["*/*"],
	    "iframe": false,
	    "language": "en",
	    "id": "1435584650723",
	    "mobile": false,
	    "app":{
	       "upsell": "false",
	       "apikey": "YOUR_API_KEY",
	       "customization":{
	          "saveas_subheader": "Save it down to your local device or onto the Cloud",
	          "folder_subheader": "Choose a folder to share with this application",
	          "open_subheader": "Choose from the files on your local device or the ones you have online",
	          "folder_header": "Select a folder",
	          "help_text": "",
	          "saveas_header": "Save your file",
	          "open_header": "Upload a file"
	       }
	    },
	    "dialogType": "open",
	    "auth": false,
	    "welcome_header": "Upload a file",
	    "welcome_subheader": "Choose from the files on your local device or the ones you have online",
	    "help_text": "",
	    "recent_path": "/",
	    "extensions": null,
	    "maxSize": 0,
	    "signature": null,
	    "policy": null,
	    "custom_providers": "imgur,cloudapp",
	    "intra": false
	  }
	}`
}

func UploadJSON() string {
	return `{
	   "action":"fp.upload",
	   "timestamp":1443444905,
	   "id":100946,
	   "text":{
	      "url":"https://www.filestackapi.com/api/file/WAunDTTqQfCNWwUUyf6n",
	      "client":"Facebook",
	      "type":"image/jpeg",
	      "filename":"1579337399020824.jpg",
	      "size":139154
	   }
	}`
}
