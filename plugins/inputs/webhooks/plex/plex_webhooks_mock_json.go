package plex

func PlexWebhookEventJSON() string {
	return `{  
    "event": "media.play",
    "user": true,
    "owner": true,
    "Account": {
       "id": 1,
       "thumb": "https://plex.tv/users/1022b120ffbaa/avatar?c=1465525047",
       "title": "elan"
    },
    "Server": {
       "title": "Office",
       "uuid": "54664a3d8acc39983675640ec9ce00b70af9cc36"
    },
    "Player": {
       "local": true,
       "publicAddress": "200.200.200.200",
       "title": "Plex Web (Safari)",
       "uuid": "r6yfkdnfggbh2bdnvkffwbms"
    },
    "Metadata": {
       "librarySectionType": "artist",
       "ratingKey": "1936545",
       "key": "/library/metadata/1936545",
       "parentRatingKey": "1936544",
       "grandparentRatingKey": "1936543",
       "guid": "com.plexapp.agents.plexmusic://gracenote/track/7572499-91016293BE6BF7F1AB2F848F736E74E5/7572500-3CBAE310D4F3E66C285E104A1458B272?lang=en",
       "librarySectionID": 1224,
       "type": "track",
       "title": "Love The One You're With",
       "grandparentKey": "/library/metadata/1936543",
       "parentKey": "/library/metadata/1936544",
       "grandparentTitle": "Stephen Stills",
       "parentTitle": "Stephen Stills",
       "summary": "",
       "index": 1,
       "parentIndex": 1,
       "ratingCount": 6794,
       "thumb": "/library/metadata/1936544/thumb/1432897518",
       "art": "/library/metadata/1936543/art/1485951497",
       "parentThumb": "/library/metadata/1936544/thumb/1432897518",
       "grandparentThumb": "/library/metadata/1936543/thumb/1485951497",
       "grandparentArt": "/library/metadata/1936543/art/1485951497",
       "addedAt": 1000396126,
       "updatedAt": 1432897518
    }
 }`
}
