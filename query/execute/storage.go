package execute

type StorageReader interface {
	Read() (DataFrame, bool)
}
