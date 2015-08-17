package gorethink

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func BenchmarkBatch200RandomWrites(b *testing.B) {

	var term Term
	var data []map[string]interface{}

	for i := 0; i < b.N; i++ {

		for is := 0; is < 200; is++ {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			cid := map[string]interface{}{
				"customer_id": strconv.FormatInt(r.Int63(), 10),
			}
			data = append(data, cid)
		}

		// Insert the new item into the database
		term = DB("benchmarks").Table("benchmarks").Insert(data)

		// Insert the new item into the database
		_, err := term.RunWrite(session, RunOpts{
			MinBatchRows: 200,
			MaxBatchRows: 200,
		})
		if err != nil {
			b.Errorf("insert failed [%s] ", err)
		}
	}

}

func BenchmarkBatch200RandomWritesParallel10(b *testing.B) {

	var term Term
	var data []map[string]interface{}

	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {
			for is := 0; is < 200; is++ {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				cid := map[string]interface{}{
					"customer_id": strconv.FormatInt(r.Int63(), 10),
				}
				data = append(data, cid)
			}

			// Insert the new item into the database
			term = DB("benchmarks").Table("benchmarks").Insert(data)

			// Insert the new item into the database
			_, err := term.RunWrite(session, RunOpts{
				MinBatchRows: 200,
				MaxBatchRows: 200,
			})
			if err != nil {
				b.Errorf("insert failed [%s] ", err)
			}
		}
	})

}

func BenchmarkBatch200SoftRandomWritesParallel10(b *testing.B) {

	var term Term
	var data []map[string]interface{}

	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {

			opts := InsertOpts{Durability: "soft"}

			for is := 0; is < 200; is++ {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				cid := map[string]interface{}{
					"customer_id": strconv.FormatInt(r.Int63(), 10),
				}
				data = append(data, cid)
			}

			// Insert the new item into the database
			term = DB("benchmarks").Table("benchmarks").Insert(data, opts)

			// Insert the new item into the database
			_, err := term.RunWrite(session, RunOpts{
				MinBatchRows: 200,
				MaxBatchRows: 200,
			})
			if err != nil {
				b.Errorf("insert failed [%s] ", err)
			}
		}
	})

}

func BenchmarkRandomWrites(b *testing.B) {

	for i := 0; i < b.N; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		data := map[string]interface{}{
			"customer_id": strconv.FormatInt(r.Int63(), 10),
		}
		// Insert the new item into the database
		_, err := DB("benchmarks").Table("benchmarks").Insert(data).RunWrite(session)
		if err != nil {
			b.Errorf("insert failed [%s] ", err)
		}
	}

}

func BenchmarkRandomWritesParallel10(b *testing.B) {

	// p*GOMAXPROCS
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			data := map[string]interface{}{
				"customer_id": strconv.FormatInt(r.Int63(), 10),
			}
			// Insert the new item into the database
			_, err := DB("benchmarks").Table("benchmarks").Insert(data).RunWrite(session)
			if err != nil {
				b.Errorf("insert failed [%s] ", err)
			}
		}
	})

}

func BenchmarkRandomSoftWrites(b *testing.B) {

	for i := 0; i < b.N; i++ {
		data := map[string]interface{}{
			"customer_id": strconv.FormatInt(rand.Int63(), 10),
		}
		// Insert the new item into the database
		opts := InsertOpts{Durability: "soft"}
		_, err := DB("benchmarks").Table("benchmarks").Insert(data, opts).RunWrite(session)
		if err != nil {
			b.Errorf("insert failed [%s] ", err)
		}
	}

}

func BenchmarkRandomSoftWritesParallel10(b *testing.B) {

	// p*GOMAXPROCS
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			data := map[string]interface{}{
				"customer_id": strconv.FormatInt(r.Int63(), 10),
			}

			// Insert the new item into the database
			opts := InsertOpts{Durability: "soft"}
			_, err := DB("benchmarks").Table("benchmarks").Insert(data, opts).RunWrite(session)
			if err != nil {
				b.Errorf("insert failed [%s] ", err)
			}
		}
	})

}

func BenchmarkSequentialWrites(b *testing.B) {

	si := 0
	for i := 0; i < b.N; i++ {
		si++
		data := map[string]interface{}{
			"customer_id": si,
		}

		// Insert the new item into the database
		_, err := DB("benchmarks").Table("benchmarks").Insert(data).RunWrite(session)
		if err != nil {
			b.Errorf("insert failed [%s] ", err)
			return
		}
	}
}

func BenchmarkSequentialWritesParallel10(b *testing.B) {

	var mu sync.Mutex
	si := 0

	// p*GOMAXPROCS
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			si++
			mu.Unlock()

			data := map[string]interface{}{
				"customer_id": si,
			}

			// Insert the new item into the database
			_, err := DB("benchmarks").Table("benchmarks").Insert(data).RunWrite(session)
			if err != nil {
				b.Errorf("insert failed [%s] ", err)
				return
			}
		}
	})

}

func BenchmarkSequentialSoftWrites(b *testing.B) {

	opts := InsertOpts{Durability: "soft"}
	si := 0

	for i := 0; i < b.N; i++ {
		si++
		data := map[string]interface{}{
			"customer_id": si,
		}

		// Insert the new item into the database
		_, err := Table("benchmarks").Insert(data, opts).RunWrite(session)
		if err != nil {
			b.Errorf("insert failed [%s] ", err)
			return
		}
	}
}

func BenchmarkSequentialSoftWritesParallel10(b *testing.B) {

	var mu sync.Mutex
	si := 0

	// p*GOMAXPROCS
	b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			si++
			mu.Unlock()

			data := map[string]interface{}{
				"customer_id": si,
			}

			opts := InsertOpts{Durability: "soft"}

			// Insert the new item into the database
			_, err := Table("benchmarks").Insert(data, opts).RunWrite(session)
			if err != nil {
				b.Errorf("insert failed [%s] ", err)
				return
			}
		}
	})

}
