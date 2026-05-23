# 01 — ADC and project ID

`google.FindDefaultCredentials(ctx, scope)` is the lower-level entry to
Application Default Credentials. Most Google Cloud client libraries call it
under the hood when you pass nothing to the constructor — the result is the
same `*google.Credentials` either way.

## The ADC chain

In order:

1. `GOOGLE_APPLICATION_CREDENTIALS` env var pointing at a service-account JSON
2. `gcloud auth application-default login` user creds at
   `~/.config/gcloud/application_default_credentials.json`
3. The GCE / Cloud Run / GKE / Cloud Functions metadata server (only when
   running on Google infrastructure)

Whichever step yields a credential first wins. There is no "source" field —
inspect `creds.JSON` (empty for metadata-server creds, populated for the
other two) if you really need to know.

## Project ID is separate

GCP has no account-ID equivalent baked into credentials. A token says "I am
service-account X" — it does not say "act on project Y." Project comes from
its own chain:

1. `GOOGLE_CLOUD_PROJECT` env var
2. The `project_id` field of the JSON key (when one is loaded)
3. The active `gcloud config get-value project` (when no key is loaded)
4. Metadata server (on GCP)

Most clients resolve this for you. `compute.NewInstancesRESTClient(ctx)`
takes the project per-call; `storage.NewClient(ctx)` doesn't need it at all
because bucket names are globally unique.

## Compare to AWS / Python

|                       | Go (`google.FindDefaultCredentials`)        | AWS Go (`config.LoadDefaultConfig`)   | Python (`google.auth.default()`) |
| --------------------- | ------------------------------------------- | ------------------------------------- | -------------------------------- |
| Where the chain lives | `golang.org/x/oauth2/google`                | `aws-sdk-go-v2/config`                | `google.auth`                    |
| Force a creds file    | `GOOGLE_APPLICATION_CREDENTIALS=/path`      | `AWS_PROFILE=p`                       | `GOOGLE_APPLICATION_CREDENTIALS` |
| Where project lives   | env / gcloud / JSON / metadata (4-way)      | not a thing (account ID in creds)     | same 4-way chain                 |
| Lazy or eager         | lazy — find != mint                         | lazy — load != retrieve               | lazy                             |

## TODO

1. Uncomment PART 1 and PART 2. Run `go run .` — note whether `creds.ProjectID` is
   set (it's empty when ADC fell back to user gcloud creds; populated when a
   service-account JSON is loaded).
2. Uncomment PART 3 — confirm a token mints in well under a second.
3. Run `GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json go run .` if you have a
   key handy — confirm `creds.ProjectID` is now populated.
4. Run `GOOGLE_APPLICATION_CREDENTIALS=/nonexistent go run .` — confirm the
   error mentions the env var (not just "creds not found").
