package importing

import "fmt"

type Status struct {
	Invalid int
	New     int
	Skipped int
}

func (st *Status) Summary() string {
	return fmt.Sprintf(
		"%d new, %d skipped, %d invalid",
		st.New,
		st.Skipped,
		st.Invalid,
	)
}
