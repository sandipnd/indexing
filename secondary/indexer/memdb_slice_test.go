package indexer

import (
	"flag"
	"fmt"
	"github.com/couchbase/indexing/secondary/common"
	"github.com/couchbase/indexing/secondary/logging"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

var N *int
var isPrimary *bool
var lockThreads *bool

func init() {
	N = flag.Int("n", 10000000, "total number of docs")
	isPrimary = flag.Bool("primary", false, "Is primary index")
	lockThreads = flag.Bool("lockThreads", false, "Lock worker goroutines to a thread")
	flag.Parse()
	logging.SetLogLevel(logging.Info)
}

const keySize = 25
const snapIncrInterval = time.Millisecond * 10
const snapInitInterval = time.Millisecond * 10

type ientry struct {
	e     []byte
	docid []byte
	m     *MutationMeta
}

func randString(r *rand.Rand, n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = alphanum[r.Intn(len(alphanum))]
	}
	return string(bytes)
}

func mutationProducer(wg *sync.WaitGroup, s Slice, offset, n, id int, isRand bool, stream chan *ientry) {
	defer wg.Done()

	if *lockThreads {
		runtime.LockOSThread()
	}

	rnd := rand.New(rand.NewSource(int64(rand.Int())))
	for i := 0; i < n; i++ {
		docN := i + offset
		if isRand {
			docN = rnd.Int()%n + offset
		}

		docid := []byte(fmt.Sprintf("docid-%d", docN))
		key := []byte("[\"" + randString(rnd, keySize) + "\"]")
		entry, err := GetIndexEntryBytesFromKey(key, docid, false)
		common.CrashOnError(err)

		meta := NewMutationMeta()
		meta.vbucket = Vbucket(id)

		stream <- &ientry{e: entry, m: meta, docid: docid}

	}
}

func flushWorker(wg *sync.WaitGroup, stream chan *ientry, n int, slice Slice) {
	defer wg.Done()

	if *lockThreads {
		runtime.LockOSThread()
	}

	for i := 0; i < n; i++ {
		entry := <-stream
		slice.Insert(entry.e, entry.docid, entry.m)
		entry.m.Free()
	}
}

func runFlusher(interval time.Duration, streams []chan *ientry, slice Slice, finch chan bool) {
	var snap Snapshot
	var wg sync.WaitGroup

	for {
		for _, ch := range streams {
			n := len(ch)
			wg.Add(1)
			go flushWorker(&wg, ch, n, slice)
		}

		wg.Wait()

		info, err := slice.NewSnapshot(nil, false)
		common.CrashOnError(err)
		if snap != nil {
			snap.Close()
		}
		snap, err = slice.OpenSnapshot(info)
		common.CrashOnError(err)

		select {
		case <-time.After(interval):
		case <-finch:
			return
		}
	}
}

func TestMemDBInsertionPerf(t *testing.T) {
	var wg sync.WaitGroup
	finch := make(chan bool)
	nw := runtime.GOMAXPROCS(0)
	nPerWriter := *N / nw
	streams := make([]chan *ientry, nw)
	stats := &IndexStats{}
	stats.Init()
	cfg := common.SystemConfig.SectionConfig("indexer.", true)
	cfg.SetValue("numSliceWriters", nw)

	slice, err := NewMemDBSlice("/tmp/mdbslice",
		SliceId(0), common.IndexDefnId(0), common.IndexInstId(0), *isPrimary,
		cfg, stats)
	common.CrashOnError(err)

	// Initial build
	t1 := time.Now()
	for i := 0; i < nw; i++ {
		wg.Add(1)
		streams[i] = make(chan *ientry, 500000)
		if i == nw-1 {
			nPerWriter = *N - nPerWriter*i
		}

		go mutationProducer(&wg, slice, i*nPerWriter, nPerWriter, i, false, streams[i])
	}

	go func() {
		wg.Wait()
		finch <- true
	}()

	runFlusher(snapInitInterval, streams, slice, finch)
	dur1 := time.Since(t1)

	// Incremental update
	t2 := time.Now()
	for i := 0; i < nw; i++ {
		wg.Add(1)
		go mutationProducer(&wg, slice, i*nPerWriter, nPerWriter, i, false, streams[i])
	}

	go func() {
		wg.Wait()
		finch <- true
	}()

	runFlusher(snapIncrInterval, streams, slice, finch)
	dur2 := time.Since(t2)
	fmt.Printf("Initial build: %d items took %v -> %v items/s\n", *N, dur1, float64(*N)/dur1.Seconds())
	fmt.Printf("Incr build: %d items took %v -> %v items/s\n", *N, dur2, float64(*N)/dur2.Seconds())
	fmt.Println("Main Index:", slice.mainstore.DumpStats())
	if !*isPrimary {
		for i := 0; i < slice.numWriters; i++ {
			fmt.Println("Back Index", i, ":", slice.back[i].Stats())
		}
	}
}