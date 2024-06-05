package providers

type Provider interface {
	FetchData() (interface{}, error)
	PrepareData(data interface{}) (interface{}, error)
}
