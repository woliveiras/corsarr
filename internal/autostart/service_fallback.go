//go:build !darwin || !cgo

package autostart

type unavailableNativeService struct{}

func newNativeMainAppService() nativeMainAppService { return unavailableNativeService{} }

func (unavailableNativeService) Status() nativeStatus { return nativeStatusUnsupported }

func (unavailableNativeService) Register() error { return ErrUnsupported }

func (unavailableNativeService) Unregister() error { return ErrUnsupported }

func (unavailableNativeService) OpenSystemSettings() {}
