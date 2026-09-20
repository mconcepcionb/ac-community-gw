package permissions

import (
	"errors"
	"testing"
)

func TestRegisterPermission(t *testing.T) {
	registry := NewRegistry()
	def := Definition{Name: "account.read", Description: "read", Owner: "test"}
	if err := registry.Register(def); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if !registry.Has("account.read") {
		t.Fatal("permission not registered")
	}
	if err := registry.Register(def); !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("expected ErrAlreadyRegistered, got %v", err)
	}
}

func TestRegisterRejectsEmptyName(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(Definition{}); !errors.Is(err, ErrEmptyName) {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestAuthorizerGrantsAndChecks(t *testing.T) {
	authorizer := NewAuthorizer()
	authorizer.Grant("admin", "azeroth.account.manage", "azeroth.account.read")
	authorizer.Grant("user", "azeroth.account.read")

	if !authorizer.Can([]Role{"admin"}, "azeroth.account.manage") {
		t.Fatal("admin should have manage")
	}
	if !authorizer.Can([]Role{"user", "admin"}, "azeroth.account.manage") {
		t.Fatal("union of roles should grant manage")
	}
	if authorizer.Can([]Role{"user"}, "azeroth.account.manage") {
		t.Fatal("user must not have manage")
	}
	if authorizer.Can(nil, "azeroth.account.read") {
		t.Fatal("no roles must not grant anything")
	}
}

func TestAuthorizerPermissionsAreSortedAndDeduplicated(t *testing.T) {
	authorizer := NewAuthorizer()
	authorizer.Grant("admin", "store.products.write", "azeroth.account.read")
	authorizer.Grant("user", "azeroth.account.read", "store.products.read")

	got := authorizer.Permissions([]Role{"admin", "user"})
	want := []Permission{"azeroth.account.read", "store.products.read", "store.products.write"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got = %v, want %v", got, want)
		}
	}

	if permissions := authorizer.Permissions(nil); len(permissions) != 0 {
		t.Fatalf("no roles must grant nothing, got %v", permissions)
	}
}

func TestAuthorizerReplaceRemovesStaleGrants(t *testing.T) {
	authorizer := NewAuthorizer()
	authorizer.Grant("admin", "azeroth.account.read", "azeroth.account.manage")

	authorizer.Replace(map[Role][]Permission{
		"admin": {"azeroth.account.read"},
	})

	if !authorizer.Can([]Role{"admin"}, "azeroth.account.read") {
		t.Fatal("retained permission should still be granted")
	}
	if authorizer.Can([]Role{"admin"}, "azeroth.account.manage") {
		t.Fatal("removed permission must no longer be granted")
	}
	if got := authorizer.Permissions([]Role{"admin"}); len(got) != 1 {
		t.Fatalf("permissions = %v, want exactly the retained one", got)
	}
}
