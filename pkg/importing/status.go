package importing

import "fmt"

type Status struct {
	overwriteExisting bool

	Invalid      int
	NewOrUpdated int
	Skipped      int
}

func (st *Status) Summary() string {
	var orUpdated string
	if st.overwriteExisting {
		orUpdated = " or updated"
	}

	return fmt.Sprintf(
		"%d new%s, %d skipped, %d invalid",
		st.NewOrUpdated,
		orUpdated,
		st.Skipped,
		st.Invalid,
	)
}
