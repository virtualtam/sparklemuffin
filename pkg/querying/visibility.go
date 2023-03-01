package querying

type Visibility string

const (
	VisibilityAll     Visibility = "all"
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)
