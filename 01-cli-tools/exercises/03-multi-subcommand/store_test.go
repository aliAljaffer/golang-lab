//go:build exercise

package multicmd

import (
	"errors"
	"testing"
)

func TestStore_CreateAndGet(t *testing.T) {
	s := NewStore()
	if err := s.CreatePod(Pod{Name: "web", Image: "nginx"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.GetPod("web")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Image != "nginx" {
		t.Errorf("image = %q, want nginx", got.Image)
	}
}

func TestStore_CreateDuplicate(t *testing.T) {
	s := NewStore()
	_ = s.CreatePod(Pod{Name: "web", Image: "nginx"})
	err := s.CreatePod(Pod{Name: "web", Image: "other"})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("got %v, want ErrAlreadyExists", err)
	}
}

func TestStore_GetMissing(t *testing.T) {
	s := NewStore()
	_, err := s.GetPod("nope")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestStore_ListSorted(t *testing.T) {
	s := NewStore()
	_ = s.CreatePod(Pod{Name: "c"})
	_ = s.CreatePod(Pod{Name: "a"})
	_ = s.CreatePod(Pod{Name: "b"})

	pods := s.ListPods()
	if len(pods) != 3 {
		t.Fatalf("len=%d, want 3", len(pods))
	}
	for i, want := range []string{"a", "b", "c"} {
		if pods[i].Name != want {
			t.Errorf("[%d] %q, want %q", i, pods[i].Name, want)
		}
	}
}

func TestStore_Delete(t *testing.T) {
	s := NewStore()
	_ = s.CreatePod(Pod{Name: "web"})
	if err := s.DeletePod("web"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetPod("web"); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete: got %v, want ErrNotFound", err)
	}
}

func TestStore_DeleteMissing(t *testing.T) {
	s := NewStore()
	err := s.DeletePod("ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}
