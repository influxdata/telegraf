package all

import (
	// Importing for initialization side effects to register JWTGenerator with SecretStore
	_ "github.com/influxdata/telegraf/plugins/secretstores/jwt"
)
