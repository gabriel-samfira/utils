// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package exec_test

import (
	"os"

    jc "github.com/juju/testing/checkers"
    gc "launchpad.net/gocheck"
	"github.com/juju/utils/exec"
	"github.com/juju/testing"
)

type MergeEnvSuite struct {
    testing.IsolationSuite
}

var _ = gc.Suite(&MergeEnvSuite{})

func (e *MergeEnvSuite) TestMergeEnviron(c *gc.C) {
    // environment does not get fully cleared on Windows
    // when using testing.IsolationSuite
    origEnv := os.Environ()
    extraExpected := []string{
        "DUMMYVAR=foo",
        "DUMMYVAR2=bar",
        "NEWVAR=ImNew",
    }
    expectEnv := append(origEnv, extraExpected...)
    os.Setenv("DUMMYVAR2", "ChangeMe")
    os.Setenv("DUMMYVAR", "foo")

    newEnv := exec.MergeEnvironment([]string{"DUMMYVAR2=bar", "NEWVAR=ImNew"})
    c.Assert(expectEnv, jc.SameContents, newEnv)
}

