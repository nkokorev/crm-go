package seeds

import "testing"

func TestPermissionSeeding(t *testing.T) {

	if len(permissions) < 1 {
		t.Error("Has not permissions for run seeding")
	}
}
