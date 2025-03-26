package presenters

import "github.com/baothaihcmut/Bibox/storage-app/internal/common/response"

type GetFileMetaDataInput struct {
	Id string `uri:"id" binding:"required"`
}

type GetFileMetaDataOuput struct {
	*response.FileOutput
}
