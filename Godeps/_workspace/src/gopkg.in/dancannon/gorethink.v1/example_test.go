package gorethink

import (
	"fmt"
)

func Example() {
	session, err := Connect(ConnectOpts{
		Address: url,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	res, err := Expr("Hello World").Run(session)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var response string
	err = res.One(&response)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println(response)

	// Output:
	// Hello World
}
