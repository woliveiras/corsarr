package orchestrator

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/woliveiras/corsarr/internal/catalog"
	containerruntime "github.com/woliveiras/corsarr/internal/runtime"
	"github.com/woliveiras/corsarr/internal/storage"
)

func TestUpdaterBacksUpAndReplacesRunningApplication(t *testing.T) {
	approved := validInstallerSpec("radarr")
	previousImage := "example.invalid/app@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	runtime := newUpdaterRuntime(previousImage)
	backup := &fakeBackupCreator{result: storage.BackupResult{ApplicationID: "radarr", Path: "/tmp/backup.tar.gz", SHA256: "sum"}}
	updater := NewUpdater(runtime, &fakeSpecResolver{spec: approved}, &fakeReadiness{}, backup)

	result, err := updater.Update(context.Background(), "radarr", "/tmp/Corsarr", catalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("update application: %v", err)
	}
	if !result.Updated || result.RolledBack || result.PreviousImage != previousImage || result.Status.Image != approved.Image {
		t.Fatalf("unexpected update result %#v", result)
	}
	if !reflect.DeepEqual(backup.calls, []string{"/tmp/Corsarr:radarr"}) {
		t.Fatalf("unexpected backup calls %v", backup.calls)
	}
	want := []string{"network", "inspect", "pull", "stop", "remove", "create:" + approved.Image, "start", "inspect"}
	if !reflect.DeepEqual(runtime.operations, want) {
		t.Fatalf("unexpected operations\nwant: %v\n got: %v", want, runtime.operations)
	}
}

func TestUpdaterRestoresPreviousImageWhenReadinessFails(t *testing.T) {
	approved := validInstallerSpec("sonarr")
	previousImage := "example.invalid/app@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	runtime := newUpdaterRuntime(previousImage)
	readiness := &sequenceReadiness{errors: []error{errors.New("new image not ready"), nil}}
	updater := NewUpdater(runtime, &fakeSpecResolver{spec: approved}, readiness, &fakeBackupCreator{})

	result, err := updater.Update(context.Background(), "sonarr", "/tmp/Corsarr", catalog.RuntimeOptions{})
	if err == nil {
		t.Fatal("expected update failure")
	}
	if result.Updated || !result.RolledBack || result.Status.Image != previousImage {
		t.Fatalf("expected successful image rollback, got %#v", result)
	}
	want := []string{
		"network", "inspect", "pull", "stop", "remove", "create:" + approved.Image,
		"start", "inspect", "remove", "create:" + previousImage, "start", "inspect",
	}
	if !reflect.DeepEqual(runtime.operations, want) {
		t.Fatalf("unexpected rollback operations\nwant: %v\n got: %v", want, runtime.operations)
	}
}

func TestUpdaterDoesNothingWhenApprovedImageIsAlreadyInstalled(t *testing.T) {
	approved := validInstallerSpec("lidarr")
	runtime := newUpdaterRuntime(approved.Image)
	backup := &fakeBackupCreator{}
	updater := NewUpdater(runtime, &fakeSpecResolver{spec: approved}, &fakeReadiness{}, backup)

	result, err := updater.Update(context.Background(), "lidarr", "/tmp/Corsarr", catalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("reconcile current image: %v", err)
	}
	if result.Updated || len(backup.calls) != 0 {
		t.Fatalf("expected no update or backup, got %#v and %v", result, backup.calls)
	}
	if !reflect.DeepEqual(runtime.operations, []string{"network", "inspect"}) {
		t.Fatalf("unexpected operations %v", runtime.operations)
	}
}

func TestUpdaterDoesNotMutateContainerWhenBackupFails(t *testing.T) {
	approved := validInstallerSpec("radarr")
	runtime := newUpdaterRuntime("example.invalid/app@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
	updater := NewUpdater(runtime, &fakeSpecResolver{spec: approved}, &fakeReadiness{}, &fakeBackupCreator{
		err: errors.New("disk full"),
	})

	if _, err := updater.Update(context.Background(), "radarr", "/tmp/Corsarr", catalog.RuntimeOptions{}); err == nil {
		t.Fatal("expected backup failure")
	}
	if !reflect.DeepEqual(runtime.operations, []string{"network", "inspect"}) {
		t.Fatalf("container must remain untouched, got operations %v", runtime.operations)
	}
}

func TestUpdaterPreservesStoppedStateAfterVerification(t *testing.T) {
	approved := validInstallerSpec("radarr")
	runtime := newUpdaterRuntime("example.invalid/app@sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	runtime.status.State = containerruntime.ContainerStateStopped
	updater := NewUpdater(runtime, &fakeSpecResolver{spec: approved}, &fakeReadiness{}, &fakeBackupCreator{})

	result, err := updater.Update(context.Background(), "radarr", "/tmp/Corsarr", catalog.RuntimeOptions{})
	if err != nil {
		t.Fatalf("update stopped application: %v", err)
	}
	if result.Status.State != containerruntime.ContainerStateStopped {
		t.Fatalf("expected stopped state after update, got %#v", result.Status)
	}
	want := []string{"network", "inspect", "pull", "remove", "create:" + approved.Image, "start", "inspect", "stop", "inspect"}
	if !reflect.DeepEqual(runtime.operations, want) {
		t.Fatalf("unexpected stopped update operations\nwant: %v\n got: %v", want, runtime.operations)
	}
}

type fakeBackupCreator struct {
	calls  []string
	result storage.BackupResult
	err    error
}

func (b *fakeBackupCreator) Backup(rootPath, applicationID string) (storage.BackupResult, error) {
	b.calls = append(b.calls, rootPath+":"+applicationID)
	return b.result, b.err
}

type sequenceReadiness struct {
	errors []error
	calls  int
}

func (r *sequenceReadiness) Wait(context.Context, string) error {
	defer func() { r.calls++ }()
	if r.calls < len(r.errors) {
		return r.errors[r.calls]
	}
	return nil
}

type updaterRuntime struct {
	operations []string
	status     containerruntime.ContainerStatus
}

func newUpdaterRuntime(image string) *updaterRuntime {
	return &updaterRuntime{status: containerruntime.ContainerStatus{
		ApplicationID: "radarr", State: containerruntime.ContainerStateRunning, Image: image,
	}}
}

func (r *updaterRuntime) EnsureNetwork(context.Context) error {
	r.operations = append(r.operations, "network")
	return nil
}
func (r *updaterRuntime) Pull(context.Context, string) error {
	r.operations = append(r.operations, "pull")
	return nil
}
func (r *updaterRuntime) Create(_ context.Context, spec containerruntime.ContainerSpec) error {
	r.operations = append(r.operations, "create:"+spec.Image)
	r.status.ApplicationID = spec.ApplicationID
	r.status.Image = spec.Image
	r.status.State = containerruntime.ContainerStateCreated
	return nil
}
func (r *updaterRuntime) Inspect(context.Context, string) (containerruntime.ContainerStatus, error) {
	r.operations = append(r.operations, "inspect")
	return r.status, nil
}
func (r *updaterRuntime) Start(context.Context, string) error {
	r.operations = append(r.operations, "start")
	r.status.State = containerruntime.ContainerStateRunning
	return nil
}
func (r *updaterRuntime) Stop(context.Context, string) error {
	r.operations = append(r.operations, "stop")
	r.status.State = containerruntime.ContainerStateStopped
	return nil
}
func (r *updaterRuntime) Restart(context.Context, string) error { return nil }
func (r *updaterRuntime) Remove(context.Context, string) error {
	r.operations = append(r.operations, "remove")
	return nil
}
func (r *updaterRuntime) Logs(context.Context, string, int) (string, error) { return "", nil }
