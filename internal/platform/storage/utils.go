package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CalculateContentHash calculates SHA256 hash of content bytes
func CalculateContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// GenerateS3Key generates S3 key for document content
// Format: {documentType}/{documentID}/{version}/content
// Example: definition/550e8400-e29b-41d4-a716-446655440000/1.0.0/content
func GenerateS3Key(documentType, documentID, version string) string {
	return fmt.Sprintf("%s/%s/%s/content", documentType, documentID, version)
}
