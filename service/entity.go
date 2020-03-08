package service

type Page struct {
	PageIndex uint32
	PageSize  uint32
}

type Target struct {
	Id     uint64
	Title  string
	Score  uint32
	Symbol string
}
