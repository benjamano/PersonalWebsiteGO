package background

import (
	"PersonalWebsiteGO/handlers"

	"fmt"
	"time"
)

func StartPlaytimeChecker() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			fmt.Println("Checking player playtime...")

			handlers.CheckAndUpdatePlaytime()

			<-ticker.C
		}
	}()
}
