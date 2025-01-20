package cache

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"
)

// lfuCache представляет LFU кэш
type lfuCache[K comparable, V any] struct {
	mu              sync.RWMutex
	maxEntries      int
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	cache           map[K]*lfuEntry[K, V]
	heap            *lfuHeap[K, V]
	cleanupTicker   *time.Ticker
	stopCleanup     chan struct{}
}

// lfuEntry представляет элемент в LFU кэше
type lfuEntry[K comparable, V any] struct {
	key       K
	value     V
	frequency int
	expiry    time.Time
	index     int
}

type lfuHeap[K comparable, V any] []*lfuEntry[K, V]

func (h lfuHeap[K, V]) Len() int { return len(h) }

func (h lfuHeap[K, V]) Less(i, j int) bool {
	return h[i].frequency < h[j].frequency
}

func (h lfuHeap[K, V]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *lfuHeap[K, V]) Push(x any) {
	n := len(*h)
	entry := x.(*lfuEntry[K, V])
	entry.index = n
	*h = append(*h, entry)
}

func (h *lfuHeap[K, V]) Pop() any {
	old := *h
	n := len(old)
	entry := old[n-1]
	entry.index = -1
	*h = old[0 : n-1]
	return entry
}

// NewLFUCache создает новый LFU кэш
func NewLFUCache[K comparable, V any](maxEntries int, defaultTTL, cleanupInterval time.Duration) interfaces.Cache[K, V] {
	h := &lfuHeap[K, V]{}
	heap.Init(h)

	c := &lfuCache[K, V]{
		maxEntries:      maxEntries,
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		cache:           make(map[K]*lfuEntry[K, V]),
		heap:            h,
		cleanupTicker:   time.NewTicker(cleanupInterval),
		stopCleanup:     make(chan struct{}),
	}

	go c.cleanupExpired()

	return c
}

// Set добавляет или обновляет элемент в кэше
func (c *lfuCache[K, V]) Set(ctx context.Context, key K, value V, ttl ...time.Duration) {
	var itemTTL time.Duration
	if len(ttl) > 0 {
		itemTTL = ttl[0]
	} else {
		itemTTL = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		elem.value = value
		elem.expiry = time.Now().Add(itemTTL)
		elem.frequency++
		heap.Fix(c.heap, elem.index)
		return
	}

	newEntry := &lfuEntry[K, V]{
		key:       key,
		value:     value,
		frequency: 1,
		expiry:    time.Now().Add(itemTTL),
	}
	heap.Push(c.heap, newEntry)
	c.cache[key] = newEntry

	if len(c.cache) > c.maxEntries {
		c.evict()
	}
}

// Get возвращает значение элемента по ключу
func (c *lfuCache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	var zero V

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		if time.Now().After(elem.expiry) {
			c.removeElement(elem)
			return zero, false
		}
		elem.frequency++
		heap.Fix(c.heap, elem.index)
		return elem.value, true
	}

	return zero, false
}

// Delete удаляет элемент из кэша по ключу
func (c *lfuCache[K, V]) Delete(ctx context.Context, key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.removeElement(elem)
	}
}

// Flush очищает весь кэш
func (c *lfuCache[K, V]) Flush(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[K]*lfuEntry[K, V])
	c.heap = &lfuHeap[K, V]{}
	heap.Init(c.heap)
}

func (c *lfuCache[K, V]) evict() {
	if c.heap.Len() == 0 {
		return
	}
	elem := heap.Pop(c.heap).(*lfuEntry[K, V])
	delete(c.cache, elem.key)
	log.Printf("Evicting key: %v due to cache size limit", elem.key)
}

func (c *lfuCache[K, V]) removeElement(elem *lfuEntry[K, V]) {
	heap.Remove(c.heap, elem.index)
	delete(c.cache, elem.key)
	log.Printf("Removing key: %v due to expiration or deletion", elem.key)
}

func (c *lfuCache[K, V]) cleanupExpired() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.mu.Lock()
			now := time.Now()
			for _, elem := range c.cache {
				if now.After(elem.expiry) {
					c.removeElement(elem)
				}
			}
			c.mu.Unlock()
		case <-c.stopCleanup:
			return
		}
	}
}

// Close останавливает процесс очистки и освобождает ресурсы
func (c *lfuCache[K, V]) Close() {
	c.cleanupTicker.Stop()
	close(c.stopCleanup)
}
