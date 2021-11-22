package gcp

import (
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

func TestGetToken(t *testing.T) {
	type args struct {
		secret string
		email  string
		url    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetToken(tt.args.secret, tt.args.email, tt.args.url); got != tt.want {
				t.Errorf("GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_callAPIEndpoint(t *testing.T) {
	type args struct {
		method  string
		urls    string
		token   string
		payload io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := callAPIEndpoint(tt.args.method, tt.args.urls, tt.args.token, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("callAPIEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("callAPIEndpoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateJWT(t *testing.T) {
	type args struct {
		saKeyfile    string
		saEmail      string
		audience     string
		expiryLength int64
	}
	tests := []struct {
		name    string
		args    args
		want    args
		wantErr bool
	}{
		// WIP: Adding test cases.
		{
			name: "Same values in claims",
			args: args{
				saKeyfile:    "./testdata/test_key_file.json",
				saEmail:      "test-service-account-email@example.com",
				audience:     "http://example.com",
				expiryLength: time.Now().Unix() + 120,
			},
			want: args{
				saEmail:      "test-service-account-email@example.com",
				audience:     "https://www.googleapis.com/oauth2/v4/token",
				expiryLength: time.Now().Unix() + 120,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("tt.args.saKeyfile", tt.args.saKeyfile, "tt.args.saEmail", tt.args.saEmail, "tt.args.audience", tt.args.audience, "tt.args.expiryLength", tt.args.expiryLength)
			got, err := generateJWT(tt.args.saKeyfile, tt.args.saEmail, tt.args.audience, tt.args.expiryLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// _, err = jwt.ParseWithClaims(h.signedJWT, &claims, func(token *jwt.Token) (interface{}, error) {
			// 	return nil, nil
			// })
			claims := jwt.StandardClaims{}
			jwt.ParseWithClaims(got, &claims, func(token *jwt.Token) (interface{}, error) {
				return nil, nil
			})

			if claims.Audience != tt.want.audience {
				t.Errorf("generateJWT() got = %v, want %v", claims.Audience, tt.want.audience)
			}
			if claims.Subject != tt.want.saEmail {
				t.Errorf("generateJWT() got = %v, want %v", claims.Subject, tt.want.saEmail)
			}
		})
	}
}

func Test_getGoogleID(t *testing.T) {
	type args struct {
		jwtToken string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGoogleID(tt.args.jwtToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGoogleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getGoogleID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
