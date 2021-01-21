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

func OccurrenceJSON() string {
	return `
	{
	  "event_name": "occurrence",
	  "data": {
		"item": {
		  "public_item_id": null,
		  "integrations_data": {},
		  "level_lock": 0,
		  "last_activated_timestamp": 1471624512,
		  "assigned_user_id": null,
		  "hash": "188fc37fa6e641a4d4a3d0198938a1937d31ddbe",
		  "id": 402860571,
		  "environment": "production",
		  "title": "Exception: test exception",
		  "last_occurrence_id": 16298872829,
		  "last_occurrence_timestamp": 1472226345,
		  "platform": 0,
		  "first_occurrence_timestamp": 1471624512,
		  "project_id": 78234,
		  "resolved_in_version": null,
		  "status": 1,
		  "unique_occurrences": null,
		  "title_lock": 0,
		  "framework": 6,
		  "total_occurrences": 8,
		  "level": 40,
		  "counter": 2,
		  "last_modified_by": 8247,
		  "first_occurrence_id": 16103102935,
		  "activating_occurrence_id": 16103102935
                },
		"occurrence": {
		  "body": {
		    "trace": {
		    "frames": [{"method": "<main>", "lineno": 27, "filename": "/Users/rebeccastandig/Desktop/Dev/php-rollbar-app/index.php"}], "exception": {
		    "message": "test 2",
		    "class": "Exception"}
		    }
		  },
		  "uuid": "84d4eccd-b24d-47ae-a42b-1a2f9a82fb82",
		  "language": "php",
		  "level": "error",
		  "timestamp": 1472226345,
		  "php_context": "cli",
		  "environment": "production",
		  "framework": "php",
		  "person": null,
		  "server": {
		    "host": "Rebeccas-MacBook-Pro.local",
		    "argv": ["index.php"]
		  },
		  "notifier": {
		    "version": "0.18.2",
		    "name": "rollbar-php"
		  },
		  "metadata": {
		    "customer_timestamp": 1472226359
		  }
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
