//go:build windows

package platform

import (
	"context"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	p := newPlatform().(*WindowsPlatform)
	
	// We don't assert true or false because it depends on the environment running the test,
	// but we assert it doesn't panic.
	isAdmin := p.IsAdmin()
	t.Logf("IsAdmin returned: %v", isAdmin)
}

func TestRequireAdmin(t *testing.T) {
	p := newPlatform().(*WindowsPlatform)
	if p.IsAdmin() {
		t.Log("Already admin, RequireAdmin should return nil")
		err := p.RequireAdmin(context.Background())
		if err != nil {
			t.Errorf("RequireAdmin returned error when already admin: %v", err)
		}
	} else {
		// If we are not admin, RequireAdmin will try to spawn a UAC prompt.
		// Since we don't want to block CI or manual test execution with a UAC prompt,
		// we skip the actual execution in unit tests or just run it with a timeout.
		t.Log("Not admin, skipping UAC prompt test to avoid hanging.")
	}
}
