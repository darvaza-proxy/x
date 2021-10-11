package sni

import (
	"fmt"
	"testing"
)

var TestHello12 = []byte{
	0x16, 0x03, 0x01, 0x00, 0xA5, 0x01, 0x00, 0x00, 0xA1,
	0x03, 0x03, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
	0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
	0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x00, 0x00,
	0x20, 0xCC, 0xA8, 0xCC, 0xA9, 0xC0, 0x2F, 0xC0, 0x30,
	0xC0, 0x2B, 0xC0, 0x2C, 0xC0, 0x13, 0xC0, 0x09, 0xC0,
	0x14, 0xC0, 0x0A, 0x00, 0x9C, 0x00, 0x9D, 0x00, 0x2F,
	0x00, 0x35, 0xC0, 0x12, 0x00, 0x0A, 0x01, 0x00, 0x00,
	0x58, 0x00, 0x00, 0x00, 0x18, 0x00, 0x16, 0x00, 0x00,
	0x13, 0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E,
	0x75, 0x6C, 0x66, 0x68, 0x65, 0x69, 0x6D, 0x2E, 0x6E,
	0x65, 0x74, 0x00, 0x05, 0x00, 0x05, 0x01, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x0A, 0x00, 0x0A, 0x00, 0x08, 0x00,
	0x1D, 0x00, 0x17, 0x00, 0x18, 0x00, 0x19, 0x00, 0x0B,
	0x00, 0x02, 0x01, 0x00, 0x00, 0x0D, 0x00, 0x12, 0x00,
	0x10, 0x04, 0x01, 0x04, 0x03, 0x05, 0x01, 0x05, 0x03,
	0x06, 0x01, 0x06, 0x03, 0x02, 0x01, 0x02, 0x03, 0xFF,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x12, 0x00, 0x00,
}

var TestHello13 = []byte{
	0x16, 0x03, 0x01, 0x00, 0xCA, 0x01, 0x00, 0x00, 0xC6,
	0x03, 0x03, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
	0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
	0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20, 0xE0,
	0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9,
	0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF, 0xF0, 0xF1, 0xF2,
	0xF3, 0xF4, 0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFB,
	0xFC, 0xFD, 0xFE, 0xFF, 0x00, 0x06, 0x13, 0x01, 0x13,
	0x02, 0x13, 0x03, 0x01, 0x00, 0x00, 0x77, 0x00, 0x00,
	0x00, 0x18, 0x00, 0x16, 0x00, 0x00, 0x13, 0x65, 0x78,
	0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E, 0x75, 0x6C, 0x66,
	0x68, 0x65, 0x69, 0x6D, 0x2E, 0x6E, 0x65, 0x74, 0x00,
	0x0A, 0x00, 0x08, 0x00, 0x06, 0x00, 0x1D, 0x00, 0x17,
	0x00, 0x18, 0x00, 0x0D, 0x00, 0x14, 0x00, 0x12, 0x04,
	0x03, 0x08, 0x04, 0x04, 0x01, 0x05, 0x03, 0x08, 0x05,
	0x05, 0x01, 0x08, 0x06, 0x06, 0x01, 0x02, 0x01, 0x00,
	0x33, 0x00, 0x26, 0x00, 0x24, 0x00, 0x1D, 0x00, 0x20,
	0x35, 0x80, 0x72, 0xD6, 0x36, 0x58, 0x80, 0xD1, 0xAE,
	0xEA, 0x32, 0x9A, 0xDF, 0x91, 0x21, 0x38, 0x38, 0x51,
	0xED, 0x21, 0xA2, 0x8E, 0x3B, 0x75, 0xE9, 0x65, 0xD0,
	0xD2, 0xCD, 0x16, 0x62, 0x54, 0x00, 0x2D, 0x00, 0x02,
	0x01, 0x01, 0x00, 0x2B, 0x00, 0x03, 0x02, 0x03, 0x04,
}

func Test_GetInfo12(t *testing.T) {
	ci := GetInfo(TestHello12)
	fmt.Println()
	fmt.Println()
	fmt.Println("Version is:", VersionName(ci.Vers))
	fmt.Println("Chiper Suites:", CipherSuites(ci.CipherSuites))
	fmt.Println("Compression Methods:", CompressionMethods(ci.CompressionMethods))
	fmt.Println("Supported Algos:", SignatureAlgos(ci.SupportedSignatureAlgorithms))
	fmt.Println("Supported Versions:", SupportedVersions(ci.SupportedVersions))
	fmt.Println("Requested SNI:", ci.ServerName)
	fmt.Println("ALPN:", ci.ALPNProtocols)
	if ci.ServerName == "" || ci.ServerName != "example.ulfheim.net" {
		t.Fatalf("decode failed, wanted example.ulfheim.net got %s", ci.ServerName)
	}
	fmt.Println()
	fmt.Println()
}
func Test_GetInfo13(t *testing.T) {
	ci := GetInfo(TestHello13)
	fmt.Println()
	fmt.Println()
	fmt.Println("Version is:", VersionName(ci.Vers))
	fmt.Println("Chiper Suites:", CipherSuites(ci.CipherSuites))
	fmt.Println("Compression Methods:", CompressionMethods(ci.CompressionMethods))
	fmt.Println("Supported Algos:", SignatureAlgos(ci.SupportedSignatureAlgorithms))
	fmt.Println("Supported Versions:", SupportedVersions(ci.SupportedVersions))
	fmt.Println("Requested SNI:", ci.ServerName)
	fmt.Println("ALPN:", ci.ALPNProtocols)
	if ci.ServerName == "" || ci.ServerName != "example.ulfheim.net" {
		t.Fatalf("decode failed, wanted example.ulfheim.net got %s", ci.ServerName)
	}
	fmt.Println()
	fmt.Println()
}
