# 01 — load config

Every client-go program starts the same way: build a `*rest.Config`, then hand
it to `kubernetes.NewForConfig(cfg)` to get a `*kubernetes.Clientset`. From
there you call `.CoreV1()`, `.AppsV1()`, etc.

## In-cluster vs out-of-cluster

| Where | How |
|---|---|
| Inside a pod | `rest.InClusterConfig()` — reads ServiceAccount token + CA from `/var/run/secrets/kubernetes.io/serviceaccount/` |
| Your laptop  | `clientcmd.BuildConfigFromFlags("", path)` — loads `~/.kube/config` |

The `loadConfig()` helper in `main.go` tries in-cluster first and falls back
to kubeconfig. This is the boilerplate every production tool ships with.

## Discovery as a smoke test

`clientset.Discovery().ServerVersion()` hits `/version` on the API server.
It's the cheapest way to confirm:

1. Network reachability to the API server
2. TLS trust (your kubeconfig CA matches the server cert)
3. Authentication (your token/cert is accepted — but NOT authorization, since
   `/version` is unauthenticated on most clusters)

## Compare to other clients

|                       | Go (`client-go`)                     | Python (`kubernetes`)                |
|-----------------------|--------------------------------------|--------------------------------------|
| Load config           | `clientcmd.BuildConfigFromFlags`     | `config.load_kube_config()`          |
| In-cluster config     | `rest.InClusterConfig()`             | `config.load_incluster_config()`     |
| Build client          | `kubernetes.NewForConfig(cfg)`       | `client.CoreV1Api()`                 |
| Server version probe  | `Discovery().ServerVersion()`        | `core_v1.get_api_resources()`        |

## TODO

1. Uncomment the TODO block.
2. Run `go run .` against a local cluster (`minikube start` or `kind create cluster`).
3. Try `KUBECONFIG=/dev/null go run .` and observe the error path.
4. Print the available API groups via `clientset.Discovery().ServerGroups()`.
