package cache

import (
	"container/list"
	"context"
	"sync"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"
)

// lruCache представляет кэш со стратегией LRU
type lruCache[K comparable, V any] struct {
	mu              sync.RWMutex
	maxEntries      int
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	cache           map[K]*list.Element
	lruList         *list.List
	cleanupTicker   *time.Ticker
	stopCleanup     chan struct{}
}

// entry представляет элемент в LRU кэше
type entry[K comparable, V any] struct {
	key    K
	value  V
	expiry time.Time
}

// NewLRUCache создает новый LRU кэш
func NewLRUCache[K comparable, V any](maxEntries int, defaultTTL, cleanupInterval time.Duration) interfaces.Cache[K, V] {
	c := &lruCache[K, V]{
		maxEntries:      maxEntries,
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		cache:           make(map[K]*list.Element),
		lruList:         list.New(),
		cleanupTicker:   time.NewTicker(cleanupInterval),
		stopCleanup:     make(chan struct{}),
	}

	go c.cleanupExpired()

	return c
}

// Set добавляет или обновляет элемент в кэше
func (c *lruCache[K, V]) Set(ctx context.Context, key K, value V, ttl ...time.Duration) {
	var itemTTL time.Duration
	if len(ttl) > 0 {
		itemTTL = ttl[0]
	} else {
		itemTTL = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.lruList.MoveToFront(elem)
		elem.Value.(*entry[K, V]).value = value
		elem.Value.(*entry[K, V]).expiry = time.Now().Add(itemTTL)
		return
	}

	newEntry := &entry[K, V]{
		key:    key,
		value:  value,
		expiry: time.Now().Add(itemTTL),
	}
	elem := c.lruList.PushFront(newEntry)
	c.cache[key] = elem

	if c.lruList.Len() > c.maxEntries {
		c.evict()
	}
}

// Get возвращает значение элемента по ключу и обновляет его позицию в списке
func (c *lruCache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	var zero V

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		if time.Now().After(elem.Value.(*entry[K, V]).expiry) {
			c.removeElement(elem)
			return zero, false
		}
		c.lruList.MoveToFront(elem)
		return elem.Value.(*entry[K, V]).value, true
	}
	//Возвращает zero value и false, если элемент не найден или истек по TTL
	return zero, false
}

// Delete удаляет элемент из кэша по ключу
func (c *lruCache[K, V]) Delete(ctx context.Context, key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.removeElement(elem)
	}
}

// Flush очищает весь кэш
func (c *lruCache[K, V]) Flush(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lruList.Init()
	c.cache = make(map[K]*list.Element)
}

// evict удаляет наименее недавно использованный элемент из кэша
func (c *lruCache[K, V]) evict() {
	elem := c.lruList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

// removeElement удаляет элемент из списка и мапы
func (c *lruCache[K, V]) removeElement(elem *list.Element) {
	c.lruList.Remove(elem)
	entry := elem.Value.(*entry[K, V])
	delete(c.cache, entry.key)
}

func (c *lruCache[K, V]) cleanupExpired() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.mu.Lock()
			now := time.Now()
			for _, elem := range c.cache {
				if now.After(elem.Value.(*entry[K, V]).expiry) {
					c.removeElement(elem)
				}
			}
			c.mu.Unlock()
		case <-c.stopCleanup:
			return
		}
	}
}

// Close останавливает процесс очистки и освобождает ресурсы.
func (c *lruCache[K, V]) Close() {
	c.cleanupTicker.Stop()
	close(c.stopCleanup)
}
