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

func VideoConversionJSON() string {
	return `{
	   "status":"completed",
	   "message":"Done",
	   "data":{
	      "thumb":"https://cdn.filestackcontent.com/f1e8V88QDuxzOvtOAq1W",
	      "thumb100x100":"https://process.filestackapi.com/AhTgLagciQByzXpFGRI0Az/resize=w:100,h:100,f:crop/output=f:jpg,q:66/https://cdn.filestackcontent.com/f1e8V88QDuxzOvtOAq1W",
	      "thumb200x200":"https://process.filestackapi.com/AhTgLagciQByzXpFGRI0Az/resize=w:200,h:200,f:crop/output=f:jpg,q:66/https://cdn.filestackcontent.com/f1e8V88QDuxzOvtOAq1W",
	      "thumb300x300":"https://process.filestackapi.com/AhTgLagciQByzXpFGRI0Az/resize=w:300,h:300,f:crop/output=f:jpg,q:66/https://cdn.filestackcontent.com/f1e8V88QDuxzOvtOAq1W",
	      "url":"https://cdn.filestackcontent.com/VgvFVdvvTkml0WXPIoGn"
	   },
	   "metadata":{
	      "result":{
	         "audio_channels":2,
	         "audio_codec":"vorbis",
	         "audio_sample_rate":44100,
	         "created_at":"2015/12/21 20:45:19 +0000",
	         "duration":10587,
	         "encoding_progress":100,
	         "encoding_time":8,
	         "extname":".webm",
	         "file_size":293459,
	         "fps":24,
	         "height":260,
	         "mime_type":"video/webm",
	         "started_encoding_at":"2015/12/21 20:45:22 +0000",
	         "updated_at":"2015/12/21 20:45:32 +0000",
	         "video_bitrate":221,
	         "video_codec":"vp8",
	         "width":300
	      },
	      "source":{
	         "audio_bitrate":125,
	         "audio_channels":2,
	         "audio_codec":"aac",
	         "audio_sample_rate":44100,
	         "created_at":"2015/12/21 20:45:19 +0000",
	         "duration":10564,
	         "extname":".mp4",
	         "file_size":875797,
	         "fps":24,
	         "height":360,
	         "mime_type":"video/mp4",
	         "updated_at":"2015/12/21 20:45:32 +0000",
	         "video_bitrate":196,
	         "video_codec":"h264",
	         "width":480
	      }
	   },
	   "timestamp":"1453850583",
	   "uuid":"638311d89d2bc849563a674a45809b7c"
	}`
}
