package mandrill

func SendEventJSON() string {
	return `
	{
	    "event": "send",
	    "msg": {
	      "ts": 1365109999,
	      "subject": "This an example webhook message",
	      "email": "example.webhook@mandrillapp.com",
	      "sender": "example.sender@mandrillapp.com",
	      "tags": [
	        "webhook-example"
	      ],
	      "opens": [

	      ],
	      "clicks": [

	      ],
	      "state": "sent",
	      "metadata": {
	        "user_id": 111
	      },
	      "_id": "exampleaaaaaaaaaaaaaaaaaaaaaaaaa",
	      "_version": "exampleaaaaaaaaaaaaaaa"
	    },
	    "_id": "id1",
	    "ts": 1384954004
	}`
}

func HardBounceEventJSON() string {
	return `
	{
	    "event": "hard_bounce",
	    "msg": {
	      "ts": 1365109999,
	      "subject": "This an example webhook message",
	      "email": "example.webhook@mandrillapp.com",
	      "sender": "example.sender@mandrillapp.com",
	      "tags": [
	        "webhook-example"
	      ],
	      "state": "bounced",
	      "metadata": {
	        "user_id": 111
	      },
	      "_id": "exampleaaaaaaaaaaaaaaaaaaaaaaaaa2",
	      "_version": "exampleaaaaaaaaaaaaaaa",
	      "bounce_description": "bad_mailbox",
	      "bgtools_code": 10,
	      "diag": "smtp;550 5.1.1 The email account that you tried to reach does not exist. Please try double-checking the recipient's email address for typos or unnecessary spaces."
	    },
	    "_id": "id2",
	    "ts": 1384954004
	}`
}
