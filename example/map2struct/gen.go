package map2struct

import (
	"context"

	"github.com/adyzng/gotool/example/model"
)

//go:generate map2struct

// MapToBookInfo ...
func MapToBookInfo(ctx context.Context, src map[string]interface{}) (*model.ApiBookInfo, error) {
	return nil, nil
}
