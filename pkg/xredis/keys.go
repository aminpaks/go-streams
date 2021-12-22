package xredis

import "fmt"

func sortedQueueProcessingReferenceKey(queue string) string {
	return fmt.Sprintf("sortedQueue::%s::processing::reference", queue)
}
func sortedQueueProcessingPriorityKey(queue string) string {
	return fmt.Sprintf("sortedQueue::%s::processing::priory", queue)
}
