package postgresql

type Tag struct {
	Name  string `db:"name"`
	Count uint   `db:"count"`
}
