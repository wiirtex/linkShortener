package memory

import "time"

type Memory interface {
	AddEntry(entry MemoryRequest) (int64, error)
	GetEntry(entry MemoryRequest) (MemoryRequest, error)
	Clear() error
}

type MemoryRequest struct {
	Short     int64     `json:"short"`
	Long      string    `json:"long"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}
