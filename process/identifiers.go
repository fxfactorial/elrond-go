package process

import (
	"fmt"
)

// ShardCacherIdentifier generates a string identifier between 2 shards
func ShardCacherIdentifier(senderShardId uint32, destinationShardId uint32) string {
	if senderShardId == destinationShardId {
		return fmt.Sprintf("%d", senderShardId)
	}

	return fmt.Sprintf("%d_%d", senderShardId, destinationShardId)
}
