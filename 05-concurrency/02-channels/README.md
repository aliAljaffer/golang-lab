# 02 — Channels

Typed pipes. The Go idiom for goroutines to talk.

## Things to notice

- **Unbuffered** (`make(chan T)`) is a synchronous *handoff*: the sender blocks until a receiver is there, and vice versa. There is never a value "in" the channel — it's a rendezvous.
- **Buffered** (`make(chan T, N)`) is a tiny queue of size N. Sends only block when the buffer is full; receives only block when it's empty.
- `close(ch)` says "no more values coming." Sending on a closed channel **panics**. Receiving from a closed channel returns the zero value with `ok=false`. The `for v := range ch` loop exits cleanly when the channel is closed and drained.
- The rule: **only the sender closes**, and only when no more sends will happen. If multiple goroutines might send, you need a separate signal (often a `done chan struct{}`).
- `chan<-` (send-only) and `<-chan` (receive-only) in function signatures document who's allowed to do what. The compiler enforces it.

## Comparison

| Concept | Go | Python | TS / Node |
|---|---|---|---|
| Synchronous handoff | `make(chan T)` | `queue.Queue(maxsize=0)` (sort of) | (no native) |
| Bounded queue | `make(chan T, N)` | `queue.Queue(maxsize=N)` | (no native; libraries) |
| Close signal | `close(ch)` | sentinel value | `EventEmitter.emit('end')` |

## Common bugs

- **Send on closed channel**: `panic: send on closed channel`. Audit who closes when there are multiple senders.
- **Forgot to close**: `for range ch` blocks forever — the goroutine leaks.
- **Receive on a never-closed empty channel**: same leak from the receiver side.

## Run

```
go run .
```
