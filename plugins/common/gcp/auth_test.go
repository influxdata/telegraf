package gcp

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2/google"
)

// Testing GetToken should be the goal
// func TestGetToken(t *testing.T) {
// 	type args struct {
// 		secret string
// 		email  string
// 		url    string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := GetToken(tt.args.secret, tt.args.email, tt.args.url); got != tt.want {
// 				t.Errorf("GetToken() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

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
				saKeyfile: "./testdata/test_key_file.json",
			},
			want: args{
				saEmail:      "test-service-account-email@example.com",
				audience:     "https://oauth2.googleapis.com/token",
				expiryLength: time.Now().Unix() + 120,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Move this configuration setup to a separate function?
			sa, _ := ioutil.ReadFile(tt.args.saKeyfile)
			conf, _ := google.JWTConfigFromJSON(sa)

			got, err := generateJWT(tt.args.saKeyfile, conf.Audience, tt.args.expiryLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			claims := jwt.StandardClaims{}
			jwt.ParseWithClaims(got, &claims, func(token *jwt.Token) (interface{}, error) {
				return nil, nil
			})

			// TODO: What all claims do we want to check here?
			if claims.Audience != tt.want.audience {
				t.Errorf("generateJWT() got = %v, want %v", claims.Audience, tt.want.audience)
			}
			if claims.Subject != tt.want.saEmail {
				t.Errorf("generateJWT() got = %v, want %v", claims.Subject, tt.want.saEmail)
			}
			// if claims.ExpiresAt != tt.want.expiryLength {
			// 	t.Errorf("generateJWT() got = %v, want %v", claims.ExpiresAt, tt.want.expiryLength)
			// }
		})
	}
}
