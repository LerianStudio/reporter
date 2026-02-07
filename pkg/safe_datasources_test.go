// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pkg

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDataSources_NewSafeDataSources(t *testing.T) {
	t.Parallel()

	initial := map[string]DataSource{
		"ds1": {DatabaseType: PostgreSQLType, Initialized: true},
		"ds2": {DatabaseType: MongoDBType, Initialized: false},
	}

	sds := NewSafeDataSources(initial)
	require.NotNil(t, sds)

	ds, exists := sds.Get("ds1")
	assert.True(t, exists)
	assert.Equal(t, PostgreSQLType, ds.DatabaseType)
	assert.True(t, ds.Initialized)
}

func TestSafeDataSources_Get_NotFound(t *testing.T) {
	t.Parallel()

	sds := NewSafeDataSources(map[string]DataSource{})

	ds, exists := sds.Get("nonexistent")
	assert.False(t, exists)
	assert.Equal(t, DataSource{}, ds)
}

func TestSafeDataSources_Set(t *testing.T) {
	t.Parallel()

	sds := NewSafeDataSources(map[string]DataSource{})

	ds := DataSource{
		DatabaseType: PostgreSQLType,
		Initialized:  true,
		Status:       "available",
	}
	sds.Set("ds1", ds)

	got, exists := sds.Get("ds1")
	assert.True(t, exists)
	assert.Equal(t, PostgreSQLType, got.DatabaseType)
	assert.True(t, got.Initialized)
}

func TestSafeDataSources_GetAll_ReturnsShallowCopy(t *testing.T) {
	t.Parallel()

	initial := map[string]DataSource{
		"ds1": {DatabaseType: PostgreSQLType},
		"ds2": {DatabaseType: MongoDBType},
	}

	sds := NewSafeDataSources(initial)

	snapshot := sds.GetAll()
	assert.Len(t, snapshot, 2)

	// Modifying the snapshot should NOT affect the internal map
	snapshot["ds3"] = DataSource{DatabaseType: "fake"}
	_, exists := sds.Get("ds3")
	assert.False(t, exists, "modifying snapshot must not affect SafeDataSources internal map")
}

func TestSafeDataSources_Len(t *testing.T) {
	t.Parallel()

	sds := NewSafeDataSources(map[string]DataSource{
		"ds1": {},
		"ds2": {},
	})

	assert.Equal(t, 2, sds.Len())
}

func TestSafeDataSources_ConcurrentAccess_NoPanic(t *testing.T) {
	t.Parallel()

	sds := NewSafeDataSources(map[string]DataSource{
		"ds1": {DatabaseType: PostgreSQLType, Initialized: true},
		"ds2": {DatabaseType: MongoDBType, Initialized: false},
	})

	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // readers, writers, iterators

	// Concurrent readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_, _ = sds.Get("ds1")
				_, _ = sds.Get("ds2")
				_, _ = sds.Get("nonexistent")
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sds.Set("ds1", DataSource{
					DatabaseType: PostgreSQLType,
					Initialized:  j%2 == 0,
					Status:       "available",
				})
			}
		}(i)
	}

	// Concurrent iterators (GetAll)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = sds.GetAll()
			}
		}()
	}

	// If this completes without panic or data race, the test passes.
	wg.Wait()
}

func TestSafeDataSources_ConcurrentReadWrite_RaceDetector(t *testing.T) {
	// This test is specifically designed to trigger the race detector (-race flag).
	// If ExternalDataSources is a plain map, this WILL fail under -race.
	t.Parallel()

	sds := NewSafeDataSources(map[string]DataSource{
		"ds1": {DatabaseType: PostgreSQLType},
	})

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Half goroutines read
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				ds, _ := sds.Get("ds1")
				_ = ds.DatabaseType
			}
		}()
	}

	// Half goroutines write
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				sds.Set("ds1", DataSource{
					DatabaseType: MongoDBType,
					Initialized:  true,
				})
			}
		}(i)
	}

	wg.Wait()
}
