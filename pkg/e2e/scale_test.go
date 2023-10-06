/*
Copyright 2020 Docker Compose CLI authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package e2e

import (
	"fmt"
	"strings"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"
)

const NO_STATE_TO_CHECK = ""

func TestScaleBasicCases(t *testing.T) {
	c := NewCLI(t, WithEnv(
		"COMPOSE_PROJECT_NAME=scale-basic-tests"))

	reset := func() {
		c.RunDockerComposeCmd(t, "down", "--rmi", "all")
	}
	t.Cleanup(reset)
	res := c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "up", "-d")
	res.Assert(t, icmd.Success)

	t.Log("scale up one service")
	res = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "dbadmin=2")
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-dbadmin", "Started", 2)

	t.Log("scale up 2 services")
	res = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "front=3", "back=2")
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-front", "Running", 2)
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-front", "Started", 1)
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-back", "Running", 1)
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-back", "Started", 1)

	t.Log("scale down one service")
	res = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "dbadmin=1")
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-dbadmin", "Running", 1)

	t.Log("scale to 0 a service")
	res = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "dbadmin=0")
	assert.Check(t, res.Stdout() == "", res.Stdout())

	t.Log("scale down 2 services")
	res = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "front=2", "back=1")
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-front", "Running", 2)
	assert.Check(t, !strings.Contains(res.Combined(), "Container scale-basic-tests-front-3  Running"), res.Combined())
	checkServiceContainer(t, res.Combined(), "scale-basic-tests-back", "Running", 1)
}

func TestScaleWithDepsCases(t *testing.T) {
	c := NewCLI(t, WithEnv(
		"COMPOSE_PROJECT_NAME=scale-deps-tests"))

	reset := func() {
		c.RunDockerComposeCmd(t, "down", "--rmi", "all")
	}
	t.Cleanup(reset)
	res := c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "up", "-d", "--scale", "db=2")
	res.Assert(t, icmd.Success)

	res = c.RunDockerComposeCmd(t, "ps")
	checkServiceContainer(t, res.Combined(), "scale-deps-tests-db", NO_STATE_TO_CHECK, 2)

	t.Log("scale up 1 service with --no-deps")
	_ = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "--no-deps", "back=2")
	res = c.RunDockerComposeCmd(t, "ps")
	checkServiceContainer(t, res.Combined(), "scale-deps-tests-back", NO_STATE_TO_CHECK, 2)
	checkServiceContainer(t, res.Combined(), "scale-deps-tests-db", NO_STATE_TO_CHECK, 2)

	t.Log("scale up 1 service without --no-deps")
	_ = c.RunDockerComposeCmd(t, "--project-directory", "fixtures/scale", "scale", "back=2")
	res = c.RunDockerComposeCmd(t, "ps")
	checkServiceContainer(t, res.Combined(), "scale-deps-tests-back", NO_STATE_TO_CHECK, 2)
	checkServiceContainer(t, res.Combined(), "scale-deps-tests-db", NO_STATE_TO_CHECK, 1)
}

func checkServiceContainer(t *testing.T, stdout, containerName, containerState string, count int) {
	found := 0
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, containerName) && strings.Contains(line, containerState) {
			found++
		}
	}
	if found == count {
		return
	}
	errMessage := fmt.Sprintf("expected %d but found %d instance(s) of container %s in stoud", count, found, containerName)
	if containerState != "" {
		errMessage += fmt.Sprintf(" with expected state %s", containerState)
	}
	testify.Fail(t, errMessage, stdout)
}
