package config

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func newTestManagerWithDevices(devices ...string) *Manager {
	cfg := &Config{}
	mgr := NewTestManager(cfg)
	for _, name := range devices {
		mgr.config.Devices[name] = model.Device{Name: name, Address: "192.168.1.1"}
	}
	return mgr
}

func TestSetLinkBasic(t *testing.T) {
	t.Parallel()

	mgr := newTestManagerWithDevices("bulb-duo", "bedroom-2pm")

	if err := mgr.SetLink("bulb-duo", "bedroom-2pm", 0); err != nil {
		t.Fatalf("SetLink: %v", err)
	}

	link, ok := mgr.GetLink("bulb-duo")
	if !ok {
		t.Fatal("expected link to exist")
	}
	if link.ParentDevice != "bedroom-2pm" {
		t.Errorf("parent = %q, want %q", link.ParentDevice, "bedroom-2pm")
	}
	if link.SwitchID != 0 {
		t.Errorf("switch_id = %d, want 0", link.SwitchID)
	}
}

func TestSetLinkWithSwitchID(t *testing.T) {
	t.Parallel()

	mgr := newTestManagerWithDevices("garage-light", "garage-switch")

	if err := mgr.SetLink("garage-light", "garage-switch", 1); err != nil {
		t.Fatalf("SetLink: %v", err)
	}

	link, ok := mgr.GetLink("garage-light")
	if !ok {
		t.Fatal("expected link to exist")
	}
	if link.SwitchID != 1 {
		t.Errorf("switch_id = %d, want 1", link.SwitchID)
	}
}

func TestSetLinkValidation(t *testing.T) {
	t.Parallel()

	t.Run("reject self-link", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("device-a")

		err := mgr.SetLink("device-a", "device-a", 0)
		if err == nil {
			t.Fatal("expected error for self-link")
		}
	})

	t.Run("reject chain link", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("device-a", "device-b", "device-c")

		if err := mgr.SetLink("device-a", "device-b", 0); err != nil {
			t.Fatalf("SetLink A->B: %v", err)
		}

		err := mgr.SetLink("device-c", "device-a", 0)
		if err == nil {
			t.Fatal("expected error for chain link (C->A where A is already a child)")
		}
	})

	t.Run("reject child device not found", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("parent")

		err := mgr.SetLink("nonexistent", "parent", 0)
		if err == nil {
			t.Fatal("expected error for nonexistent child")
		}
	})

	t.Run("reject parent device not found", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("child")

		err := mgr.SetLink("child", "nonexistent", 0)
		if err == nil {
			t.Fatal("expected error for nonexistent parent")
		}
	})
}

func TestSetLinkUpdateAndMultiple(t *testing.T) {
	t.Parallel()

	t.Run("update existing link", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb", "switch-a", "switch-b")

		if err := mgr.SetLink("bulb", "switch-a", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}
		if err := mgr.SetLink("bulb", "switch-b", 1); err != nil {
			t.Fatalf("SetLink update: %v", err)
		}

		link, ok := mgr.GetLink("bulb")
		if !ok {
			t.Fatal("expected link to exist")
		}
		if link.ParentDevice != "switch-b" || link.SwitchID != 1 {
			t.Errorf("link = %+v, want parent=switch-b switch_id=1", link)
		}
	})

	t.Run("multiple children same parent", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb-1", "bulb-2", "parent-switch")

		if err := mgr.SetLink("bulb-1", "parent-switch", 0); err != nil {
			t.Fatalf("SetLink bulb-1: %v", err)
		}
		if err := mgr.SetLink("bulb-2", "parent-switch", 0); err != nil {
			t.Fatalf("SetLink bulb-2: %v", err)
		}

		children := mgr.GetLinkedChildren("parent-switch")
		if len(children) != 2 {
			t.Errorf("expected 2 children, got %d", len(children))
		}
	})
}

func TestDeleteLink(t *testing.T) {
	t.Parallel()

	t.Run("basic delete", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb", "switch")
		if err := mgr.SetLink("bulb", "switch", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}

		if err := mgr.DeleteLink("bulb"); err != nil {
			t.Fatalf("DeleteLink: %v", err)
		}

		if _, ok := mgr.GetLink("bulb"); ok {
			t.Error("expected link to be deleted")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb")

		err := mgr.DeleteLink("bulb")
		if err == nil {
			t.Fatal("expected error for nonexistent link")
		}
	})
}

func TestListLinks(t *testing.T) {
	t.Parallel()

	mgr := newTestManagerWithDevices("bulb-1", "bulb-2", "switch-a", "switch-b")
	if err := mgr.SetLink("bulb-1", "switch-a", 0); err != nil {
		t.Fatalf("SetLink: %v", err)
	}
	if err := mgr.SetLink("bulb-2", "switch-b", 1); err != nil {
		t.Fatalf("SetLink: %v", err)
	}

	links := mgr.ListLinks()
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}

	link1, ok := links["bulb-1"]
	if !ok {
		t.Fatal("expected link for bulb-1")
	}
	if link1.ParentDevice != "switch-a" {
		t.Errorf("bulb-1 parent = %q, want %q", link1.ParentDevice, "switch-a")
	}
}

func TestGetLinkedChildren(t *testing.T) {
	t.Parallel()

	mgr := newTestManagerWithDevices("bulb-1", "bulb-2", "bulb-3", "switch")
	if err := mgr.SetLink("bulb-1", "switch", 0); err != nil {
		t.Fatalf("SetLink: %v", err)
	}
	if err := mgr.SetLink("bulb-2", "switch", 0); err != nil {
		t.Fatalf("SetLink: %v", err)
	}

	children := mgr.GetLinkedChildren("switch")
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}

	// bulb-3 has no link
	children = mgr.GetLinkedChildren("bulb-3")
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}

func TestCascadingDeleteRemovesLinks(t *testing.T) {
	t.Parallel()

	t.Run("delete child device removes its link", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb", "switch")
		if err := mgr.SetLink("bulb", "switch", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}

		if err := mgr.UnregisterDevice("bulb"); err != nil {
			t.Fatalf("UnregisterDevice: %v", err)
		}

		links := mgr.ListLinks()
		if len(links) != 0 {
			t.Errorf("expected 0 links after deleting child, got %d", len(links))
		}
	})

	t.Run("delete parent device removes its children links", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb-1", "bulb-2", "switch")
		if err := mgr.SetLink("bulb-1", "switch", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}
		if err := mgr.SetLink("bulb-2", "switch", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}

		if err := mgr.UnregisterDevice("switch"); err != nil {
			t.Fatalf("UnregisterDevice: %v", err)
		}

		links := mgr.ListLinks()
		if len(links) != 0 {
			t.Errorf("expected 0 links after deleting parent, got %d", len(links))
		}
	})
}

func TestCascadingRenameUpdatesLinks(t *testing.T) {
	t.Parallel()

	t.Run("rename child device updates link key", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("old-bulb", "switch")
		if err := mgr.SetLink("old-bulb", "switch", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}

		if err := mgr.RenameDevice("old-bulb", "new-bulb"); err != nil {
			t.Fatalf("RenameDevice: %v", err)
		}

		if _, ok := mgr.GetLink("old-bulb"); ok {
			t.Error("expected old link key to be removed")
		}
		link, ok := mgr.GetLink("new-bulb")
		if !ok {
			t.Fatal("expected link under new key")
		}
		if link.ParentDevice != "switch" {
			t.Errorf("parent = %q, want %q", link.ParentDevice, "switch")
		}
	})

	t.Run("rename parent device updates link references", func(t *testing.T) {
		t.Parallel()
		mgr := newTestManagerWithDevices("bulb", "old-switch")
		if err := mgr.SetLink("bulb", "old-switch", 0); err != nil {
			t.Fatalf("SetLink: %v", err)
		}

		if err := mgr.RenameDevice("old-switch", "new-switch"); err != nil {
			t.Fatalf("RenameDevice: %v", err)
		}

		link, ok := mgr.GetLink("bulb")
		if !ok {
			t.Fatal("expected link to exist")
		}
		if link.ParentDevice != "new-switch" {
			t.Errorf("parent = %q, want %q", link.ParentDevice, "new-switch")
		}
	})
}
