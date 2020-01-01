package common

type VmmManager interface {
	Create()
	Destroy(id string)
	Update(id string)
	Get(id string)
}
