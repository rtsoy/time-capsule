package domain

type File struct {
	Bytes []byte `json:"-"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
}
