// +build go1.7

package xmlutil

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

type implicitPayload struct {
	_ struct{} `type:"structure"`

	StrVal *string     `type:"string"`
	Second *nestedType `type:"structure"`
	Third  *nestedType `type:"structure"`
}

type namedImplicitPayload struct {
	_ struct{} `type:"structure" locationName:"namedPayload"`

	StrVal *string     `type:"string"`
	Second *nestedType `type:"structure"`
	Third  *nestedType `type:"structure"`
}

type explicitPayload struct {
	_ struct{} `type:"structure" payload:"Second"`

	Second *nestedType `type:"structure" locationName:"Second"`
}

type nestedType struct {
	_ struct{} `type:"structure"`

	IntVal *int64  `type:"integer"`
	StrVal *string `type:"string"`
}

func TestBuildXML(t *testing.T) {
	cases := map[string]struct {
		Input  interface{}
		Expect string
	}{
		"explicit payload": {
			Input: &explicitPayload{
				Second: &nestedType{
					IntVal: aws.Int64(1234),
					StrVal: aws.String("string value"),
				},
			},
			Expect: `<Second><IntVal>1234</IntVal><StrVal>string value</StrVal></Second>`,
		},
		"implicit payload": {
			Input: &implicitPayload{
				StrVal: aws.String("string value"),
				Second: &nestedType{
					IntVal: aws.Int64(1111),
					StrVal: aws.String("second string"),
				},
				Third: &nestedType{
					IntVal: aws.Int64(2222),
					StrVal: aws.String("third string"),
				},
			},
			Expect: `<Second><IntVal>1111</IntVal><StrVal>second string</StrVal></Second><StrVal>string value</StrVal><Third><IntVal>2222</IntVal><StrVal>third string</StrVal></Third>`,
		},
		"named implicit payload": {
			Input: &namedImplicitPayload{
				StrVal: aws.String("string value"),
				Second: &nestedType{
					IntVal: aws.Int64(1111),
					StrVal: aws.String("second string"),
				},
				Third: &nestedType{
					IntVal: aws.Int64(2222),
					StrVal: aws.String("third string"),
				},
			},
			Expect: `<namedPayload><Second><IntVal>1111</IntVal><StrVal>second string</StrVal></Second><StrVal>string value</StrVal><Third><IntVal>2222</IntVal><StrVal>third string</StrVal></Third></namedPayload>`,
		},
		"empty nested type": {
			Input: &namedImplicitPayload{
				StrVal: aws.String("string value"),
				Second: &nestedType{},
				Third: &nestedType{
					IntVal: aws.Int64(2222),
					StrVal: aws.String("third string"),
				},
			},
			Expect: `<namedPayload><Second></Second><StrVal>string value</StrVal><Third><IntVal>2222</IntVal><StrVal>third string</StrVal></Third></namedPayload>`,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var w bytes.Buffer
			if err := buildXML(c.Input, xml.NewEncoder(&w), true); err != nil {
				t.Fatalf("expect no error, %v", err)
			}

			if e, a := c.Expect, w.String(); e != a {
				t.Errorf("expect:\n%s\nactual:\n%s\n", e, a)
			}
		})
	}
}
