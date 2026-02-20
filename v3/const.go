package v3

import (
	"fmt"
	"runtime"
)

const Endpoint = "https://api.gh.ink/v3"

var Version = [3]int{1, 0, 9}
var UserAgent = fmt.Sprintf(
	"GhinkOpenAPISDK-Go/%d.%d.%d (%s; %s)",
	Version[0], Version[1], Version[2],
	runtime.GOOS, runtime.GOARCH,
)

type MapAny map[string]any
