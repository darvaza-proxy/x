// Package sni provides logic to work with TLS SNI fields
package sni

import (
	"strings"

	"golang.org/x/crypto/cryptobyte"
)

/*  type ClientHelloInfo struct {
	Vers                             uint16
	CipherSuites                     []uint16
	CompressionMethods               []uint8
	ServerName                       string
	SupportedSignatureAlgorithms     []SignatureScheme
	SupportedSignatureAlgorithmsCert []SignatureScheme
	SupportedVersions                []uint16
}*/

// GetInfo returns a pointer to a ClientHelloInfo:
func GetInfo(buf []byte) *ClientHelloInfo {
	n := len(buf)
	if n <= 5 {
		return nil
	}

	// tls record type
	if recordType(buf[0]) != recordTypeHandshake {
		return nil
	}

	// tls major Version
	if buf[1] != 3 {
		return nil
	}

	// handshake message type
	if buf[5] != typeClientHello {
		return nil
	}

	// parse client hello message
	msg := &ClientHelloInfo{}

	ret := msg.unmarshal(buf[5:n])
	if !ret {
		return nil
	}
	// Lower than TLS13,
	if len(msg.SupportedVersions) == 0 {
		msg.SupportedVersions = append(msg.SupportedVersions, msg.Vers)
	}
	return msg
}

/* below this line is an edited portion of
$GOROOT/src/crypto/tls/handshake_messages.go version 1.17

Copyright 2009 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.*/

// ClientHelloInfo contains information from a ClientHello message in order to
// guide application logic in the GetCertificate and GetConfigForClient callbacks.
type ClientHelloInfo struct {
	raw                              []byte
	Vers                             uint16
	random                           []byte
	sessionID                        []byte
	CipherSuites                     []uint16
	CompressionMethods               []uint8
	ServerName                       string
	ocspStapling                     bool
	supportedCurves                  []CurveID
	supportedPoints                  []uint8
	ticketSupported                  bool
	sessionTicket                    []uint8
	SupportedSignatureAlgorithms     []SignatureScheme
	SupportedSignatureAlgorithmsCert []SignatureScheme
	secureRenegotiationSupported     bool
	secureRenegotiation              []byte
	ALPNProtocols                    []string
	scts                             bool
	SupportedVersions                []uint16
	cookie                           []byte
	keyShares                        []keyShare
	earlyData                        bool
	pskModes                         []uint8
	pskIdentities                    []pskIdentity
	pskBinders                       [][]byte
}

func readUint8LengthPrefixed(s *cryptobyte.String, out *[]byte) bool {
	return s.ReadUint8LengthPrefixed((*cryptobyte.String)(out))
}

func readUint16LengthPrefixed(s *cryptobyte.String, out *[]byte) bool {
	return s.ReadUint16LengthPrefixed((*cryptobyte.String)(out))
}

//revive:disable:cognitive-complexity
//revive:disable:cyclomatic
func (m *ClientHelloInfo) unmarshal(data []byte) bool {
	//revive:enable:cognitive-complexity
	//revive:enable:cyclomatic
	*m = ClientHelloInfo{raw: data}
	s := cryptobyte.String(data)

	if !s.Skip(4) || // message type and uint24 length field
		!s.ReadUint16(&m.Vers) || !s.ReadBytes(&m.random, 32) ||
		!readUint8LengthPrefixed(&s, &m.sessionID) {
		return false
	}
	// TODO: Fix revive
	//revive:disable:unexported-naming
	var CipherSuites cryptobyte.String
	//revive:enable:unexported-naming
	if !s.ReadUint16LengthPrefixed(&CipherSuites) {
		return false
	}
	m.CipherSuites = []uint16{}
	m.secureRenegotiationSupported = false
	for !CipherSuites.Empty() {
		var suite uint16
		if !CipherSuites.ReadUint16(&suite) {
			return false
		}
		if suite == scsvRenegotiation {
			m.secureRenegotiationSupported = true
		}
		m.CipherSuites = append(m.CipherSuites, suite)
	}

	if !readUint8LengthPrefixed(&s, &m.CompressionMethods) {
		return false
	}

	if s.Empty() {
		// ClientHello is optionally followed by extension data
		return true
	}

	return m.unmarshalExtensions(s)
}

// revive:disable:cognitive-complexity
// revive:disable:cyclomatic
func (m *ClientHelloInfo) unmarshalExtensions(s cryptobyte.String) bool {
	// revive:enable:cognitive-complexity
	// revive:enable:cyclomatic
	var extensions cryptobyte.String
	if !s.ReadUint16LengthPrefixed(&extensions) || !s.Empty() {
		return false
	}

	for !extensions.Empty() {
		var extension uint16
		var extData cryptobyte.String
		var ok bool

		if !extensions.ReadUint16(&extension) ||
			!extensions.ReadUint16LengthPrefixed(&extData) {
			return false
		}

		switch extension {
		case extensionServerName:
			ok = m.unmarshalServerName(&extData)
		case extensionStatusRequest:
			ok = m.unmarshalStatusRequest(&extData)
		case extensionSupportedCurves:
			ok = m.unmarshalSupportedCurves(&extData)
		case extensionSupportedPoints:
			ok = m.unmarshalSupportedPoints(&extData)
		case extensionSessionTicket:
			ok = m.unmarshalSessionTicket(&extData)
		case extensionSignatureAlgorithms:
			ok = m.unmarshalSignatureAlgorithms(&extData)
		case extensionSignatureAlgorithmsCert:
			ok = m.unmarshalSignatureAlgorithmsCert(&extData)
		case extensionRenegotiationInfo:
			ok = m.unmarshalRenegotiationInfo(&extData)
		case extensionALPN:
			ok = m.unmarshalALPN(&extData)
		case extensionSCT:
			ok = m.unmarshalSCT(&extData)
		case extensionSupportedVersions:
			ok = m.unmarshalSupportedVersions(&extData)
		case extensionCookie:
			ok = m.unmarshalCookie(&extData)
		case extensionKeyShare:
			ok = m.unmarshalKeyShare(&extData)
		case extensionEarlyData:
			ok = m.unmarshalEarlyData(&extData)
		case extensionPSKModes:
			ok = m.unmarshalPSKModes(&extData)
		case extensionPreSharedKey:
			if !extensions.Empty() {
				return false // pre_shared_key must be the last extension
			}
			ok = m.unmarshalPreSharedKey(&extData)
		default:
			// Ignore unknown extensions.
			continue
		}

		if !ok || !extData.Empty() {
			return false
		}
	}

	return true
}

// revive:disable:cognitive-complexity
func (m *ClientHelloInfo) unmarshalServerName(extData *cryptobyte.String) bool {
	// revive:enable:cognitive-complexity
	// RFC 6066, Section 3
	var nameList cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&nameList) || nameList.Empty() {
		return false
	}
	for !nameList.Empty() {
		var nameType uint8
		var serverName cryptobyte.String
		if !nameList.ReadUint8(&nameType) ||
			!nameList.ReadUint16LengthPrefixed(&serverName) ||
			serverName.Empty() {
			return false
		}
		if nameType != 0 {
			continue
		}
		if len(m.ServerName) != 0 {
			// Multiple names of the same name_type are prohibited.
			return false
		}
		m.ServerName = string(serverName)
		// An SNI value may not include a trailing dot.
		if strings.HasSuffix(m.ServerName, ".") {
			return false
		}
	}
	return true
}

func (m *ClientHelloInfo) unmarshalStatusRequest(extData *cryptobyte.String) bool {
	// RFC 4366, Section 3.6
	var statusType uint8
	var ignored cryptobyte.String
	if !extData.ReadUint8(&statusType) ||
		!extData.ReadUint16LengthPrefixed(&ignored) ||
		!extData.ReadUint16LengthPrefixed(&ignored) {
		return false
	}
	m.ocspStapling = statusType == statusTypeOCSP
	return true
}

func (m *ClientHelloInfo) unmarshalSupportedCurves(extData *cryptobyte.String) bool {
	// RFC 4492, sections 5.1.1 and RFC 8446, Section 4.2.7
	var curves cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&curves) || curves.Empty() {
		return false
	}
	for !curves.Empty() {
		var curve uint16
		if !curves.ReadUint16(&curve) {
			return false
		}
		m.supportedCurves = append(m.supportedCurves, CurveID(curve))
	}
	return true
}

func (m *ClientHelloInfo) unmarshalSupportedPoints(extData *cryptobyte.String) bool {
	// RFC 4492, Section 5.1.2
	if !readUint8LengthPrefixed(extData, &m.supportedPoints) ||
		len(m.supportedPoints) == 0 {
		return false
	}
	return true
}

func (m *ClientHelloInfo) unmarshalSessionTicket(extData *cryptobyte.String) bool {
	// RFC 5077, Section 3.2
	m.ticketSupported = true
	extData.ReadBytes(&m.sessionTicket, len(*extData))
	return true
}

func (m *ClientHelloInfo) unmarshalSignatureAlgorithms(extData *cryptobyte.String) bool {
	// RFC 5246, Section 7.4.1.4.1
	var sigAndAlgs cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&sigAndAlgs) || sigAndAlgs.Empty() {
		return false
	}
	for !sigAndAlgs.Empty() {
		var sigAndAlg uint16
		if !sigAndAlgs.ReadUint16(&sigAndAlg) {
			return false
		}
		m.SupportedSignatureAlgorithms = append(
			m.SupportedSignatureAlgorithms, SignatureScheme(sigAndAlg))
	}
	return true
}

func (m *ClientHelloInfo) unmarshalSignatureAlgorithmsCert(extData *cryptobyte.String) bool {
	// RFC 8446, Section 4.2.3
	var sigAndAlgs cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&sigAndAlgs) || sigAndAlgs.Empty() {
		return false
	}
	for !sigAndAlgs.Empty() {
		var sigAndAlg uint16
		if !sigAndAlgs.ReadUint16(&sigAndAlg) {
			return false
		}
		m.SupportedSignatureAlgorithmsCert = append(
			m.SupportedSignatureAlgorithmsCert, SignatureScheme(sigAndAlg))
	}
	return true
}

func (m *ClientHelloInfo) unmarshalRenegotiationInfo(extData *cryptobyte.String) bool {
	// RFC 5746, Section 3.2

	if !readUint8LengthPrefixed(extData, &m.secureRenegotiation) {
		return false
	}
	m.secureRenegotiationSupported = true
	return true
}

func (m *ClientHelloInfo) unmarshalALPN(extData *cryptobyte.String) bool {
	// RFC 7301, Section 3.1
	var protoList cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&protoList) || protoList.Empty() {
		return false
	}
	for !protoList.Empty() {
		var proto cryptobyte.String
		if !protoList.ReadUint8LengthPrefixed(&proto) || proto.Empty() {
			return false
		}
		m.ALPNProtocols = append(m.ALPNProtocols, string(proto))
	}
	return true
}

func (m *ClientHelloInfo) unmarshalSCT(_ *cryptobyte.String) bool {
	// RFC 6962, Section 3.3.1
	m.scts = true
	return true
}

func (m *ClientHelloInfo) unmarshalSupportedVersions(extData *cryptobyte.String) bool {
	// RFC 8446, Section 4.2.1
	// TODO: Fix revive
	//revive:disable:unexported-naming
	var VersList cryptobyte.String
	//revive:enable:unexported-naming
	if !extData.ReadUint8LengthPrefixed(&VersList) || VersList.Empty() {
		return false
	}
	for !VersList.Empty() {
		// TODO: Fix revive
		//revive:disable:unexported-naming
		var Vers uint16
		//revive:enable:unexported-naming
		if !VersList.ReadUint16(&Vers) {
			return false
		}
		m.SupportedVersions = append(m.SupportedVersions, Vers)
	}
	return true
}

func (m *ClientHelloInfo) unmarshalCookie(extData *cryptobyte.String) bool {
	// RFC 8446, Section 4.2.2
	if !readUint16LengthPrefixed(extData, &m.cookie) ||
		len(m.cookie) == 0 {
		return false
	}
	return true
}

func (m *ClientHelloInfo) unmarshalKeyShare(extData *cryptobyte.String) bool {
	// RFC 8446, Section 4.2.8
	var clientShares cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&clientShares) {
		return false
	}
	for !clientShares.Empty() {
		var ks keyShare
		if !clientShares.ReadUint16((*uint16)(&ks.group)) ||
			!readUint16LengthPrefixed(&clientShares, &ks.data) ||
			len(ks.data) == 0 {
			return false
		}
		m.keyShares = append(m.keyShares, ks)
	}
	return true
}

func (m *ClientHelloInfo) unmarshalEarlyData(_ *cryptobyte.String) bool {
	// RFC 8446, Section 4.2.10
	m.earlyData = true
	return true
}

func (m *ClientHelloInfo) unmarshalPSKModes(extData *cryptobyte.String) bool {
	// RFC 8446, Section 4.2.9
	return readUint8LengthPrefixed(extData, &m.pskModes)
}

// revive:disable:cognitive-complexity
// revive:disable:cyclomatic
func (m *ClientHelloInfo) unmarshalPreSharedKey(extData *cryptobyte.String) bool {
	// revive:enable:cognitive-complexity
	// revive:enable:cyclomatic
	// RFC 8446, Section 4.2.11
	var identities cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&identities) || identities.Empty() {
		return false
	}
	for !identities.Empty() {
		var psk pskIdentity
		if !readUint16LengthPrefixed(&identities, &psk.label) ||
			!identities.ReadUint32(&psk.obfuscatedTicketAge) ||
			len(psk.label) == 0 {
			return false
		}
		m.pskIdentities = append(m.pskIdentities, psk)
	}

	var binders cryptobyte.String
	if !extData.ReadUint16LengthPrefixed(&binders) || binders.Empty() {
		return false
	}
	for !binders.Empty() {
		var binder []byte
		if !readUint8LengthPrefixed(&binders, &binder) ||
			len(binder) == 0 {
			return false
		}
		m.pskBinders = append(m.pskBinders, binder)
	}
	return true
}
