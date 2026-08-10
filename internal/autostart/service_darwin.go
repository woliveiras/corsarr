//go:build darwin && cgo

package autostart

/*
#cgo CFLAGS: -x objective-c -fmodules
#cgo LDFLAGS: -framework Foundation -framework ServiceManagement

#include <stdlib.h>
#import <Foundation/Foundation.h>
#import <ServiceManagement/ServiceManagement.h>

static int corsarr_main_app_status(void) {
  if (@available(macOS 13.0, *)) {
    return (int)[SMAppService mainAppService].status;
  }
  return -1;
}

static char *corsarr_copy_error(NSError *error) {
  const char *detail = error.localizedDescription.UTF8String;
  return detail == NULL ? strdup("unknown Service Management error") : strdup(detail);
}

static char *corsarr_register_main_app(void) {
  if (@available(macOS 13.0, *)) {
    NSError *error = nil;
    if ([[SMAppService mainAppService] registerAndReturnError:&error]) {
      return NULL;
    }
    return corsarr_copy_error(error);
  }
  return strdup("macOS 13 or later is required");
}

static char *corsarr_unregister_main_app(void) {
  if (@available(macOS 13.0, *)) {
    NSError *error = nil;
    if ([[SMAppService mainAppService] unregisterAndReturnError:&error]) {
      return NULL;
    }
    return corsarr_copy_error(error);
  }
  return strdup("macOS 13 or later is required");
}

static void corsarr_open_login_item_settings(void) {
  if (@available(macOS 13.0, *)) {
    [SMAppService openSystemSettingsLoginItems];
  }
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

type serviceManagementClient struct{}

func newNativeMainAppService() nativeMainAppService { return serviceManagementClient{} }

func (serviceManagementClient) Status() nativeStatus {
	return nativeStatus(C.corsarr_main_app_status())
}

func (serviceManagementClient) Register() error {
	return serviceManagementError(C.corsarr_register_main_app())
}

func (serviceManagementClient) Unregister() error {
	return serviceManagementError(C.corsarr_unregister_main_app())
}

func (serviceManagementClient) OpenSystemSettings() {
	C.corsarr_open_login_item_settings()
}

func serviceManagementError(detail *C.char) error {
	if detail == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(detail))
	return errors.New(C.GoString(detail))
}
