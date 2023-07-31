package lustre2

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Fields map[string]interface{}

type Tags map[string]string

// RetrvHealthCheck check health status of a node in lustre.
//
//	@param Log
//	@return Fields
//	@return Tags
//	@return error
func RetrvHealthCheck() (Fields, Tags, error) {

	var err error
	var stdout, stderr bytes.Buffer

	/* Executing command and get result. */
	cmd := exec.Command("lctl", "get_param", "-n", "health_check")
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err = cmd.Run(); err != nil {
		// log.Errorf("unable to execute `%s`: %w", cmd.String(), err)
		return Fields{}, Tags{}, fmt.Errorf("unable to execute `%s`: %w", cmd.String(), err)
	}
	rlt := strings.Replace(stdout.String(), "\n", "", -1)

	/* Checking health. */
	if strings.ToLower(rlt) == "healthy" {
		return Fields{
			"health_check": 1,
		}, Tags{}, nil
	}

	return Fields{
		"health_check": 0,
	}, Tags{}, nil
}
