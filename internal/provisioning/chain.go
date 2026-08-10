package provisioning

import (
	"context"
	"fmt"
)

type ApplicationProvisioner interface {
	Provision(ctx context.Context, rootPath string, applicationID string) error
}

type ChainProvisioner struct {
	steps []ApplicationProvisioner
}

func NewChainProvisioner(steps ...ApplicationProvisioner) *ChainProvisioner {
	return &ChainProvisioner{steps: append([]ApplicationProvisioner(nil), steps...)}
}

func (p *ChainProvisioner) Provision(
	ctx context.Context,
	rootPath string,
	applicationID string,
) error {
	for index, step := range p.steps {
		if err := step.Provision(ctx, rootPath, applicationID); err != nil {
			return fmt.Errorf("provisioning step %d for %s: %w", index+1, applicationID, err)
		}
	}
	return nil
}
