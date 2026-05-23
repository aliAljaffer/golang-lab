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
		// TODO: type-assert obj to *corev1.Event.
		//       Handle cache.DeletedFinalStateUnknown if you want defensive code,
		//       though it's only emitted on Delete, not Add/Update.
		// TODO: if !f.ShouldAlert(event) { return }.
		// TODO: key := DedupKey(event); if !d.ShouldAlert(key) { return }.
		// TODO: alert := FormatSlackMessage(event, now()).
		// TODO: if err := sink.Send(context.Background(), alert); err != nil { fmt.Fprintln(errOut, "sink:", err) }.
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
	// TODO: factory := informers.NewSharedInformerFactoryWithOptions(clientset, 30*time.Second).
	// TODO: eventInformer := factory.Core().V1().Events().Informer().
	// TODO: _, err := eventInformer.AddEventHandler(newEventHandler(f, d, sink, errOut, time.Now))
	// TODO: if err != nil { return err }.
	// TODO: factory.Start(ctx.Done()).
	// TODO: if !cache.WaitForCacheSync(ctx.Done(), eventInformer.HasSynced) { return errors.New("event informer cache sync failed") }.
	// TODO: <-ctx.Done(); return nil.

	// keep the imports live during scaffolding
	_ = informers.NewSharedInformerFactoryWithOptions
	_ = cache.WaitForCacheSync
	_ = (*corev1.Event)(nil)
	_ = fmt.Sprintf
	return errors.New("Run not implemented")
}
