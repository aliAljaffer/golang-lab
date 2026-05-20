# Plan: projects/

Capstones combining multiple sections. Each should be a non-trivial mini-tool that could plausibly exist in a real environment.

## Projects

### 1. `kube-events-to-slack`
- **Combines:** 04 (HTTP servers), 05 (concurrency), 08 (kubernetes)
- **What:** watches k8s events via an informer; filters by severity/namespace; posts formatted alerts to a Slack webhook
- **Why:** real ops use case; teaches informer + webhook patterns end-to-end

### 2. `s3-log-shipper`
- **Combines:** 02 (files/os), 03 (http-clients), 05 (concurrency), 07 (AWS)
- **What:** tails local log files, batches lines, gzips, uploads to S3 on a schedule
- **Why:** the "Fluentbit lite" pattern; teaches batching, backpressure, retries

### 3. `deploy-bot`
- **Combines:** 01 (CLI), 03 (http-clients), 09 (docker)
- **What:** CLI that: fetches a GitHub release artifact, builds a Docker image from it, runs the container, reports health
- **Why:** ties together API calls, container ops, real human-facing CLI

## Status

- [ ] `kube-events-to-slack` built
- [ ] `s3-log-shipper` built
- [ ] `deploy-bot` built
