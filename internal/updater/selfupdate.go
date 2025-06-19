package updater

import (
	"fmt"
)


func SelfUpdate() {
	fmt.Println("Checking for updates...")
	err := UpdateTrack()
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return
	}
	fmt.Println("track was updated successfully!")
}
