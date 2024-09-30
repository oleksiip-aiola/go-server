package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
)

// ParsedAttestation contains data extracted from attestationObject
type ParsedAttestation struct {
	AuthenticatorData []byte
	PublicKey         []byte
}

// Helper function to parse the attestation object and extract the public key
func parseAttestationObject(attestationObject []byte) ([]byte, []byte, error) {
	// Parse the CBOR attestationObject
	var attestationMap map[string]interface{}
	if err := cbor.Unmarshal(attestationObject, &attestationMap); err != nil {
		return nil, nil, errors.New("failed to parse attestation object")
	}
	authData, ok := attestationMap["authData"].([]byte)
	if !ok {
		return nil, nil, errors.New("authenticator data not found")
	}
	// Extract public key from authData (assuming COSE key format)
	publicKey, err := extractPublicKeyFromAuthData(authData)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}
	fmt.Println(publicKey)
	return authData, publicKey, nil
}

// ExtractPublicKeyFromAuthData extracts the public key from the COSE key in the authenticator data
func extractPublicKeyFromAuthData(authData []byte) ([]byte, error) {
	// Ensure the authData has enough length for basic fields (RP ID hash, flags, sign count)
	if len(authData) < 37 {
		return nil, errors.New("authData is too short")
	}
	// Step 1: Skip RP ID hash (32 bytes), flags (1 byte), sign count (4 bytes)
	offset := 37

	// Step 2: Extract the AAGUID (16 bytes)
	offset += 16

	// Step 3: Extract the credential ID length (2 bytes)
	credentialIDLength := binary.BigEndian.Uint16(authData[offset : offset+2])
	offset += 2

	// Step 4: Extract the credential ID (variable length, based on credentialIDLength)
	offset += int(credentialIDLength)

	// Step 5: Extract the COSE key (starts after credential ID)
	coseKey := authData[offset:]

	// Use a CBOR decoder to decode the COSE key
	var coseKeyMap map[int]interface{}
	if err := cbor.Unmarshal(coseKey, &coseKeyMap); err != nil {
		return nil, err
	}

	// Ensure that the COSE key contains 'x' and 'y' coordinates for ECDSA
	xCoord, ok := coseKeyMap[-2].([]byte)
	if !ok {
		return nil, errors.New("failed to extract x coordinate from COSE key")
	}
	yCoord, ok := coseKeyMap[-3].([]byte)
	if !ok {
		return nil, errors.New("failed to extract y coordinate from COSE key")
	}

	// Step 6: Construct the ECDSA public key using the x and y coordinates
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xCoord),
		Y:     new(big.Int).SetBytes(yCoord),
	}

	// Step 7: Marshal the public key to PKIX format for storage
	return x509.MarshalPKIXPublicKey(publicKey)
}

// extractAAGUID extracts the AAGUID from authenticator data
func extractAAGUID(authenticatorData []byte) []byte {
	// The AAGUID is usually found in the first 16 bytes of the authenticator data
	if len(authenticatorData) < 16 {
		return nil
	}
	return authenticatorData[:16]
}

// ExtractSignCount extracts the sign count from authenticatorData (assuming WebAuthn format)
func extractSignCount(authenticatorData []byte) uint32 {
	// The sign count is typically located in the last 4 bytes of the authenticatorData
	if len(authenticatorData) < 37 {
		return 0
	}
	return uint32(authenticatorData[33])<<24 | uint32(authenticatorData[34])<<16 | uint32(authenticatorData[35])<<8 | uint32(authenticatorData[36])
}

// // validateAttestation verifies the attestationObject's authenticity
// func validateAttestation(attestationObject, clientDataJSON, authenticatorData []byte) error {
// 	// Extract the attestation certificate and signature from the attestationObject (simplified)
// 	cert, signature, err := extractAttestationCertAndSignature(attestationObject)
// 	if err != nil {
// 		return err
// 	}

// 	// Parse and verify the certificate chain (optional for direct attestation)
// 	attestationCert, err := x509.ParseCertificate(cert)
// 	if err != nil {
// 		return errors.New("failed to parse attestation certificate")
// 	}

// 	// Verify the certificate chain (optional, depending on attestation type)
// 	// In direct attestation, this requires the manufacturer's root certificates.
// 	roots := x509.NewCertPool()
// 	// Add your trusted attestation roots here (you would get these from the authenticator manufacturers)
// 	if ok := roots.AppendCertsFromPEM([]byte("root_cert_pem_data")); !ok {
// 		return errors.New("failed to append root certificates")
// 	}
// 	opts := x509.VerifyOptions{Roots: roots}
// 	if _, err := attestationCert.Verify(opts); err != nil {
// 		return errors.New("failed to verify attestation certificate chain")
// 	}

// 	// Validate the signature using the public key from the attestation cert
// 	dataToVerify := append(authenticatorData, clientDataJSON...)
// 	hash := crypto.SHA256.New()
// 	hash.Write(dataToVerify)
// 	digest := hash.Sum(nil)

// 	// Check signature validity
// 	isValid := ecdsa.VerifyASN1(attestationCert.PublicKey.(*ecdsa.PublicKey), digest, signature)
// 	if !isValid {
// 		return errors.New("invalid attestation signature")
// 	}

// 	return nil
// }

// Placeholder function to extract attestation cert and signature
// func extractAttestationCertAndSignature(attestationObject []byte) ([]byte, []byte, error) {
// 	// Parse the attestation object to extract the certificate and signature
// 	// This is specific to the attestation format (e.g., Packed, TPM, etc.)
// 	// Simplified for demonstration purposes
// 	return []byte("certificate_data"), []byte("signature_data"), nil
// }

func verifySignature(publicKey []byte, authenticatorData, clientDataJSON, signature []byte) (bool, error) {
	// Concatenate authenticatorData and clientDataJSON
	clientDataHash := sha256.Sum256(clientDataJSON)

	// Step 2: Concatenate authenticatorData and the hashed clientDataJSON
	dataToVerify := append(authenticatorData, clientDataHash[:]...)

	// Step 3: Hash the concatenated data (authenticatorData + clientDataHash)
	digest := sha256.Sum256(dataToVerify) // Hash this concatenated data

	// Step 4: Parse the stored public key
	parsedKey, err := x509.ParsePKIXPublicKey(publicKey)
	if err != nil {
		return false, errors.New("failed to parse public key")
	}

	// Ensure the parsed key is an ECDSA key (assuming ES256 algorithm is used)
	ecdsaKey, ok := parsedKey.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("public key is not ECDSA")
	}

	// Step 5: Verify the signature using ECDSA and the digest
	signatureValid := ecdsa.VerifyASN1(ecdsaKey, digest[:], signature)
	if !signatureValid {
		return false, errors.New("invalid signature")
	}

	// Return true if the signature is valid
	return true, nil
}
