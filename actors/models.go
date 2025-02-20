package actors

type Serializable interface {
	Serialize() ([]byte, error)
	Parse([]byte) error
}
