package map2struct

type BookType int64

const (
	BookType_STRIP      BookType = 1
	BookType_PAGE_LEFT  BookType = 2
	BookType_PAGE_RIGHT BookType = 3
)

type ApiBookInfo struct {
	Id             int64     `json:"book_id"`
	Name           string    `json:"book_name"`
	CopyrightInfo  string    `json:"copyright_info"`
	CreateTime     string    `json:"create_time"`
	SerialCount    *int32    `json:"serial_count,omitempty"`
	ThumbUrl       string    `json:"thumb_url"`
	BookType       *BookType `json:"book_type,omitempty"`
	LatestReadTime *int64    `json:"latest_read_time,omitempty"`
	Category       *string   `json:"category,omitempty"`
	IsFirstRead    bool      `json:"is_first_read,omitempty"`
}
