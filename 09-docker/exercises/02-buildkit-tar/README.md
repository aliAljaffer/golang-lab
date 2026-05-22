# Exercise 02 — buildkit tar

Build an in-memory tar archive suitable for passing to `cli.ImageBuild` as
the build context. No file-system dependency, no Dockerfile on disk.

## What you implement

```go
type File struct {
    Name string  // path inside the tar
    Body []byte
    Mode int64
}

func BuildContext(files []File) ([]byte, error)
```

## Why this is the SDK pattern

`cli.ImageBuild(ctx, contextReader, types.ImageBuildOptions{...})` expects
an `io.Reader` of a tar (optionally gzipped). The Docker CLI builds this tar
by walking your `docker build .` context directory. When you want to:

- generate a Dockerfile on the fly
- ship a build that doesn't touch the file system
- bundle a tiny image-build job into a single Go binary

…you build the tar yourself, in a `bytes.Buffer`, and pass
`bytes.NewReader(buf.Bytes())` to `ImageBuild`. This exercise is the part you
write once and then forget about.

## Contract pinned by tests

- Round-trips through `tar.NewReader`.
- Order preserved.
- Empty input → valid empty tar (no error, no entries).
- Binary bytes survive intact (256-byte payload of every byte value).
- Tar terminator block is written (catches missing `tw.Close()`).

## The `tar.Header` you'll need

```go
&tar.Header{
    Name:     f.Name,
    Mode:     f.Mode,
    Size:     int64(len(f.Body)),
    Typeflag: tar.TypeReg,
    ModTime:  time.Now(),
}
```

`Size` is required and MUST match the body length exactly. Get it wrong and
`tar.NewReader` will eat into the next entry's bytes and break in confusing ways.

## Once you have it

```go
tarBytes, _ := BuildContext([]File{
    {Name: "Dockerfile", Body: []byte("FROM alpine:3\nRUN echo hi\n"), Mode: 0644},
})
resp, _ := cli.ImageBuild(ctx, bytes.NewReader(tarBytes), types.ImageBuildOptions{
    Tags:        []string{"my-image:latest"},
    Remove:      true,
})
defer resp.Body.Close()
io.Copy(os.Stdout, resp.Body)  // drain the build output, JSON lines
```

## Run the failing suite

```bash
go test -tags=exercise ./09-docker/exercises/02-buildkit-tar/
```
