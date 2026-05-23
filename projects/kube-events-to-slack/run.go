package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// newEventHandler returns a cache.ResourceEventHandler that runs the
// filter -> dedup -> format -> sink pipeline on each event.
//
// errOut receives non-fatal errors from sink.Send (so a single bad webhook
// POST doesn't kill the informer). filter, deduper, sink are closed over.
//
// The same closure is used for AddFunc and UpdateFunc — informers re-deliver
// existing items on resync, and an updated event with a higher .Count
// (e.g. CrashLoopBackOff happens again) should also trigger.
func newEventHandler(f Filter, d *Deduper, sink Sink, errOut io.Writer, now func() time.Time) cache.ResourceEventHandler {
	check := func(obj interface{}) {
		// TODO: run filter -> dedup -> format -> sink on this event. The
		//   informer hands you `interface{}`, so type-assert to *corev1.Event
		//   first (and just return on the unexpected-type path; this handler
		//   is only wired to Events). Sink errors go to errOut — don't return
		//   them, or you'll cause the informer to drop the item.
		_ = obj
	}
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    check,
		UpdateFunc: func(_, obj interface{}) { check(obj) },
	}
}

// Run starts the informer factory and blocks until ctx is cancelled.
//
// Namespace handling — the Filter.Namespaces field is the authoritative
// allow-list. The informer itself watches every namespace; the in-process
// filter is what restricts which events become alerts. This is simpler than
// configuring one informer per namespace and keeps the handler pure.
//
// Returns nil on clean shutdown, or an error if the informer caches never sync.
func Run(ctx context.Context, clientset kubernetes.Interface, f Filter, d *Deduper, sink Sink, errOut io.Writer) error {
	// TODO: build a SharedInformerFactory, attach newEventHandler to the
	//   core/v1 Events informer, Start the factory, and block on ctx.Done().
	//   The non-obvious bit is WaitForCacheSync — without it, the initial
	//   list-vs-add backfill will fire ALL existing events as "new" alerts,
	//   spamming Slack on every restart. Returning an error if the sync
	//   fails is part of the contract.

	_ = informers.NewSharedInformerFactoryWithOptions
	_ = cache.WaitForCacheSync
	_ = (*corev1.Event)(nil)
	_ = fmt.Sprintf
	return errors.New("Run not implemented")
}
