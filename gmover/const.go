package gmover

import (
	"time"
)

// LabelsToAdd are automatically added to all moved messages for safety
var LabelsToAdd = []string{"[Gmoved]", "Moved-" + time.Now().Format(time.DateOnly)}
