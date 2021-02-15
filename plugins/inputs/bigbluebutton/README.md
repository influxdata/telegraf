# BigBlueButton Input Plugin

The BigBlueButton Input Plugin gathers metrics from [BigBlueButton](https://bigbluebutton.org/) server. It uses [BigBlueButton API](https://docs.bigbluebutton.org/dev/api.html) `getMeetings` and `getRecordings` endpoints to query the data.

## Configuration

```toml
[[inputs.bigbluebutton]]
	## Required BigBlueButton server url
	url = "http://localhost:8090"

	## Required BigBlueButton api endpoint
	api_endpoint = "/bigbluebutton/api/"

	## Required BigBlueButton secret key
	# secret_key =

	## Server name. Used as "server" tag if not empty. Otherwise use url
	# server_name = 
```

## Metrics

- bigbluebutton_meetings:
  - tags:
    - server_name (configured server_name or server url)
  - fields:
    - participant_count
    - listener_count
    - voice_participant_count
    - video_count
    - active_recording
- bigbluebutton_recordings:
  - tags:
    - server_name (configured server_name or server url)
  - fields:
    - recordings_count
    - published_recordings_count

## Example output
``` 
bigbluebutton_meetings,host=codespaces_700987,server_name=http://localhost:8090 voice_participant_count=0i,video_count=0i,active_recording=0i,participant_count=5i,listener_count=0i 1613389390000000000
bigbluebutton_recordings,host=codespaces_700987,server_name=http://localhost8090 published_recordings_count=0i,recordings_count=0i 1613389390000000000
```