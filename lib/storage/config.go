package storage

type StorageConfig struct {
	ID     string                 `json:"id"`
	Driver string                 `json:"driver"`
	Config map[string]interface{} `json:"config"`
}
