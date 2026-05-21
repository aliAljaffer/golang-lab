# 07 — Streaming a response

`resp.Body` is an `io.Reader`. Anything that consumes an `io.Reader` works directly — no need to ever buffer the whole body.

## When streaming matters

- Logs / NDJSON endpoints with no upper bound on size.
- File downloads (`io.Copy(file, resp.Body)`).
- Server-Sent Events / chunked transfer.
- Any time `Content-Length` is large or absent.

## Patterns

```go
// Line by line
sc := bufio.NewScanner(resp.Body)
for sc.Scan() { handle(sc.Bytes()) }

// JSON stream (one object per line)
dec := json.NewDecoder(resp.Body)
for dec.More() {
    var ev Event
    if err := dec.Decode(&ev); err != nil { return err }
    handle(ev)
}

// Copy to disk
f, _ := os.Create("download.bin")
defer f.Close()
io.Copy(f, resp.Body)
```

## Things to notice

- `bufio.Scanner` has a default 64KB token-size limit. For really long lines, call `scanner.Buffer(...)` with a bigger ceiling.
- `io.Copy` returns the number of bytes copied — handy for progress logging.
- You still need `defer resp.Body.Close()`. Streaming doesn't change ownership.

## Comparison

| Concept | Go | Python | TS |
|---|---|---|---|
| Stream lines | `bufio.Scanner(resp.Body)` | `for line in resp.iter_lines():` | `resp.body.getReader()` |
| Stream to file | `io.Copy(f, resp.Body)` | `shutil.copyfileobj(resp.raw, f)` | `Readable.from(resp.body).pipe(fs)` |

## Run

```
go run .
```
