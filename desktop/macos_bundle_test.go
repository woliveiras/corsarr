package main

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
	"testing"
)

func TestMacOSBundleUsesCorsarrIcon(t *testing.T) {
	plist, err := os.ReadFile("build/darwin/Info.plist")
	if err != nil {
		t.Fatal(err)
	}
	iconDeclaration := []byte("<key>CFBundleIconFile</key>\n    <string>iconfile</string>")
	if !bytes.Contains(plist, iconDeclaration) {
		t.Fatal("macOS bundle does not declare the generated iconfile.icns resource")
	}

	appIcon, err := openImage("build/appicon.png")
	if err != nil {
		t.Fatal(err)
	}
	corsarrLogo, err := openImage("../assets/corsarr-logo-transparent.png")
	if err != nil {
		t.Fatal(err)
	}
	if !appIcon.Bounds().Eq(corsarrLogo.Bounds()) {
		t.Fatal("macOS app icon dimensions differ from the canonical Corsarr logo")
	}
	if appIcon.Bounds().Dx() != 1024 || appIcon.Bounds().Dy() != 1024 {
		t.Fatal("macOS app icon must be 1024 by 1024 pixels")
	}
	for y := appIcon.Bounds().Min.Y; y < appIcon.Bounds().Max.Y; y++ {
		for x := appIcon.Bounds().Min.X; x < appIcon.Bounds().Max.X; x++ {
			appRed, appGreen, appBlue, appAlpha := appIcon.At(x, y).RGBA()
			logoRed, logoGreen, logoBlue, logoAlpha := corsarrLogo.At(x, y).RGBA()
			if appRed != logoRed || appGreen != logoGreen || appBlue != logoBlue || appAlpha != logoAlpha {
				t.Fatalf("macOS app icon differs from the canonical Corsarr logo at %d,%d", x, y)
			}
		}
	}
}

func openImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoded, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
