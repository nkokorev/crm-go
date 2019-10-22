package seeds

import (
	"os"
)

func init() {

	// seeding can be: "" / "true" / "fresh"
	seeding := os.Getenv("seeding")
	if seeding == "true" ||  seeding == "fresh"{
		//UserSeeding()
		//PermissionSeeding()
	}
}

