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
	"golang.org/x/oauth2/jwt"
)

// https://cloud.google.com/endpoints/docs/openapi/service-account-authentication#go
// https://developers.google.com/identity/protocols/oauth2
type GoogleID struct {
	Token string `json:"id_token"`
}

func GetAccessToken(saKeyfile string, url string) (string, error) {
	sa, err := ioutil.ReadFile(saKeyfile)
	if err != nil {
		return "", fmt.Errorf("could not read service account file: %v", err)
	}

	conf, err := google.JWTConfigFromJSON(sa)
	if err != nil {
		return "", fmt.Errorf("could not parse service account JSON: %v", err)
	}

	signedJWT, err := generateJWT(conf, url, 120)

	if err != nil {
		println(err.Error())
	}
	// aud ~= token_uri ~= conf.TokenURL
	accessToken, err := getGoogleID(signedJWT, conf.TokenURL)
	if err != nil {
		println(err.Error())
	}

	return accessToken, nil
}

func generateJWT(conf *jwt.Config, audience string, expiryLength int64) (string, error) {
	now := time.Now().Unix()
	// Build the JWT payload.
	jwt := &ClaimSet{
		Iat: now,
		// expires after 'expiraryLength' seconds.
		Exp: now + expiryLength,
		// Iss must match 'issuer' in the security configuration in your
		// swagger spec (e.g. service account email). It can be any string.
		Iss: conf.Email,
		// Aud must be either your Endpoints service name, or match the value
		// specified as the 'x-google-audience' in the OpenAPI document.
		Aud: conf.TokenURL,
		// Sub and Email should match the service account's email address.
		Sub:           conf.Email,
		PrivateClaims: map[string]interface{}{"target_audience": audience},
	}
	jwsHeader := &Header{
		Algorithm: "RS256",
		Typ:       "JWT",
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

func getGoogleID(jwtToken, googleidurl string) (string, error) {
	var googleID GoogleID

	responseBody, err := callAPIEndpoint("POST", googleidurl, jwtToken, nil)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(responseBody, &googleID)
	if err != nil {
		return "", err
	}

	return googleID.Token, err
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
