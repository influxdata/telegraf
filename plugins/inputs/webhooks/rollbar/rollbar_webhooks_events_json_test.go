package rollbar

func NewItemJSON() string {
	return `
	{
	  "event_name": "new_item",
	  "data": {
		"item": {
		  "public_item_id": null,
		  "integrations_data": {},
		  "last_activated_timestamp": 1382655421,
		  "unique_occurrences": null,
		  "id": 272716944,
		  "environment": "production",
		  "title": "testing aobg98wrwe",
		  "last_occurrence_id": 481761639,
		  "last_occurrence_timestamp": 1382655421,
		  "platform": 0,
		  "first_occurrence_timestamp": 1382655421,
		  "project_id": 90,
		  "resolved_in_version": null,
		  "status": 1,
		  "hash": "c595b2ae0af9b397bb6bdafd57104ac4d5f6b382",
		  "last_occurrence": {
			"body": {
			  "message": {
				"body": "testing aobg98wrwe"
			  }
			},
			"uuid": "d2036647-e0b7-4cad-bc98-934831b9b6d1",
			"language": "python",
			"level": "error",
			"timestamp": 1382655421,
			"server": {
			  "host": "dev",
			  "argv": [
				""
			  ]
			},
			"environment": "production",
			"framework": "unknown",
			"notifier": {
			  "version": "0.5.12",
			  "name": "pyrollbar"
			},
			"metadata": {
			  "access_token": "",
			  "debug": {
				"routes": {
				  "start_time": 1382212080401,
				  "counters": {
					"post_item": 3274122
				  }
				}
			  },
			  "customer_timestamp": 1382655421,
			  "api_server_hostname": "web6"
			}
		  },
		  "framework": 0,
		  "total_occurrences": 1,
		  "level": 40,
		  "counter": 4,
		  "first_occurrence_id": 481761639,
		  "activating_occurrence_id": 481761639
		}
	  }
	}`
}

func DeployJSON() string {
	return `
    {
      "event_name": "deploy",
      "data": {
        "deploy": {
          "comment": "deploying webs",
          "user_id": 1,
          "finish_time": 1382656039,
          "start_time": 1382656038,
          "id": 187585,
          "environment": "production",
          "project_id": 90,
          "local_username": "brian",
          "revision": "e4b9b7db860b2e5ac799f8c06b9498b71ab270bb"
        }
      }
    }`
}

func UnknowJSON() string {
	return `
    {
      "event_name": "roger"
    }`
}
