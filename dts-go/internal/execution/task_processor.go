package execution

type TaskProcessor interface {
	Start(topic string) error
	Stop() error
}
