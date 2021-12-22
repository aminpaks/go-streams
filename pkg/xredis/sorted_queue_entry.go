package xredis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func EnqueueSortedEntry(client *RedisClient, ctx context.Context, queue string, entry XSortedQueueEntry) error {
	return enqueueSortedEntryWithPayload(client, ctx, queue, entry, 0)
}
func enqueueSortedEntry(client *RedisClient, ctx context.Context, queue string, referenceUri string, priority float64) error {
	if err := client.ZAddNX(ctx, queue, &redis.Z{
		Score:  priority,
		Member: referenceUri,
	}).Err(); err != nil {
		return err
	}
	return nil
}
func enqueueSortedEntryWithPayload(client *RedisClient, ctx context.Context, queue string, entry XSortedQueueEntry, retry int) error {
	if err := client.Set(ctx, entry.ReferenceUri, serializedXSortedQueueEntry(entry, retry), entry.Expiration).Err(); err != nil {
		return err
	}
	if err := enqueueSortedEntry(client, ctx, queue, entry.ReferenceUri, entry.Priority); err != nil {
		return err
	}

	return nil
}
func markSortedEntryForProcessing(client *RedisClient, queue string, entries ...XSortedQueueEntry) error {
	if err := client.SAdd(context.Background(), sortedQueueProcessingReferenceKey(queue), strToInterface(getRefs(entries...)...)...).Err(); err != nil {
		return err
	}
	for _, e := range entries {
		if err := client.HSetNX(context.Background(), sortedQueueProcessingPriorityKey(queue), e.ReferenceUri, e.Priority).Err(); err != nil {
			return err
		}
	}
	return nil
}
func ackSortedQueueEntry(client *RedisClient, queue string, entry XSortedQueueEntry) error {
	ctx := context.Background()
	if err := client.SRem(ctx, sortedQueueProcessingReferenceKey(queue), strToInterface(getRefs(entry)...)...).Err(); err != nil {
		return err
	}
	if err := client.HDel(ctx, sortedQueueProcessingPriorityKey(queue), getRefs(entry)...).Err(); err != nil {
		return err
	}
	return nil
}
func getRefs(entries ...XSortedQueueEntry) []string {
	refs := []string{}
	for _, e := range entries {
		refs = append(refs, e.ReferenceUri)
	}
	return refs
}
func strToInterface(v ...string) []interface{} {
	i := []interface{}{}
	for _, s := range v {
		i = append(i, s)
	}
	return i
}

func cleanSortedQueueEntry(client *RedisClient, e XSortedQueueEntry) error {
	ctx := context.Background()
	if err := client.SRem(ctx, sortedQueueProcessingReferenceKey(e.queue), strToInterface(e.ReferenceUri)[0]).Err(); err != nil {
		return fmt.Errorf("failed to clean up processing reference of sorted queue entry %s: %v", e.ReferenceUri, err)
	}
	if err := client.HDel(ctx, sortedQueueProcessingPriorityKey(e.queue), e.ReferenceUri).Err(); err != nil {
		return fmt.Errorf("failed to clean up processing priority of sorted queue entry %s: %v", e.ReferenceUri, err)
	}
	if err := client.Del(ctx, e.ReferenceUri).Err(); err != nil {
		return fmt.Errorf("failed to clean up sorted queue entry %s: %v", e.ReferenceUri, err)
	}
	return nil
}

func getXSortedQueueEntryByUri(client *RedisClient, uri interface{}) (string, error) {
	_uri, ok := uri.(string)
	if !ok {
		return "", nil
	}

	if !IsValidUri(_uri) {
		return "", fmt.Errorf("invalid URI: %v", uri)
	}

	v, err := client.Get(context.Background(), _uri).Result()
	if err != nil {
		if err == redis.Nil {
			err = fmt.Errorf("failed to read from URI %v", uri)
		}
		return "", err
	}
	return v, nil
}

func parseXSortedQueueEntry(i string) (*XSortedQueueEntry, error) {
	var e internalPersistedXSortedQueueEntry
	if err := json.Unmarshal([]byte(i), &e); err != nil {
		return nil, fmt.Errorf("failed to parse entry: %v", err)
	}
	if e.Failures == nil {
		e.Failures = make([]string, 0)
	}
	return &XSortedQueueEntry{
		CurrentRetries: e.Retries,
		Value:          e.Value,
		Priority:       e.Priority,
		ReferenceUri:   e.ReferenceUri,
		Failures:       e.Failures,
	}, nil
}
