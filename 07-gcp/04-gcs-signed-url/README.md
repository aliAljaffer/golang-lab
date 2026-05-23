# 04 — V4 signed URL

A signed URL embeds the authentication into the URL itself. A third party
without any GCP creds can `curl` the URL and fetch the object for the
expiry window. The server that *generates* the URL is the one that needs
permission.

## V4 vs V2

Always V4. V2 is the older signing scheme and is deprecated for new code
(but the Go SDK still supports it for compat). `Scheme: storage.SigningSchemeV4`
is the right answer; the SDK defaults to V2 for backward compatibility, so
**always set it explicitly**.

## Signing identity

Signing needs a private key. Two paths:

1. **JSON service-account key on disk.** ADC loads it; the storage client
   reads `GoogleAccessID` and `PrivateKey` from it automatically. You can
   leave both empty in `SignedURLOptions` — they're filled in from the key.

2. **No private key (user creds, metadata server, Workload Identity).** The
   SDK falls back to calling `iam.signBlob` on a target service account.
   You must set `GoogleAccessID` to that SA's email, and the calling
   principal needs `roles/iam.serviceAccountTokenCreator` on it. The SDK
   does the IAM call for you.

Path 2 is more common in production. Nobody wants to mount a JSON key on
every container; it's a credential management nightmare.

## What can be signed

| Method   | Use case                                                     |
| -------- | ------------------------------------------------------------ |
| `GET`    | "Download this once" links (most common)                     |
| `PUT`    | "Upload directly from a browser" forms (avoids server proxy) |
| `DELETE` | "Click this link to remove your data" flows                  |

You can also bind headers: `opts.Headers = []string{"x-goog-content-length-range:0,1048576"}`
caps an upload at 1 MB. The client must include the header on the request
or the signature fails.

## Compare to AWS

| Concept           | Go (GCS)                                              | Go (AWS S3)                                              |
| ----------------- | ----------------------------------------------------- | -------------------------------------------------------- |
| Make a signed URL | `bucket.SignedURL(key, &SignedURLOptions{...})`       | `s3.NewPresignClient(c).PresignGetObject(ctx, in, ...)`  |
| Default expiry    | none — you must set `Expires` time                    | `s3.WithPresignExpires(d)`                               |
| Default scheme    | V2 (set V4 explicitly)                                | SigV4 (only option)                                      |
| Signing identity  | JSON key OR IAM signBlob fallback                     | the calling principal's creds                            |

## TODO

1. Uncomment the block. Run `go run . <bucket> <existing-key>`.
2. `curl "$URL" > /tmp/out` — confirm the bytes match what's in GCS.
3. Wait 5+ minutes, `curl` again — confirm 403.
4. Try `--data-binary @some-file` against a PUT-signed URL (change `Method`)
   — confirm the upload lands.
