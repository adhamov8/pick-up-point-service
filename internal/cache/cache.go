/*
Package cache предоставляет универсальный in-memory кэш с поддержкой различных стратегий.

Поддерживаемые стратегии:
- LRU (Least Recently Used): Удаляет наименее недавно использованные элементы при достижении максимального размера.
- LFU (Least Frequently Used): Удаляет наименее часто используемые элементы при достижении максимального размера.

Конфигурация кэша:
- Strategy: Выбор стратегии кэширования (LRUStrategy, LFUStrategy).
- MaxEntries: Максимальное количество элементов в кэше.
- DefaultTTL: Дефолтное время жизни элемента в кэше.
- CleanupInterval: Интервал очистки просроченных элементов.
*/
package cache

import (
	"fmt"
	"log"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"
)

// CacheStrategy определяет стратегии кэширования
type CacheStrategy int

const (
	LRUStrategy CacheStrategy = iota
	LFUStrategy
)

// CacheConfig содержит конфигурационные параметры для кэша
type CacheConfig struct {
	Strategy        CacheStrategy
	MaxEntries      int
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
}

// NewCache создает новый экземпляр кэша в соответствии с заданной конфигурацией
//
// Возвращает интерфейс Cache и ошибку, если стратегия не поддерживается
func NewCache[K comparable, V any](config CacheConfig) (interfaces.Cache[K, V], error) {
	switch config.Strategy {
	case LRUStrategy:
		return NewLRUCache[K, V](config.MaxEntries, config.DefaultTTL, config.CleanupInterval), nil
	case LFUStrategy:
		return NewLFUCache[K, V](config.MaxEntries, config.DefaultTTL, config.CleanupInterval), nil
	default:
		err := fmt.Errorf("unsupported cache strategy: %v", config.Strategy)
		log.Println(err)
		return nil, err
	}
}

// CloseCache закрывает кэш и освобождает связанные с ним ресурсы
//
// Если кэш имеет метод Close, он будет вызван. Если тип кэша неизвестен, функция выведет предупреждение
func CloseCache[K comparable, V any](c interfaces.Cache[K, V]) {
	switch cacheInstance := c.(type) {
	case *lruCache[K, V]:
		cacheInstance.Close()
	case *lfuCache[K, V]:
		cacheInstance.Close()
	default:
		log.Printf("Warning: Cache instance of unknown type %T cannot be closed", cacheInstance)
	}
}
