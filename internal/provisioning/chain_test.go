package provisioning

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestChainProvisionerRunsInDeclaredOrder(t *testing.T) {
	operations := []string{}
	chain := NewChainProvisioner(
		chainStep{name: "arr", operations: &operations},
		chainStep{name: "qbittorrent", operations: &operations},
	)

	if err := chain.Provision(context.Background(), "/host/Corsarr", "radarr"); err != nil {
		t.Fatalf("run provisioning chain: %v", err)
	}
	want := []string{"arr:radarr", "qbittorrent:radarr"}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("unexpected provisioning order\nwant: %v\n got: %v", want, operations)
	}
}

func TestChainProvisionerStopsAfterFailure(t *testing.T) {
	operations := []string{}
	chain := NewChainProvisioner(
		chainStep{name: "arr", operations: &operations, err: errors.New("failed")},
		chainStep{name: "qbittorrent", operations: &operations},
	)

	if err := chain.Provision(context.Background(), "/host/Corsarr", "radarr"); err == nil {
		t.Fatal("expected provisioning failure")
	}
	if !reflect.DeepEqual(operations, []string{"arr:radarr"}) {
		t.Fatalf("expected later provisioner not to run, got %v", operations)
	}
}

type chainStep struct {
	name       string
	operations *[]string
	err        error
}

func (s chainStep) Provision(_ context.Context, _ string, applicationID string) error {
	*s.operations = append(*s.operations, s.name+":"+applicationID)
	return s.err
}
