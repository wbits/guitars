package images

import "errors"

var (
	errEmptySourceURL  = errors.New("source image url is empty")
	errDownloadFailed  = errors.New("source image download failed")
	errPresignFailed   = errors.New("image presign failed")
	errUploadFailed    = errors.New("image upload failed")
)
