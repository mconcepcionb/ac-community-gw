package gatewayadmin

import "testing"

func TestPermissionDefsIncludeAuditRead(t *testing.T) {
	var found bool
	for _, def := range permissionDefs() {
		if def.Name != PermissionAuditRead {
			continue
		}
		found = true
		if def.Namespace != "gw" {
			t.Fatalf("namespace = %q, want gw", def.Namespace)
		}
		if def.Owner != Name {
			t.Fatalf("owner = %q, want %q", def.Owner, Name)
		}
	}
	if !found {
		t.Fatal("permissionDefs() does not include gw.audit.read")
	}
}
