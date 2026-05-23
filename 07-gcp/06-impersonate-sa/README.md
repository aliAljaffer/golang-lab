# 06 — service account impersonation

GCP's answer to AWS `sts:AssumeRole`. Your caller (any principal — user,
workload identity, CI SA) mints short-lived tokens for a target SA, then
uses those tokens to make API calls "as" the target.

## The three pieces

```go
ts, _ := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
    TargetPrincipal: "readonly@my-project.iam.gserviceaccount.com",
    Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
    Lifetime:        15 * time.Minute,
})

client, _ := storage.NewClient(ctx, option.WithTokenSource(ts))
// every call client makes is signed as readonly@..., not as the caller
```

1. **Token source** — calls `iamcredentials.generateAccessToken` to mint a
   token for the target SA. Caches the token in memory; refreshes before
   expiry.
2. **`option.WithTokenSource(ts)`** — the "use these creds" knob that every
   GCP Go client constructor accepts. Without it, the client falls back to
   ADC. With it, you've replaced the credential chain for this client only.
3. **The target SA's permissions** — IAM bindings on the target are what
   determine what the client can do. The caller's bindings are now irrelevant
   for any call this client makes.

## The IAM bridge

The caller principal needs `roles/iam.serviceAccountTokenCreator` on the
target SA. That role grants `iam.serviceAccounts.getAccessToken`, which is
what `generateAccessToken` calls under the hood. **Bind the role on the
target SA, not at the project level** — project-level grants let the caller
impersonate every SA in the project, which is rarely what you want.

```bash
gcloud iam service-accounts add-iam-policy-binding \
  readonly@my-project.iam.gserviceaccount.com \
  --member="user:you@example.com" \
  --role="roles/iam.serviceAccountTokenCreator"
```

## Why short-lived

`Lifetime` caps the token at 15 minutes (max 12 hours, with org policy
overrides). Compared to a long-lived JSON key, this is enormously safer:
a leaked impersonation token expires fast; a leaked SA key is valid until
someone deletes it.

The `CredentialsTokenSource` caches across calls — you don't pay the IAM
round-trip per API call. The cache lives inside `ts`; share it across all
clients in your process.

## Compare to AWS

| Concept              | Go (GCP)                                                     | Go (AWS)                                              |
| -------------------- | ------------------------------------------------------------ | ----------------------------------------------------- |
| Cross-identity creds | `impersonate.CredentialsTokenSource(ctx, cfg)`               | `stscreds.NewAssumeRoleProvider(stsClient, roleArn)` |
| Wrap with cache      | built into the token source                                  | `aws.NewCredentialsCache(provider)` — explicit       |
| Pass to client       | `option.WithTokenSource(ts)`                                 | `cfg.Credentials = cache; client.NewFromConfig(cfg)`  |
| Permission needed    | `iam.serviceAccountTokenCreator` on target SA                | `sts:AssumeRole` allowed by target role's trust policy |
| Max lifetime         | 12 hours (configurable; default 1 hour, request up to 12h)   | up to 12 hours via `DurationSeconds`                  |

## When to use this

- **Local dev hitting a project you don't own creds for.** Impersonate the
  project's SA from your gcloud login instead of mounting a JSON key.
- **CI signing as a least-priv SA.** Your CI SA gets `tokenCreator` on the
  deploy SA; CI impersonates the deploy SA only for the deploy step.
- **Anywhere you'd reach for a JSON SA key.** Impersonation almost always
  replaces it. Keys are credentials sprawl waiting to happen.

## TODO

1. Pick a target SA (or create one: `gcloud iam service-accounts create demo-target --display-name="impersonation target"`).
2. Grant yourself token-creator on it: see the gcloud command above.
3. Grant the target SA `roles/storage.objectViewer` on the project (so the
   smoke test succeeds).
4. Uncomment the block. Run `go run . demo-target@<project>.iam.gserviceaccount.com <project>`.
5. Revoke `roles/storage.objectViewer` from the target — confirm the same
   command now fails with a Permission Denied (proves you're signing as the
   target, not the caller).
