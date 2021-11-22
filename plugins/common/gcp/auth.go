package gcp

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2/google"
)

func GetToken(secret string, email string, url string) string {
	token, err := generateJWT(secret, email, url, 120)

	if err != nil {
		println(err.Error())
	}

	accessToken, err := getGoogleID(token)
	if err != nil {
		println(err.Error())
	}
	return accessToken
}

// https://cloud.google.com/endpoints/docs/openapi/service-account-authentication#go
func generateJWT(saKeyfile, saEmail, audience string, expiryLength int64) (string, error) {
	now := time.Now().Unix()
	gcpauth := "https://www.googleapis.com/oauth2/v4/token"

	// Build the JWT payload.
	jwt := &ClaimSet{
		Iat: now,
		// expires after 'expiraryLength' seconds.
		Exp: now + expiryLength,
		// Iss must match 'issuer' in the security configuration in your
		// swagger spec (e.g. service account email). It can be any string.
		Iss: saEmail,
		// Aud must be either your Endpoints service name, or match the value
		// specified as the 'x-google-audience' in the OpenAPI document.
		Aud: gcpauth,
		// Sub and Email should match the service account's email address.
		Sub:           saEmail,
		PrivateClaims: map[string]interface{}{"target_audience": audience},
	}
	jwsHeader := &Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	// Extract the RSA private key from the service account keyfile.
	sa, err := ioutil.ReadFile(saKeyfile)
	if err != nil {
		return "", fmt.Errorf("could not read service account file: %v", err)
	}

	conf, err := google.JWTConfigFromJSON(sa)
	if err != nil {
		return "", fmt.Errorf("could not parse service account JSON: %v", err)
	}

	block, _ := pem.Decode(conf.PrivateKey)

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("private key parse error: %v", err)
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	// Sign the JWT with the service account's private key.
	if !ok {
		return "", errors.New("private key failed rsa.PrivateKey type assertion")
	}

	return Encode(jwsHeader, jwt, rsaKey)
}

func getGoogleID(jwtToken string) (string, error) {
	var accessToken GoogleID
	googleidurl := "https://www.googleapis.com/oauth2/v4/token"
	responseBody, err := callAPIEndpoint("POST", googleidurl, jwtToken, nil)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(responseBody, &accessToken)
	if err != nil {
		return "", err
	}
	return accessToken.Token, err
}

//GoogleID is used to capture token
type GoogleID struct {
	Token string `json:"id_token"`
}

// CallAPIEndpoint Makes a call to a specified endpoint taking parameters method, url token and some payload
func callAPIEndpoint(method string, urls string, token string, payload io.Reader) ([]byte, error) {
	granttype := "urn:ietf:params:oauth:grant-type:jwt-bearer"

	res, err := http.PostForm(urls, url.Values{"grant_type": {granttype}, "assertion": {token}})
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	// TODO: Should retry a set number of times before erroring out
	if res.StatusCode >= 400 {
		return []byte{}, fmt.Errorf("error generating google id token jwt")
	}
	return body, nil
}
