package cqrs

import "fmt"

func generateMessagesTableName(topic string) string {
	// Use a shorter prefix for message queue tables
	prefix := "mq_"

	// Maximum PostgreSQL identifier length is 63 bytes
	maxLength := 63

	if len(prefix)+len(topic) <= maxLength {
		return prefix + topic
	}

	// For long topics, include the start and a suffix with length
	// This ensures uniqueness for topics of different lengths
	suffixFormat := "_len%d"
	suffix := fmt.Sprintf(suffixFormat, len(topic))

	// Calculate maximum topic prefix length
	maxPrefixLength := maxLength - len(prefix) - len(suffix)

	// Truncate the topic and add the length suffix
	return prefix + topic[:maxPrefixLength] + suffix
}

func generateMessagesOffsetsTableName(topic string) string {
	// Use a shorter prefix for offsets queue tables
	prefix := "mq_offsets_"

	// Maximum PostgreSQL identifier length is 63 bytes
	maxLength := 63

	if len(prefix)+len(topic) <= maxLength {
		return prefix + topic
	}

	// For long topics, include the start and a suffix with length
	// This ensures uniqueness for topics of different lengths
	suffixFormat := "_len%d"
	suffix := fmt.Sprintf(suffixFormat, len(topic))

	// Calculate maximum topic prefix length
	maxPrefixLength := maxLength - len(prefix) - len(suffix)

	// Truncate the topic and add the length suffix
	return prefix + topic[:maxPrefixLength] + suffix
}
