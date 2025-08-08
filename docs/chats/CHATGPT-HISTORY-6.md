# CHAT HISTORY #6 — Batching and Goroutines strategy


You said:
Regarding the Go-based app we are writing to sync Gmail messages to a Sqlite database and also to archive them (meaning to delete emails that have been synced) I am trying to come up with a criteria and strategy for how to batch messages per API request.

I think we want to request the number of messages in batch that is optimally performant, but we also don't want batches to be so big that is anything happens to cause the sync to fail that we don't have to redo the batch.

I am also wanting to decide if we should use Goroutines for concurrency and how to determine the number of Goroutines that are optimal for performance with potential concern that Gmail might throttle our app or flag it for too much or too rapid of use.

And if we use Goroutines then we have to worry about some batches failing while other batches succeed. and being out of sync chronologically because we are thinking that once we sync an account we won't need to resync any of the older messages, that we will only have to pull down newer messages (we'll have a -force to resync the entire account, but we expect to rarely have to use that.)

Can you help me decide about how to optimize batch sizes, and since I assume we'll want to do so dynamically it will be helpful if you can present algorithms and even Go code should how we might approach it.

Thought for 1m 6s


ChatGPT said:
Awesome problem. Here’s a battle-tested way to choose batch sizes, run with goroutines safely, and keep the sync resilient and restartable.

TL;DR (what to implement)
Slice by time (month/week) and page within each slice. This bounds blast radius if anything fails and makes resumability trivial. ARCHIVING

Discover IDs in pages (e.g., maxResults 200–500), then fetch bodies concurrently with a bounded worker pool + rate limiter. ARCHIVING

Write to SQLite in small transactions (e.g., 50–200 msgs/tx) and only delete from Gmail in batches after the DB commit succeeds.

Dynamic tuning: start conservative (e.g., 6–8 workers, 3–5 RPS), use AIMD (additive-increase/multiplicative-decrease) based on latency and 429/403 errors.

Idempotency: dedupe by Message-ID; generate one if missing; keep Gmail id for API ops. EMAIL_IDs EMAIL_IDs

Persist cursors/checkpoints outside the archive DB so restarts continue where they left off. adr-003-gmail-metadata-…

1) Batching strategy
   Outer loop: time slices
   Iterate monthly (or weekly for finer control) with Gmail’s query syntax (after:/before:). If a run dies, you re-enter the same slice and continue; failures don’t poison other months. ARCHIVING

Page size (IDs → users.messages.list)
Use a page size of 200–500 (up to Gmail’s max). Bigger pages reduce round-trips; time slicing keeps the redo scope sane if a page fails mid-run. ARCHIVING

Delete/Archive batch size
Only delete after you’ve durably committed those messages to SQLite. Use batch deletes in chunks (e.g., 100–500 IDs per call) to cap redo. (You’re already planning “archive == delete after sync”.)

Why not “one huge batch”?
Gmail quotas are high, but transient throttling happens. Smaller, atomic units (time slice → pages → small DB txns → delete batches) let you retry without redoing hours of work. ARCHIVING

2) Concurrency without angering Gmail
   Use a bounded worker pool + rate limiter
   Workers fetch message bodies (users.messages.get) concurrently.

A token-bucket rate limiter caps requests/second.

AIMD tuner nudges concurrency up when latency is good and cuts it in half on throttling (429 / userRateLimitExceeded).

Start with 8 workers and ~4 RPS. Raise workers by +1 every few healthy seconds until:

p95 latency > threshold (e.g., 1.2s), or

any 429/403 → cut concurrency by 50% and backoff.

This keeps you fast but polite.

Tiny code sketch (pool + limiter + tuner):

go
Copy
Edit
type Tuner struct {
mu      sync.Mutex
cur     int // current workers
min, max int
}

func (t *Tuner) Inc() { t.mu.Lock(); if t.cur < t.max { t.cur++ }; t.mu.Unlock() }
func (t *Tuner) Halve() { t.mu.Lock(); if t.cur > t.min { t.cur = max(t.min, t.cur/2) }; t.mu.Unlock() }
func (t *Tuner) Cur() int { t.mu.Lock(); defer t.mu.Unlock(); return t.cur }

func isThrottle(err error) bool {
// check googleapi.Error with Code==429 or 403 & reason=userRateLimitExceeded
var g *googleapi.Error
if errors.As(err, &g) && (g.Code == 429 || g.Code == 403) { return true }
return false
}

func fetchWithPool(ctx context.Context, svc *gmail.Service, ids []string,
qps float64, tuner *Tuner, out chan<- *gmail.Message) error {

    limiter := rate.NewLimiter(rate.Limit(qps), int(math.Max(1, qps)))
    jobs := make(chan string, 2000)
    errCh := make(chan error, 1)

    spawn := func(n int) {
        for i := 0; i < n; i++ {
            go func() {
                for id := range jobs {
                    if err := limiter.Wait(ctx); err != nil { return }
                    t0 := time.Now()
                    msg, err := svc.Users.Messages.Get("me", id).Format("raw").Do()
                    if err != nil {
                        if isThrottle(err) { tuner.Halve(); time.Sleep(time.Second); jobs <- id; continue }
                        select { case errCh <- err: default: }
                        continue
                    }
                    _ = time.Since(t0) // record latency if you want p95 windows
                    out <- msg
                }
            }()
        }
    }

    spawn(tuner.Cur())
    go func() {
        ticker := time.NewTicker(3 * time.Second) // additive increase
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                spawn(1); tuner.Inc()
            case <-ctx.Done():
                return
            }
        }
    }()

    for _, id := range ids { jobs <- id }
    close(jobs)
    return nil
}
Gmail-specific notes: listing by date range + paging + get calls is the idiomatic loop, and quotas are generous (keep backoff on RateLimitExceeded). ARCHIVING ARCHIVING

3) SQLite writes: fast, safe, resumable
   Put SQLite in WAL mode; use a single writer goroutine that batches inserts in small transactions (e.g., 100 msgs or 8–16 MB, whichever first). This keeps the redo scope small if the process dies mid-run.

If you’re sharding (by year, size, etc.), route to the right shard and keep one writer per shard to avoid write contention. Cross-shard queries use ATTACH/UNION ALL later. adr-002-sqlite-sharding…

Idempotency: upsert on Message-ID (generate one when missing) so replays are harmless; also store Gmail’s id/threadId. EMAIL_IDs EMAIL_IDs

4) Delete/Archive safely
   Two-phase:

Fetch → parse → commit to SQLite (transaction complete).

Batch delete those committed IDs from Gmail (or remove INBOX label if you truly mean “archive”). Keep batch size 100–500; retry failures without re-fetching.

Because you only delete after a durable commit, mixed success across goroutines can’t cause data loss.

5) Being “chronological” without blocking concurrency
   You worried about concurrency causing out-of-order results. The outer time slice gives you chronological boundaries, and your database rows carry the real Date/internalDate, so insertion order doesn’t matter. For future incremental runs, store a watermark (e.g., last processed internalDate or “last month completed”) and query “after watermark minus an overlap window (say 48h)”, relying on idempotent inserts to dedupe. The monthly slicing pattern is already in your archive doc. ARCHIVING

Persist this watermark (and per-slice progress) in your metadata store so restarts resume cleanly. adr-003-gmail-metadata-…

6) Putting it together (sketch)
   go
   Copy
   Edit
   // 1) For each time slice:
   for _, slice := range monthlySlices(start, end) {
   // Load/restore slice progress (page token, last ID) from metadata store
   // ...

   // 2) List message IDs for this slice in pages
   var ids []string
   call := svc.Users.Messages.List("me").
   Q(slice.Query()).MaxResults(400) // 200-500 is fine

   for {
   lr, err := call.Do()
   if err != nil { backoffRetry(...); continue }
   for _, m := range lr.Messages { ids = append(ids, m.Id) }
   if lr.NextPageToken == "" { break }
   call = svc.Users.Messages.List("me").Q(slice.Query()).
   MaxResults(400).PageToken(lr.NextPageToken)
   }

   // 3) Fetch concurrently
   out := make(chan *gmail.Message, 200)
   tuner := &Tuner{cur: 8, min: 2, max: 32}
   go func(){ _ = fetchWithPool(ctx, svc, ids, 4.0, tuner, out); close(out) }()

   // 4) Write in small DB txns; 5) queue for batch delete after commit
   toDelete := make([]string, 0, 500)
   buf := make([]*gmail.Message, 0, 200)
   for msg := range out {
   buf = append(buf, msg)
   if len(buf) >= 100 {
   commitToSQLite(buf) // WAL, single writer, UPSERT by Message-ID
   for _, m := range buf { toDelete = append(toDelete, m.Id) }
   buf = buf[:0]
   if len(toDelete) >= 300 {
   batchDelete(ctx, svc, toDelete) // retry on 429/403 with backoff
   toDelete = toDelete[:0]
   }
   }
   }
   if len(buf) > 0 { commitToSQLite(buf); for _, m := range buf { toDelete = append(toDelete, m.Id) } }
   if len(toDelete) > 0 { batchDelete(ctx, svc, toDelete) }

   // Update slice cursor/watermark in metadata store (atomic)
   }
   The list→get→decode→store flow matches your existing archival guidance. ARCHIVING ARCHIVING

7) Defaults to ship with (then tune dynamically)
   Time slice: month (switch to week for massive mailboxes). ARCHIVING

List page size: 400 (raise/lower as needed).

Workers: start 8; cap 32 via tuner.

Rate limit: start 4 RPS; tuner increases slowly; halve on throttle.

DB txn size: 100 msgs or 8–16 MB, whichever first.

Delete batch: 300 IDs.

Retry: exponential backoff with jitter on 429/403; resume from metadata cursor. adr-003-gmail-metadata-…

8) Why this fits your architecture
   Works cleanly with your sharded SQLite plan (one writer per shard, ATTACH for reads). adr-002-sqlite-sharding…

Uses Gmail’s query + paging the way your archive doc recommends. ARCHIVING

Resumability uses the metadata layer you’ve chosen (Apps Script), not the archive DB. adr-003-gmail-metadata-… adr-003-gmail-metadata-…

If you want, I can turn this into a tiny Go package (syncengine) with:

the worker pool + limiter + tuner,

a Storage interface (so you can plug in your shard writer),

and a ProgressStore that uses your Apps Script metadata client.






You said:
I envision archiving to be a step that is separate from syncing but that requires the records to be synced to run. I do not envision running a sync batch then deleting that sync batch in the same process.  At least not unless and until we identify some reason we must do so.

> 1) Batching strategy

AFAICT you do not address managing multiple GoRoutine and how a later time slice can fail while an earlier one can succeed. However, as I am reading it occurs to me how to do it and we do it similar to how TCP/IP can send packets out of order but they get reassembled in order by the receiver.

So the process that invokes Goroutines would add an entry into a slice for each Goroutine where each Goroutine is given a time range of messages it is responsible for. Once a Goroutine finishes it markes its locks the slice, makes its slice entry as "done" and then from the beginning of the slice it iterates the slice until it finds a Goroutine that has not completed and it takes the info from the last completed time slice, submits an update to the account metadata telling it the latest date that has been synced, reslices the slice to start at the first pending Goroutine, then it unlocks the slice.

Let's call this time-slice reassembly (unless you have a better name) so we can refer to this approach.

Please elaborate on token-bucket rate limiter and AIMD tuner. I had to google them so it would be helpful to provide an overview of the basics.

What do you mean by RPS in this context?

> The outer time slice gives you chronological boundaries, and your database rows carry the real Date/internalDate, so insertion order doesn’t matter.

It is not insertions that matter here, is it the account metadata that needs to keep track of sync progress, especially when starting with a new database.

> query “after watermark minus an overlap window (say 48h)”,

I do not like this because it presumes the delta will never be more than 48h.  And it we make it 96h, then it presumes the delta will never be more than 96h. IOW, that approach will always fail at edge cases. OTOH I think the approach I named "time-slice reassembly" will always work determinatively, unless you see holes in the approach that I missed.

Before we do more, let's make sure we have hashed out all of my concerns and any follow up concerns you have after which I want to have you write another ADR on this topic.


ChatGPT said:

Thought for 41s
Great—splitting sync and archive as separate jobs is the right call. We’ll design sync to be order-agnostic internally but order-aware when committing progress, so your metadata (watermark) only moves forward when it’s safe.

Names & concepts (quick)
RPS = requests per second (how many API calls you fire at Gmail per second).

Token-bucket rate limiter: think “bucket” refilled at rate r tokens/sec, max capacity b. Every API call must take a token. If empty, you wait. That lets you do short bursts up to b but caps long-run rate at r.

AIMD tuner (Additive-Increase / Multiplicative-Decrease): slowly increase concurrency/rate while things look healthy; when you see throttling/latency spikes, cut your concurrency/rate in half. Rinse, repeat. It converges near the service’s comfortable limit without you picking a magic number.

“Time-slice reassembly”
Your idea is solid. I’d call it ordered frontier commit (same idea as TCP’s cumulative ACK): let many slices run out of order, but only advance the frontier (your sync watermark) when all slices before it are done.

How it fits
Split the target time range into slices (e.g., days or weeks).

Enqueue all slices to a worker pool. Workers may finish them out of order.

After any slice finishes, mark it done and advance the frontier over any contiguous prefix that is now complete. Persist the watermark at each advancement (e.g., to the slice’s End timestamp).

The archiver only operates up to the watermark, so it never touches mail that wasn’t fully synced.

This avoids fragile “overlap windows.” No matter how late a slice finishes, your frontier moves only when it’s safe, deterministically.

Edge handling
If one slice stalls forever, the frontier stops—but that’s correct. You can detect it (timeout / retry budget), split it smaller, or quarantine it.

If new mail lands “behind” the frontier (rare, but clock skew / imports happen), it’ll be in a later run’s slices; and because inserts are idempotent, reprocessing adjacent slices is harmless.

Practical slicing
Granularity: start with 1 day for large mailboxes, 1 week for moderate, 1 month for small. Make it a config knob.

Dynamic split (optional): if a worker sees “too many pages” for a slice (e.g., >20 list pages at maxResults=500), split the remaining half of the time range into a new slice and enqueue it. Don’t bother merging—keeping code simple is worth it.

Concurrency & throttling
Use a bounded worker pool (e.g., 6–12 workers to start) to fetch message bodies in parallel.

Gate every API call (list, get, delete) behind a token bucket.

Drive both workers and RPS with an AIMD tuner:

Add +1 worker (or +0.5 RPS) every few healthy seconds when p95 latency < threshold and no throttling.

On 429/403 “rate limit” (or p95 blows past threshold), halve workers and/or RPS and back off with jitter.

What to store in metadata
Keep it minimal:

Watermark: “highest fully-synced timestamp (exclusive).”

Run id (for observability), plus optional per-slice status if you want fancy dashboards.
If you crash, you can just discard in-flight slices and reschedule the window from the watermark forward. Inserts are idempotent, so retries don’t corrupt state.

Reference implementation sketches (Go)
These are compact on purpose. You can adapt to your “single return + goto end” style later.

1) Slice coordinator (“ordered frontier commit”)
   go
   Copy
   Edit
   type Slice struct {
   Index       int
   Start, End  time.Time // [Start, End)
   }

type Coordinator struct {
mu       sync.Mutex
slices   []Slice
done     []bool
frontier int
// Persist watermark to external metadata (account-level)
persist func(ctx context.Context, ts time.Time) error
}

func NewCoordinator(slices []Slice, persist func(context.Context, time.Time) error) *Coordinator {
return &Coordinator{
slices: slices,
done:   make([]bool, len(slices)),
persist: persist,
}
}

// call when a slice finishes (err == nil means success)
func (c *Coordinator) Complete(ctx context.Context, idx int, err error) error {
c.mu.Lock()
if err == nil {
c.done[idx] = true
// advance frontier over any contiguous done prefix
for c.frontier < len(c.slices) && c.done[c.frontier] {
end := c.slices[c.frontier].End
c.mu.Unlock()                   // persist outside lock
perr := c.persist(ctx, end)     // watermark := end
c.mu.Lock()
if perr != nil {
c.mu.Unlock()
return perr
}
c.frontier++
}
}
c.mu.Unlock()
return err // up to caller to decide retries/splits
}
2) Worker pool with token-bucket + AIMD
   go
   Copy
   Edit
   type Tuner struct {
   mu        sync.Mutex
   workers   int
   minW, maxW int
   qps       float64
   minQPS, maxQPS float64
   }

func (t *Tuner) Snapshot() (int, float64) {
t.mu.Lock(); defer t.mu.Unlock()
return t.workers, t.qps
}
func (t *Tuner) Inc() {
t.mu.Lock()
if t.workers < t.maxW { t.workers++ }
if t.qps < t.maxQPS { t.qps += 0.5 }
t.mu.Unlock()
}
func (t *Tuner) Halve() {
t.mu.Lock()
if t.workers > t.minW { t.workers = max(t.minW, t.workers/2) }
t.qps = math.Max(t.minQPS, t.qps/2)
t.mu.Unlock()
}

type GmailClient interface {
ListIDs(ctx context.Context, start, end time.Time, pageTok string, pageSize int) (ids []string, next string, err error)
GetMessage(ctx context.Context, id string) (*Message, error)
}

func RunSync(ctx context.Context, cli GmailClient, slices []Slice, persist func(context.Context, time.Time) error) error {
coord := NewCoordinator(slices, persist)
// Initial bounds (tune later)
tuner := &Tuner{workers: 8, minW: 2, maxW: 32, qps: 4, minQPS: 1, maxQPS: 16}

    jobs := make(chan Slice)
    errs := make(chan error, len(slices))
    var wg sync.WaitGroup

    // adaptive spawner
    go func() {
        // simple loop that periodically re-reads desired worker count
        current := 0
        for {
            w, _ := tuner.Snapshot()
            for current < w {
                wg.Add(1)
                go worker(ctx, cli, coord, tuner, jobs, errs, &wg)
                current++
            }
            time.Sleep(2 * time.Second)
            if ctx.Err() != nil { return }
        }
    }()

    // feed work
    go func() {
        for _, s := range slices { jobs <- s }
        close(jobs)
    }()

    // simple health loop: bump up if healthy
    go func() {
        t := time.NewTicker(5 * time.Second)
        defer t.Stop()
        for {
            select {
            case <-t.C:
                // in real code, check p95 latency / error counters first
                tuner.Inc()
            case <-ctx.Done():
                return
            }
        }
    }()

    wg.Wait()
    close(errs)
    // return first error if any
    for err := range errs {
        if err != nil { return err }
    }
    return nil
}

func worker(ctx context.Context, cli GmailClient, coord *Coordinator, tuner *Tuner, jobs <-chan Slice, errs chan<- error, wg *sync.WaitGroup) {
defer wg.Done()

    // token bucket limiter per goroutine; share a global limiter if you prefer
    // (golang.org/x/time/rate is ideal; sketching a simple Waiter here)
    limiter := newLimiter() // implements Wait(ctx) honoring tuner.qps snapshot

    for s := range jobs {
        err := processSlice(ctx, cli, limiter, s)
        if isThrottle(err) || isTooSlow(err) {
            tuner.Halve()
        }
        // mark completion & possibly advance frontier/watermark
        cerr := coord.Complete(ctx, s.Index, err)
        if err == nil && cerr == nil {
            errs <- nil
        } else {
            // caller can implement retry/split policies based on err
            errs <- firstNonNil(err, cerr)
        }
    }
}
3) Slice processing (list → get → store), with mid-slice split option
   go
   Copy
   Edit
   func processSlice(ctx context.Context, cli GmailClient, limiter *Limiter, s Slice) error {
   // list IDs in pages
   const pageSize = 500
   const splitAfterPages = 20

   pages := 0
   pageTok := ""

   for {
   if err := limiter.Wait(ctx); err != nil { return err }

        ids, next, err := cli.ListIDs(ctx, s.Start, s.End, pageTok, pageSize)
        if err != nil { return err }

        // fetch bodies with local concurrency (bounded)
        if err := fetchAndStore(ctx, cli, limiter, ids); err != nil { return err }

        pages++
        if pages > splitAfterPages && next != "" {
            // OPTIONAL: split remaining time in half and enqueue new slice…
            // (You’d need a thread-safe “enqueue” callback; omitted for brevity.)
            // s.End = midpoint(s.Start, s.End) // shrink current slice
        }

        if next == "" { break }
        pageTok = next
   }
   return nil
   }

func fetchAndStore(ctx context.Context, cli GmailClient, limiter *Limiter, ids []string) error {
// single-writer batching to SQLite is assumed behind this call
// sketch: small batch txns (e.g., 100 msgs)
batch := make([]*Message, 0, 100)
for _, id := range ids {
if err := limiter.Wait(ctx); err != nil { return err }
msg, err := cli.GetMessage(ctx, id)
if err != nil { return err }
batch = append(batch, msg)
if len(batch) >= 100 {
if err := commitToSQLite(batch); err != nil { return err }
batch = batch[:0]
}
}
if len(batch) > 0 {
if err := commitToSQLite(batch); err != nil { return err }
}
return nil
}
Notes:

commitToSQLite should be idempotent (UPSERT on Gmail id or RFC Message-ID) so replays are safe.

Use WAL mode and one writer per shard.

4) Minimal token-bucket (or use x/time/rate)
   go
   Copy
   Edit
   type Limiter struct {
   mu       sync.Mutex
   tokens   float64
   capacity float64
   rate     float64
   last     time.Time
   tuner    *Tuner
   }

func newLimiter() *Limiter {
return &Limiter{tokens: 1, capacity: 4, rate: 4, last: time.Now()}
}

func (l *Limiter) Wait(ctx context.Context) error {
for {
l.mu.Lock()
// pull the latest qps snapshot
_, qps := l.tuner.Snapshot()
if qps > 0 {
l.rate = qps
l.capacity = math.Max(2, qps*1.5) // small burst allowance
}
now := time.Now()
elapsed := now.Sub(l.last).Seconds()
l.last = now
l.tokens = math.Min(l.capacity, l.tokens+elapsed*l.rate)
if l.tokens >= 1 {
l.tokens -= 1
l.mu.Unlock()
return nil
}
need := (1 - l.tokens) / l.rate
l.mu.Unlock()

        // sleep or abort if context cancelled
        t := time.NewTimer(time.Duration(need*float64(time.Second)))
        select {
        case <-t.C:
        case <-ctx.Done():
            t.Stop()
            return ctx.Err()
        }
    }
}
Why this addresses your concerns
No overlap guesses: the frontier only advances over a contiguous set of completed slices. No 48h/96h compromises.

Out-of-order concurrency: safe. Reassembly is built-in; the DB doesn’t care about order; progress only moves when it’s safe.

Partial failures: a failed slice doesn’t block other slices’ processing—but it does block the frontier beyond it. That’s exactly what you want for correctness.

Gmail friendliness: token bucket + AIMD keeps you under invisible limits while still finding a fast operating point.

If this approach looks good, I’ll write the ADR next with:

the ordered frontier commit design,

concurrency & throttling policy (token bucket + AIMD),

slice sizing (static + optional dynamic split),

failure/retry semantics,

metadata schema for the watermark and observability,

and the rationale vs. alternatives (overlap windows, single-threaded, history API, etc.).






You said:
I know we would be foolish to not be able to handle when a slice stalls forever, but I have run a large move of messages from two accounts to a third and none of the requests ever stalled. So I think this will be rare, but we do need to somehow accommodate for it.

I did not understand what you meant when you explained "Dynamic split (optional)"; specifically what you meant 1.) by enqueuing and 2.) by merging.


ChatGPT said:

Thought for 20s
Gotcha. A slice “stall” should be rare, but we still want a clean, deterministic way to detect it, contain it, and keep the rest of the system busy.

How we handle a slice that stalls
Heartbeat + timeout: Each worker periodically reports progress (e.g., pages processed, last page token, last Gmail internalDate). If no heartbeat for N minutes (say 5–10), we mark the slice stalled.

Retry budget: Retry the slice with backoff up to K attempts (e.g., 5). On each retry, we resume from its last page token/progress marker.

Dynamic split on trouble: If a slice consistently exceeds a “too big/slow” threshold (pages > 20, messages > 10k, p95 latency above T), we split it into smaller time ranges and try again.

Quarantine, but don’t block work: A stalled slice blocks the frontier from advancing past it (your “time-slice reassembly” rule), but does not stop workers from processing later slices. They can finish and wait behind the frontier. You’ll see 100% CPU usage continue, just no watermark advancement past the hole.

Manual override (optional): Provide an admin switch to mark a slice “skipped” (hole allowed) if you must archive past it. You probably won’t use this, but it’s useful in emergencies.

“Dynamic split (optional)” — what I meant
Think of it like quicksort on time ranges.

1) “Enqueue”
   Enqueue means “add a new slice to the shared work queue so another goroutine can pick it up.” Concretely, you keep a dispatcher that owns a channel and a WaitGroup, and exposes Enqueue() for workers:

go
Copy
Edit
type Slice struct{ Index int; Start, End time.Time }

type Dispatcher struct {
jobs chan Slice
wg   sync.WaitGroup
}

func NewDispatcher(buf int) *Dispatcher {
return &Dispatcher{jobs: make(chan Slice, buf)}
}

func (d *Dispatcher) Enqueue(s Slice) {
d.wg.Add(1)
d.jobs <- s
}

func (d *Dispatcher) Done() { d.wg.Done() }

func (d *Dispatcher) Jobs() <-chan Slice { return d.jobs }

func (d *Dispatcher) CloseWhenIdle() {
go func() { d.wg.Wait(); close(d.jobs) }()
}
Seed the queue by calling Enqueue() for your initial daily/weekly slices.

Workers call d.Done() when they finish a slice.

If a worker decides to split a slice, it calls d.Enqueue() with the new sub-slice(s) before d.Done() on the current one. The WaitGroup accounts for dynamic growth/shrink without races.

A goroutine calls CloseWhenIdle() once seeding is complete; it will auto-close jobs when all work (including dynamically added slices) is finished.

2) “Merging”
   You don’t merge anything back together. After splitting:

You replace one slice by two logical slices in the coordinator and treat them independently.

The database already holds message rows; there’s nothing to merge there.

Your frontier (cumulative commit) advances once all slices before it—including the newly created sub-slices—are marked done. No explicit “merge” step.

Minimal split logic (inside a worker)
go
Copy
Edit
func maybeSplit(s Slice, pagesSeen int) (left Slice, right *Slice, doSplit bool) {
if pagesSeen <= 20 { return s, nil, false }
mid := s.Start.Add(s.End.Sub(s.Start) / 2)
left = Slice{Start: s.Start, End: mid}
r := Slice{Start: mid, End: s.End}
return left, &r, true
}
The worker shrinks the current slice to [Start, mid) and enqueues the right half [mid, End) as a new slice. No merging required later.

“Time-slice reassembly” (a.k.a. ordered frontier commit)
Your name works; if you want a nerdy alternative, call it ordered frontier commit (same idea as TCP’s cumulative ACK).

Coordinator change to support dynamic split:

go
Copy
Edit
type Coordinator struct {
mu       sync.Mutex
slices   []Slice
done     []bool
frontier int
persist  func(context.Context, time.Time) error // writes watermark
}

func (c *Coordinator) AddSliceAfter(i int, s Slice) int {
c.mu.Lock()
defer c.mu.Unlock()
idx := len(c.slices)
// append at the end; order only matters logically for frontier,
// which uses Start/End timestamps, not just indices:
c.slices = append(c.slices, s)
c.done = append(c.done, false)
return idx
}

func (c *Coordinator) Complete(ctx context.Context, idx int, err error) error {
c.mu.Lock()
if err == nil { c.done[idx] = true }
// Advance frontier by **timestamp order**, not array index:
// find the max End time such that *every* slice ending <= that time is done.
frontierTime := c.computeContiguousDoneEndLocked()
c.mu.Unlock()

    if !frontierTime.IsZero() {
        if perr := c.persist(ctx, frontierTime); perr != nil { return perr }
    }
    return err
}
Implementation detail: Frontier calculation should be based on time ordering, not just indices, because dynamic splits add new entries. Easiest path: keep slices in a min-heap or sorted set keyed by End, and a bitmap of “done”; advance to the largest End for which all slices with End <= that are done.

Putting stalls + split together (flow)
Worker starts slice S, emits heartbeats.

If S trips thresholds (too many pages, too slow), it splits S at midpoint:

Shrink S to left half.

Enqueue right half R.

If S stalls (no heartbeat in N minutes), cancel and retry S (resume from last page token). After K failed attempts, mark S quarantined and alert.

The frontier only moves to maxEnd where all slices with End <= maxEnd are done. Quarantined S prevents crossing its End, but won’t stop other slices from finishing.

Why this is robust
No “overlap window” guesses—progress moves only when all prior time is covered.

A single pathological slice doesn’t idle the cluster; it only pauses the watermark beyond it.

Dynamic splitting naturally homes in on slice sizes that your account + network + Gmail are happy with.

If this clears it up, I’ll fold these into the ADR (with diagrams, thresholds, and the frontier-by-time algorithm spelled out).






You said:
> Manual override (optional): Provide an admin switch to mark a slice “skipped” (hole allowed) if you must archive past it.

I think this should be automated. For anything that stalls repeatedly I think we should keep subdividing until we get known individual failures (messages?) and then mark them in account metadata like hard disks mark bad sectors. We also need a way to keep track of when we start getting a lot of bad messages and/or when we exceed 50% of available space for metadata and we need to start throwing highly visible warnings (maybe we send emails when this happens?). We can wait to implement an alternative way to track the "bad sectors" unless and until it becomes a problem, but we need to be prepared to recognize it is becoming a problem so we can take action in advance.

I did not realize we had discussed a "shared work queue" yet. How is that implemented?

Also, one thing is that we will get a slice of messages by having a start and end time but once we do that we will have a list of message IDs. If we split then I think we should enqueue those message IDs and not the time range since we already have those message IDs. Translating back to time range and then back to message IDs seems like both unecessary extra effort and less precise ability to track the pending messages.

> ordered frontier commit

That name is fine, we can use it instead of mine.


ChatGPT said:

Thought for 21s
Love it—you’re thinking exactly like a storage engine. Let’s bake those ideas in.

Bad-message isolation (“bad sectors”), automated
Policy: when a batch keeps failing, recursively split it until you isolate single message IDs that fail deterministically. Mark those message IDs as bad in account metadata with:

messageId, reason (e.g., decode error, 410 gone, malformed RFC822), firstSeen, lastTried, retryCount.

Don’t block the slice because of them; count a slice “done” when all its IDs are either stored or flagged bad.

Why: this guarantees forward progress without manual overrides and avoids unbounded reprocessing.

Alerting / guardrails:

Trigger warning if bad_count / processed_count exceeds a threshold (e.g., 0.2%) over a moving window (10k msgs).

Trigger critical if it exceeds a higher threshold (e.g., 1%) or bursts (e.g., >200 bad in 1h).

Also track metadata store pressure: if estimated bytes used exceeds 50% of budget (configurable), raise an alert. (We can estimate each record’s JSON length + overhead we add; if the backing store exposes usage, use that.)

Notification: log + metric + optional self-email via Gmail API (rate-limited to avoid spam).

Shared work queue — how it’s implemented
Think dispatcher + channel + WaitGroup, with everything in one process (we can swap in a persistent queue later if needed).

Enumerator (per time slice) lists message IDs and chunks them into fixed-size batches (e.g., 200–500 IDs/batch). It registers the slice with the coordinator: slice.totalBatches = N.

Dispatcher owns a channel of Job{sliceID, ids[]}. Workers pull from it.

Workers process a job (fetch/store); on error they split the job into halves and re-enqueue; when a job completes, worker informs the coordinator, which tracks per-slice outstanding batches. When outstanding==0 (success + bad isolated), the slice is done.

Frontier (ordered frontier commit): advance the account watermark to the greatest slice.End for which every earlier slice is done. Archiver works only up to that watermark.

You’re right to enqueue message ID batches, not (re)time ranges. Time ranges are for slice definition & frontier semantics; once you’ve listed IDs, the queue deals only in IDs.

Key structs (concise)
go
Copy
Edit
type Slice struct {
ID          int
Start, End  time.Time // [Start, End)
TotalBatches int64
DoneBatches  int64
Outstanding  int64 // = TotalBatches - DoneBatches - BadBatches
BadBatches   int64
}

type Job struct {
SliceID int
IDs     []string // a batch of Gmail message IDs
Tries   int
}

type BadMessage struct {
ID         string
Reason     string
FirstSeen  time.Time
LastTried  time.Time
RetryCount int
}
Dispatcher + queue (in-process)
go
Copy
Edit
type Dispatcher struct {
jobs chan Job
wg   sync.WaitGroup
}

func NewDispatcher(buf int) *Dispatcher {
return &Dispatcher{jobs: make(chan Job, buf)}
}

func (d *Dispatcher) Enqueue(j Job) {
d.wg.Add(1)
d.jobs <- j
}

func (d *Dispatcher) JobChan() <-chan Job { return d.jobs }

func (d *Dispatcher) Done() { d.wg.Done() }

// Close channel automatically when all enqueued work (including splits) is finished.
func (d *Dispatcher) CloseWhenIdle() { go func() { d.wg.Wait(); close(d.jobs) }() }
Coordinator (ordered frontier commit + per-slice accounting)
go
Copy
Edit
type Coordinator struct {
mu        sync.Mutex
slices    map[int]*Slice
// sorted end times for frontier; simplest: keep slices keyed by End
ends      []time.Time
persistWatermark func(context.Context, time.Time) error
}

func (c *Coordinator) AddSlice(s *Slice) {
c.mu.Lock()
c.slices[s.ID] = s
c.ends = append(c.ends, s.End)
sort.Slice(c.ends, func(i, j int) bool { return c.ends[i].Before(c.ends[j]) })
c.mu.Unlock()
}

func (c *Coordinator) BatchDone(ctx context.Context, sliceID int, good, bad int) error {
var frontier time.Time
c.mu.Lock()
s := c.slices[sliceID]
if s != nil {
s.DoneBatches += int64(good)
s.BadBatches  += int64(bad)
s.Outstanding  = s.TotalBatches - s.DoneBatches - s.BadBatches
}
// advance frontier to the max End such that all slices with End<=that are finished
// finished = Outstanding==0 for that slice
doneUpTo := time.Time{}
for _, end := range c.ends {
all := true
for _, sl := range c.slices {
if !sl.End.After(end) && sl.Outstanding > 0 { all = false; break }
}
if all { doneUpTo = end } else { break }
}
frontier = doneUpTo
c.mu.Unlock()

    if !frontier.IsZero() {
        return c.persistWatermark(ctx, frontier)
    }
    return nil
}
Worker logic: process → split → isolate bad
This is the crux: never block, just shrink the problem until you can mark a single ID as bad.

go
Copy
Edit
func handleJob(ctx context.Context, j Job, g GmailClient, lim *Limiter) (goodBatches, badBatches int, err error) {
// fetch/store the whole batch; if it fails, split and re-enqueue halves
var ok bool
ok, err = processBatch(ctx, j.IDs, g, lim)
if ok {
goodBatches = 1
goto end
}
if len(j.IDs) == 1 {
// isolate bad message
markBad(j.IDs[0], err)
badBatches = 1
err = nil
goto end
}

    // split
    mid := len(j.IDs) / 2
    left := Job{SliceID: j.SliceID, IDs: append([]string(nil), j.IDs[:mid]...), Tries: j.Tries + 1}
    right:= Job{SliceID: j.SliceID, IDs: append([]string(nil), j.IDs[mid:]...), Tries: j.Tries + 1}
    enqueue(left)
    enqueue(right)
    // current job’s accounting will be done when children finish
    err = nil
end:
return
}
processBatch should be robust to per-message issues (retry transient errors; treat durable failures as batch failure so the splitter kicks in). The splitter then finds the minimum failing unit.

This binary-search isolation is fast: O(log N) extra steps per bad message, with N = batch size.

Stall detection (heartbeats) & retries
Each worker updates a progress map: {jobID: lastHeartbeatTS, lastIDIndex}.

A supervisor scans every N seconds; if no heartbeat for T (e.g., 5–10 min), it cancels & re-enqueues the job (preserving its IDs).

Maintain a retryCount. If retryCount for a batch exceeds K (e.g., 5), force split even if errors were ambiguous.

Metadata pressure & alerting (simple)
Keep a rolling estimate of bytes stored in metadata: sum(len(json(record))+overhead).

If usage >= 0.5 * budget, set state=WARNING.

If usage >= 0.8 * budget or bad-rate spikes, set state=CRITICAL and send a self-email (with throttling, e.g., at most once per hour per account & condition).

You’ll likely never hit this if you only store a handful of bad IDs, but it’s cheap insurance.

Putting IDs (not time) on the queue while preserving ordered frontier
Slice lifecycle:

Define time slice.

List all IDs for that slice (paged).

Chunk into ID batches; set slice.TotalBatches = count(batches).

Enqueue all batches.

As workers finish or isolate bad IDs, call Coordinator.BatchDone(...).

When Outstanding==0, the slice is finished; frontier may advance.

This keeps work precise (IDs) while the frontier still uses time.

Minimal worker loop (sketch)
go
Copy
Edit
func worker(ctx context.Context, d *Dispatcher, c *Coordinator, g GmailClient, lim *Limiter) {
for j := range d.JobChan() {
good, bad, err := handleJob(ctx, j, g, lim)
_ = err // already re-enqueued on split; only terminal per-batch errors bubble
_ = c.BatchDone(ctx, j.SliceID, good, bad)
d.Done()
}
}
Next steps I can bake into the ADR
Design: ID-batch queue + ordered frontier commit + binary-search isolation for bad messages.

Failure semantics: retry budgets, stall detection, quarantine rules.

Rate control: token bucket + AIMD (we covered; I’ll include concrete thresholds).

Observability: counters for processed/bad/retried, p95 latencies, metadata usage estimate, alert channels.

Extensibility: switch to a persistent queue later (SQLite table with jobs(id, slice_id, payload, tries, state)).

Want me to draft the ADR with this exact shape (plus a tiny diagram of the pipeline)?






You said:
I think we are ready to have you draft a complete, and detailed ADR. I want you to capture everything relevant that we both discussed and decided as well as rationale / pros & cons for each, including things we decided against.  Please include code but make it clear that code is for example/illustration purposes are is not explicitly part of the decisions made.  We will finalize code when we write the code.

And if you want to get really creative feel free to include Mermaid diagrams if there is any value to including diagrams.


ChatGPT said:

Thought for 19s
Adr-004: Sync Concurrency, Batching, And Progress Semantics For Gmail → Sqlite (with Separate Archiver)

ADR-004: Sync Concurrency, Batching, and Progress Semantics for Gmail → SQLite (with Separate Archiver)
Status: Accepted (Draft for implementation)

Date: 2025-08-07

Owners: Mike S. (product/architecture), ChatGPT (drafting)

Related ADRs:

ADR-002: SQLite sharding strategy (referenced)

ADR-003: Gmail metadata storage (referenced)

Note on code: All code herein is illustrative (pseudocode/Go-like) to make the design concrete. It is not normative. We will finalize implementation separately and align with house style (e.g., single return with goto end).

1. Context & Problem
   We are building a Go-based synchronizer that reads Gmail messages and stores them in a SQLite archive for long-term search/analysis. Archiving (deleting or de-labelling messages from Gmail after they are safely captured) is a separate job that must only operate on fully synced ranges.

We need a design that:

Maximizes throughput without getting throttled or flagged by Gmail.

Keeps failure “blast radius” small and makes restarts deterministic.

Allows concurrent processing where slices may finish out of order, yet overall progress advances in proper chronological order.

Automatically isolates pathological content/items ("bad messages") without operator intervention.

Scales with sharded SQLite archives and remains idempotent.

2. Decision (Overview)
   We will:

Slice by time (days/weeks/months; configurable) and discover all message IDs per slice.

Split each slice’s IDs into fixed-size batches and push them onto a shared, in-process work queue.

Pull jobs off the queue with a bounded worker pool. All Gmail API calls (list/get/delete) pass through a token-bucket rate limiter whose target RPS is tuned via AIMD (Additive-Increase/Multiplicative-Decrease).

Write parsed messages to SQLite using WAL mode, small transactions, and one writer per shard. Inserts are idempotent (UPSERT by Gmail id and/or RFC Message-ID).

Maintain an Ordered Frontier Commit (OFC) for progress: only when all slices up to time T are fully processed (including isolating any bad messages) do we advance the account watermark to T. The archiver processes only up to the watermark.

On repeated batch failures, we recursively split failing batches (binary split) until the failure scope is a single message ID; mark such IDs as bad in metadata with reason/counters and continue. We alert if bad-rate or metadata storage usage crosses thresholds.

This avoids fragile overlap heuristics (e.g., “48h window”) and provides deterministic, resumable progress akin to TCP’s cumulative ACK.

3. Glossary
   Slice: A time range [Start, End) used for scoping discovery and progress semantics.

Job: A batch of message IDs (e.g., 200–500) derived from a slice.

Dispatcher / Work Queue: In-process producer/consumer channel for Jobs.

Ordered Frontier Commit (OFC): Progress (watermark) advances only to the greatest End time such that all slices ending ≤ that time are fully finished.

Bad message: A single Gmail message that deterministically fails processing; recorded in metadata and excluded from blocking progress.

RPS: Requests per second (API call rate).

4. Constraints & Non-Goals
   Do not archive (delete/remove labels) inside the sync job. Archiver is a separate phase.

Respect Gmail’s variable, undocumented practical rate tolerances by keeping to adaptive ceilings.

SQLite writes must be robust (WAL), idempotent, and work with shards.

We will not adopt an external/persistent queue in v1; an in-process queue suffices. We’ll keep the design portable so we can migrate to persistent jobs later if needed.

5. Detailed Design
   5.1 Slicing & Progress (Ordered Frontier Commit)
   Why slices? They bound failure impact, enable deterministic restarts, and make human reasoning/observability easier.

Granularity:

Default daily for very large mailboxes

Default weekly for moderate

Default monthly for small (Adjustable; can dynamically split further if needed.)

Flow:

For each slice, list all message IDs via Gmail query (after:/before:) with paging (e.g., maxResults=500).

Chunk IDs into Jobs (e.g., 200–500 IDs per job).

The coordinator tracks per-slice totals/outstanding jobs. When all jobs in a slice are finished (including isolated bad IDs), the slice is done.

The watermark advances to the largest End for which **every slice with **`` is done.

Mermaid — OFC high-level

flowchart LR
A[Enumerate time slices] --> B[For each slice: list all IDs]
B --> C[Chunk IDs into Jobs]
C --> D[Shared Work Queue]
D --> E[Workers fetch & store]
E --> F[Coordinator updates per-slice accounting]
F --> G{All slices up to T done?}
G -- yes --> H[Advance Watermark to T]
G -- no --> D
H --> I[Archiver operates ≤ Watermark]
Determinism: The OFC model makes progress monotonic, avoids overlap heuristics, and cleanly separates concurrency from ordering.

5.2 Work Queue, Concurrency, Throttling
Work queue: In-process Dispatcher with a buffered channel of Jobs and a WaitGroup. The dispatcher closes the queue when idle (all enqueued jobs—including dynamically added split-jobs—are completed).

Workers: A bounded pool (e.g., start 8, cap 32) that pulls Jobs, fetches each message’s full content, parses, and hands off to the shard writer.

Rate control:

Token-bucket limiter: bucket refills at r tokens/sec (target RPS), capacity b tokens (burst). Each API call consumes a token; if empty, the caller waits. This caps sustained load while allowing short bursts.

AIMD tuner: periodically increase workers/RPS additively (e.g., +1 worker, +0.5 RPS) while health is good; on 429/403 or p95 latency exceedance, halve workers/RPS and back off with jitter. This converges near the service’s comfortable capacity without a priori thresholds.

Initial defaults (tunable at runtime):

List maxResults: 500

Job size (IDs/batch): 300 (range 200–500)

Workers: start 8, max 32

RPS: start 4.0, max 16.0; burst capacity ≈ 1.5×RPS

SQLite txn: 100 messages or ~8–16 MB per txn, whichever first

Mermaid — concurrency & rate

sequenceDiagram
participant W as Worker
participant L as Token Bucket
participant G as Gmail API
participant T as AIMD Tuner

loop For each API call
W->>L: Wait() for token
L-->>W: Token granted (or wait)
W->>G: GET message/list
G-->>W: Response (lat, status)
W->>T: Record latency/status
alt Throttle/slow
T-->>W: Halve workers/RPS
else Healthy window
T-->>W: Increment workers/RPS slightly
end
end
5.3 Failure Handling & Bad-Message Isolation
Goal: Never let a small number of problematic items block progress.

Strategy:

If a job (batch of IDs) fails after retries, split it into two smaller jobs and re-enqueue both (binary search).

Continue until a job contains one ID; if that also fails deterministically, mark it bad and count the job as finished (bad). The slice can complete without that message.

Store bad-message metadata: {messageId, reason, firstSeen, lastTried, retryCount}.

Stall detection: Workers emit heartbeats (pages processed, last token). A supervisor cancels & re-enqueues jobs that fail to heartbeat within T (5–10 min). Jobs keep a retry budget (e.g., 5); exceeding budget forces a split.

Alerting & guardrails:

Warn when bad_count / processed_count ≥ 0.2% over a 10k message window.

Critical when ≥ 1% or spikes (e.g., >200 bad in 1h).

Track metadata usage; if estimated usage ≥ 50% of budget → warn; ≥ 80% → critical. Optionally send a self-email (rate-limited) for visibility.

Why this works: Binary-splitting gives ~O(log N) isolation steps per bad item (N = batch size). Progress remains monotonic.

5.4 SQLite Write Strategy
WAL mode, synchronous=NORMAL (or FULL during finalization phases).

One writer per shard to avoid contention; write via a single goroutine per shard.

Small transactions (100 msgs or 8–16 MB) to cap redo on crash and keep checkpoints fast.

Idempotent upserts keyed by Gmail id and RFC Message-ID (when absent, generate a stable surrogate). Keep both Gmail id & threadId for API ops and relational joins.

Works with ADR-002 sharding: router selects shard by policy (time/size), writer commits locally, reads can ATTACH/UNION ALL when needed.

5.5 Sync vs Archive (Separation of Concerns)
Sync guarantees durable storage of message content & metadata in SQLite, plus progress in metadata (watermark).

Archive (a separate job) deletes or removes labels only up to the watermark. It batches deletes (e.g., 300 IDs/call) and retries independently.

This separation reduces blast radius and simplifies reasoning (no “fetch-then-delete” coupling).

5.6 Metadata Model (Account-Level)
Minimal but sufficient, stored outside the archive DB (see ADR-003):

watermark: highest fully-synced timestamp (exclusive).

slices[] (optional/ephemeral): status per slice (start, end, totalJobs, doneJobs, badJobs, outstanding), last heartbeat.

badMessages{ id → {reason, firstSeen, lastTried, retryCount} }.

counters: processed, retried, throttled, bad, etc. (windowed).

limits: metadata budget (bytes), alert thresholds; last-alert timestamps for rate-limiting notifications.

We will estimate metadata storage by serialized JSON lengths unless the store gives exact usage.

5.7 Dynamic Split (Optional) at Enumeration Time
While listing IDs for a slice, if we observe that paging is too deep (e.g., >20 pages at maxResults=500), we may split the remaining time of that slice in half and enumerate each half independently. This keeps job creation latency predictable. No merging is required; both halves are just additional slices tracked by OFC.

Once IDs are enumerated, we do not convert back to time ranges—jobs are ID batches only.

5.8 Ordering & Determinism
DB insert order is irrelevant; we store true dates (internalDate, RFC dates) for queries.

Progress order is handled exclusively by OFC; concurrency cannot advance the watermark out of order.

Restarts are deterministic: re-enumerate from the watermark forward (or resume unfinished slices), re-generate jobs, and continue. Upserts make replays harmless.

6. Alternatives Considered
   Overlap window after watermark (e.g., always reprocess last 48–96h):

✅ Simple to implement; tolerates minor out-of-order arrival.

❌ Non-deterministic; can miss edge cases outside the window or reprocess too much. We prefer OFC.

Queue time ranges instead of ID batches:

✅ Conceptually clean.

❌ Requires re-listing & paging at split time; less precise and more waste. We prefer ID batches once IDs are known.

Single-threaded sequential sync:

✅ Simplest and safest for throttling.

❌ Far too slow for large mailboxes; poor utilization.

Fetch then delete in same process:

✅ Reduces intermediate state.

❌ Increases blast radius and complexity; failure modes multiply. We keep archiver separate.

Gmail History API for incremental sync:

✅ Efficient change feed for ongoing deltas.

❌ More complex semantics and edge-cases (gaps, retention). We may revisit after baseline sync is complete; OFC works well for first full ingest.

External/persistent queue (e.g., SQLite-backed jobs, Redis, SQS):

✅ Durability and horizontal scale.

❌ Overhead/complexity not needed for v1; design remains portable if we need it later.

7. Risks & Mitigations
   Throttling / API changes → AIMD + token bucket, aggressive backoff on 429/403; config caps.

Pathological messages → Binary split isolation + bad-message registry.

Metadata store pressure → Usage estimation + warning/critical alerts; potential migration to dedicated store if needed.

Shard writer contention → Single writer per shard, bounded batch size, WAL.

Coordinator bugs → Unit tests against OFC invariants; property tests for reordering/random failures.

8. Observability & Alerts
   Metrics: processed/sec, p50/p95/p99 latency per API, error rates, split count, retries, bad rate, metadata usage.

Logs: job lifecycle, split decisions, stall detection, alert emissions.

Alerts: bad-rate thresholds, metadata thresholds, sustained throttling.

Optional self-email notification (rate-limited: e.g., once/hour per condition).

9. Illustrative Code (non-normative)
   9.1 Types
   // A time slice for progress semantics
   type Slice struct {
   ID            int
   Start, End    time.Time // [Start, End)
   TotalJobs     int64
   DoneJobs      int64
   BadJobs       int64
   Outstanding   int64 // = TotalJobs - DoneJobs - BadJobs
   }

// A job: a batch of message IDs
type Job struct {
SliceID int
IDs     []string // 200–500 typical
Tries   int
}

// A bad message registry entry
type BadMessage struct {
ID         string
Reason     string
FirstSeen  time.Time
LastTried  time.Time
RetryCount int
}
9.2 Dispatcher (work queue)
type Dispatcher struct {
jobs chan Job
wg   sync.WaitGroup
}

func NewDispatcher(buf int) *Dispatcher {
return &Dispatcher{jobs: make(chan Job, buf)}
}
func (d *Dispatcher) Enqueue(j Job) { d.wg.Add(1); d.jobs <- j }
func (d *Dispatcher) Done()         { d.wg.Done() }
func (d *Dispatcher) Jobs() <-chan Job { return d.jobs }
func (d *Dispatcher) CloseWhenIdle()   { go func(){ d.wg.Wait(); close(d.jobs) }() }
9.3 Coordinator (Ordered Frontier Commit)
type Coordinator struct {
mu    sync.Mutex
byID  map[int]*Slice
ends  []time.Time // sorted distinct slice.End values
persistWatermark func(context.Context, time.Time) error
}

func (c *Coordinator) AddSlice(s *Slice) {
c.mu.Lock(); defer c.mu.Unlock()
if c.byID == nil { c.byID = make(map[int]*Slice) }
c.byID[s.ID] = s
c.ends = append(c.ends, s.End)
sort.Slice(c.ends, func(i,j int) bool { return c.ends[i].Before(c.ends[j]) })
}

func (c *Coordinator) JobDone(ctx context.Context, sliceID int, good, bad int) error {
var frontier time.Time
c.mu.Lock()
if s := c.byID[sliceID]; s != nil {
s.DoneJobs += int64(good)
s.BadJobs  += int64(bad)
s.Outstanding = s.TotalJobs - s.DoneJobs - s.BadJobs
}
// advance frontier: largest End where all slices with End<=that are finished
doneUpTo := time.Time{}
for _, end := range c.ends {
all := true
for _, sl := range c.byID {
if !sl.End.After(end) && sl.Outstanding > 0 { all = false; break }
}
if all { doneUpTo = end } else { break }
}
frontier = doneUpTo
c.mu.Unlock()

    if !frontier.IsZero() { return c.persistWatermark(ctx, frontier) }
9.4 Worker, splitting & bad isolation
func processJob(ctx context.Context, j Job, g GmailClient, lim *Limiter, enqueue func(Job), markBad func(string, error)) (good, bad int) {
ok, err := processBatch(ctx, j.IDs, g, lim) // fetch+store; internal retries
if ok {
good = 1
goto end
}
if len(j.IDs) == 1 {
markBad(j.IDs[0], err)
bad = 1
goto end
}
// split and re-enqueue halves; current job completes when children do
mid := len(j.IDs) / 2
left := Job{SliceID: j.SliceID, IDs: append([]string(nil), j.IDs[:mid]...), Tries: j.Tries+1}
right:= Job{SliceID: j.SliceID, IDs: append([]string(nil), j.IDs[mid:]...), Tries: j.Tries+1}
enqueue(left)
enqueue(right)
end:
return
}
9.5 Token Bucket + AIMD (concept)
type Tuner struct {
mu              sync.Mutex
workers         int
minW, maxW      int
qps             float64
minQPS, maxQPS  float64
}
func (t *Tuner) Snapshot() (int, float64) { t.mu.Lock(); defer t.mu.Unlock(); return t.workers, t.qps }
func (t *Tuner) Inc() { t.mu.Lock(); if t.workers<t.maxW {t.workers++}; if t.qps<t.maxQPS {t.qps+=0.5}; t.mu.Unlock() }
func (t *Tuner) Halve(){ t.mu.Lock(); if t.workers>t.minW {t.workers=max(t.minW, t.workers/2)}; t.qps=maxf(t.minQPS, t.qps/2); t.mu.Unlock() }

// Limiter.Wait() consumes a token; it refills at current qps and allows a small burst.
10. Test & Acceptance Criteria
    Correctness

Progress (watermark) never advances past a slice that has unfinished jobs.

Bad-message isolation allows slices to complete despite individual bad IDs.

Restart from any crash resumes deterministically from the watermark forward.

Performance

Achieve stable throughput without sustained throttling (AIMD converges).

SQLite writers maintain steady commit rate with chosen batch sizes.

Resiliency

Stall detector rescues stuck jobs; retries/splits function as specified.

Alerts fire when bad-rate or metadata usage crosses thresholds; self-email throttles appropriately.

11. Operational Notes
    Expose knobs via config/flags: slice granularity, job size, initial/max workers, initial/max RPS, retry budgets, alert thresholds.

Export metrics and logs; maintain dashboards for RPS, latencies, retries, splits, bad rates, watermark.

Archiver runs on an independent cadence; it must read the current watermark before each batch.

12. Open Questions / Future Work
    Switch to a persistent job queue if we need multi-process scale or crash-proof job state.

Add Gmail History API for incremental sync after initial backfill (with OFC still gating progress).

Consider per-label slicing as an additional dimension if certain labels dominate volume.

Explore adaptive job sizes based on observed error/latency distributions.

Evaluate back-pressure from shard writers to throttle job dispatch when queues back up.

13. Parameter Defaults (Initial)
    Slice granularity: weekly (large mailboxes: daily; small: monthly)

List page size: 500

Job (ID batch) size: 300

Workers: start 8, cap 32

RPS: start 4.0, cap 16.0, burst ≈ 1.5×

SQLite txn size: 100 msgs or 8–16 MB

Retry budget per job: 5 attempts; then forced split

Stall timeout: 10 minutes without heartbeat

Alerts: bad-rate warn 0.2%, crit 1% (10k-window); metadata usage warn 50%, crit 80%

End of ADR-004

Answer in chat instead
Done! I drafted the full ADR with diagrams and illustrative Go code. Want any tweaks to naming, thresholds, or the alerting rules before we lock it in?