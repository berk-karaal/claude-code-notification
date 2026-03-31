package script

import "os/exec"

// execCommand is defined to allow mocking exec.Command in tests.
var execCommand = exec.Command
