package exception

import (
	"errors"
	"net/http"
)

var (
	ErrTokenExpire                = errors.New("token expire")
	ErrInvalidToken               = errors.New("invalid token")
	ErrStorageSizeExceedLimitSize = errors.New("storage size exceed limit size")
	ErrStorageSizeLessThanZero    = errors.New("storage size cannot less than 0")
	ErrInvalidObjectId            = errors.New("object id is invalid")
	ErrParenFileNotExist          = errors.New("parent file not exist")
	ErrTagNotExist                = errors.New("tag not exist")
	ErrFileNotFound               = errors.New("file not found")
	ErrFileIsFolder               = errors.New("folder cannot be uploaded")
	ErrUnAllowedSortField         = errors.New("unallow sort field")
	ErrUserNotFound               = errors.New("user not found")
)
var errMap = map[error]int{
	ErrTokenExpire:                http.StatusUnauthorized,
	ErrInvalidToken:               http.StatusUnauthorized,
	ErrStorageSizeExceedLimitSize: http.StatusBadRequest,
	ErrInvalidObjectId:            http.StatusBadRequest,
	ErrParenFileNotExist:          http.StatusNotFound,
	ErrTagNotExist:                http.StatusNotFound,
	ErrFileNotFound:               http.StatusNotFound,
	ErrUserNotFound:               http.StatusNotFound,
	ErrFileIsFolder:               http.StatusConflict,
	ErrUnAllowedSortField:         http.StatusForbidden,
}

func ErrorStatusMapper(err error) int {
	for e, status := range errMap {
		if errors.Is(err, e) {
			return status
		}
	}

	return http.StatusInternalServerError
}
