//go:build !darwin

package hostreadiness

import "context"

type unsupportedChecker struct {
	platform     string
	architecture string
}

func NewChecker(platform, architecture, _ string) Checker {
	return unsupportedChecker{platform: platform, architecture: architecture}
}

func (c unsupportedChecker) Check(context.Context) Status {
	return Evaluate(Facts{Platform: c.platform, Architecture: c.architecture})
}
