package cache

import (
	"context"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Отключаем логирование во время тестов
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	m.Run()
}

// Определяем mockCache, который реализует интерфейс Cache без метода Close
type mockCache[K comparable, V any] struct{}

func (m *mockCache[K, V]) Set(ctx context.Context, key K, value V, ttl ...time.Duration) {}

func (m *mockCache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	var zero V
	return zero, false
}

func (m *mockCache[K, V]) Delete(ctx context.Context, key K) {}

func (m *mockCache[K, V]) Flush(ctx context.Context) {}

// функция для создания нового кэша LRU для тестирования
func newTestLRUCache[K comparable, V any](maxEntries int, defaultTTL, cleanupInterval time.Duration) *lruCache[K, V] {
	return NewLRUCache[K, V](maxEntries, defaultTTL, cleanupInterval).(*lruCache[K, V])
}

// функция для создания нового кэша LFU для тестирования
func newTestLFUCache[K comparable, V any](maxEntries int, defaultTTL, cleanupInterval time.Duration) *lfuCache[K, V] {
	return NewLFUCache[K, V](maxEntries, defaultTTL, cleanupInterval).(*lfuCache[K, V])
}

// Тесты LRU

func TestLRUCache_SetGet(t *testing.T) {
	t.Run("Basic Set and Get", func(t *testing.T) {
		cache := newTestLRUCache[string, string](2, 5*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1")
		val, found := cache.Get(ctx, "key1")
		require.True(t, found)
		assert.Equal(t, "value1", val)
	})

	t.Run("Overwrite Existing Key", func(t *testing.T) {
		cache := newTestLRUCache[string, string](2, 5*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1")
		cache.Set(ctx, "key1", "value2")

		val, found := cache.Get(ctx, "key1")
		require.True(t, found)
		assert.Equal(t, "value2", val)
	})

	t.Run("Custom TTL", func(t *testing.T) {
		cache := newTestLRUCache[string, string](2, 5*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1", 1*time.Second)
		time.Sleep(2 * time.Second)

		_, found := cache.Get(ctx, "key1")
		assert.False(t, found)
	})

	t.Run("Default TTL", func(t *testing.T) {
		cache := newTestLRUCache[string, string](2, 2*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1")
		time.Sleep(3 * time.Second)

		_, found := cache.Get(ctx, "key1")
		assert.False(t, found)
	})
}

func TestLRUCache_Eviction(t *testing.T) {
	t.Run("Evict Least Recently Used", func(t *testing.T) {
		cache := newTestLRUCache[string, string](2, 5*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1")
		cache.Set(ctx, "key2", "value2")

		// Access key1 to make it most recently used
		cache.Get(ctx, "key1")

		cache.Set(ctx, "key3", "value3") // Should evict key2

		_, found := cache.Get(ctx, "key2")
		assert.False(t, found, "key2 should have been evicted")

		val, found := cache.Get(ctx, "key1")
		require.True(t, found)
		assert.Equal(t, "value1", val)

		val, found = cache.Get(ctx, "key3")
		require.True(t, found)
		assert.Equal(t, "value3", val)
	})
}

func TestLRUCache_Delete(t *testing.T) {
	cache := newTestLRUCache[string, string](2, 5*time.Second, 1*time.Second)
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1")
	cache.Delete(ctx, "key1")

	_, found := cache.Get(ctx, "key1")
	assert.False(t, found, "key1 should have been deleted")
}

func TestLRUCache_Flush(t *testing.T) {
	cache := newTestLRUCache[string, string](3, 5*time.Second, 1*time.Second)
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1")
	cache.Set(ctx, "key2", "value2")
	cache.Set(ctx, "key3", "value3")

	cache.Flush(ctx)

	_, found := cache.Get(ctx, "key1")
	assert.False(t, found, "key1 should have been flushed")
	_, found = cache.Get(ctx, "key2")
	assert.False(t, found, "key2 should have been flushed")
	_, found = cache.Get(ctx, "key3")
	assert.False(t, found, "key3 should have been flushed")
}

func TestLRUCache_TTLExpiration(t *testing.T) {
	cache := newTestLRUCache[string, string](2, 1*time.Second, 500*time.Millisecond)
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1")
	time.Sleep(2 * time.Second)

	_, found := cache.Get(ctx, "key1")
	assert.False(t, found, "key1 should have expired")
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache := newTestLRUCache[int, int](10000, 5*time.Second, 1*time.Second)
	defer cache.Close()

	ctx := context.Background()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cache.Set(ctx, id*1000+j, id*1000+j)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cache.Get(ctx, id*1000+j)
			}
		}(i)
	}

	wg.Wait()

	val, found := cache.Get(ctx, 5000)
	assert.True(t, found)
	assert.Equal(t, 5000, val)
}

// Тесты LFU

func TestLFUCache_SetGet(t *testing.T) {
	t.Run("Basic Set and Get", func(t *testing.T) {
		cache := newTestLFUCache[string, string](2, 5*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1")
		val, found := cache.Get(ctx, "key1")
		require.True(t, found)
		assert.Equal(t, "value1", val)
	})

	t.Run("Custom TTL", func(t *testing.T) {
		cache := newTestLFUCache[string, string](2, 5*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1", 1*time.Second)
		time.Sleep(2 * time.Second)

		_, found := cache.Get(ctx, "key1")
		assert.False(t, found, "key1 should have expired")
	})
}

func TestLFUCache_Eviction(t *testing.T) {
	t.Run("Evict Least Frequently Used", func(t *testing.T) {
		cache := newTestLFUCache[string, string](2, 10*time.Second, 1*time.Second)
		defer cache.Close()

		ctx := context.Background()

		cache.Set(ctx, "key1", "value1")
		cache.Set(ctx, "key2", "value2")

		cache.Get(ctx, "key1")
		cache.Get(ctx, "key1")

		cache.Set(ctx, "key3", "value3")

		_, found := cache.Get(ctx, "key2")
		assert.False(t, found, "key2 should have been evicted")

		val, found := cache.Get(ctx, "key1")
		require.True(t, found)
		assert.Equal(t, "value1", val)

		val, found = cache.Get(ctx, "key3")
		require.True(t, found)
		assert.Equal(t, "value3", val)
	})
}

func TestLFUCache_Delete(t *testing.T) {
	cache := newTestLFUCache[string, string](2, 10*time.Second, 1*time.Second)
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1")
	cache.Delete(ctx, "key1")

	_, found := cache.Get(ctx, "key1")
	assert.False(t, found, "key1 should have been deleted")
}

func TestLFUCache_Flush(t *testing.T) {
	cache := newTestLFUCache[string, string](3, 10*time.Second, 1*time.Second)
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1")
	cache.Set(ctx, "key2", "value2")
	cache.Set(ctx, "key3", "value3")

	cache.Flush(ctx)

	_, found := cache.Get(ctx, "key1")
	assert.False(t, found, "key1 should have been flushed")
	_, found = cache.Get(ctx, "key2")
	assert.False(t, found, "key2 should have been flushed")
	_, found = cache.Get(ctx, "key3")
	assert.False(t, found, "key3 should have been flushed")
}

func TestLFUCache_TTLExpiration(t *testing.T) {
	cache := newTestLFUCache[string, string](2, 1*time.Second, 500*time.Millisecond)
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", "value1")
	time.Sleep(2 * time.Second)

	_, found := cache.Get(ctx, "key1")
	assert.False(t, found, "key1 should have expired")
}

func TestLFUCache_ConcurrentAccess(t *testing.T) {
	cache := newTestLFUCache[int, int](10000, 5*time.Second, 1*time.Second)
	defer cache.Close()

	ctx := context.Background()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cache.Set(ctx, id*1000+j, id*1000+j)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cache.Get(ctx, id*1000+j)
			}
		}(i)
	}

	wg.Wait()

	val, found := cache.Get(ctx, 5000)
	assert.True(t, found)
	assert.Equal(t, 5000, val)
}

// Дополнительные тесты

func TestCache_NewCacheInvalidStrategy(t *testing.T) {
	config := CacheConfig{
		Strategy:        CacheStrategy(999), // Invalid strategy
		MaxEntries:      100,
		DefaultTTL:      5 * time.Second,
		CleanupInterval: 1 * time.Second,
	}

	_, err := NewCache[string, string](config)
	assert.Error(t, err, "should return error for invalid cache strategy")
}

func TestCache_CloseCacheUnknownType(t *testing.T) {
	var c interfaces.Cache[string, string] = &mockCache[string, string]{}

	CloseCache(c)
}
