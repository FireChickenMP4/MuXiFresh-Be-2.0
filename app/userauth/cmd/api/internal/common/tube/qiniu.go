package tube

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"context"
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Keep upload credentials short-lived. The Qiniu SDK treats Expires as a
	// duration in seconds and converts it to an absolute deadline when signing.
	uploadTokenExpires uint64 = 5 * 60

	// Server-side captcha uploads remain image-only. Use an explicit raster-image
	// allowlist so active content such as HTML and SVG cannot be hosted on the
	// application's trusted storage domain.
	allowedImageMimeTypes = "image/jpeg;image/png;image/gif;image/webp;image/bmp;image/avif;image/heic;image/heif"

	// Client uploads also support passive document attachments. Keep this as an
	// allowlist: in particular, do not allow application/zip even though DOCX is
	// a ZIP container, because Qiniu can identify DOCX from its content and key.
	allowedDocumentMimeTypes = "application/pdf;application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	allowedClientUploadMimes = allowedImageMimeTypes + ";" + allowedDocumentMimeTypes
	maxUploadSize            = int64(10 << 20)
)

type Qiniu struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

var Q Qiniu

func Load(c config.Config) {
	Q = Qiniu{
		AccessKey: c.Infra.ObjectStorage.AccessKey,
		SecretKey: c.Infra.ObjectStorage.SecretKey,
		Bucket:    c.Infra.ObjectStorage.BucketName,
		Domain:    c.Infra.ObjectStorage.DomainName,
	}
}

func UploadFileToQiniu(localFilePath string) (string, error) {
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	cfg := storage.Config{
		Zone:          &storage.ZoneHuanan,
		UseCdnDomains: false,
		UseHTTPS:      true,
	}

	uploader := storage.NewFormUploader(&cfg)
	remoteFileName := fmt.Sprintf("captcha/%d-%s", time.Now().UnixNano(), filepath.Base(localFilePath))
	putPolicy := newServerImageUploadPolicy(Q.Bucket, remoteFileName)
	token := putPolicy.UploadToken(mac)
	ret := storage.PutRet{}
	err := uploader.PutFile(context.Background(), &ret, token, remoteFileName, localFilePath, nil)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(Q.Domain, "/") + "/" + ret.Key, nil
}

func GetQNToken(endUser string) string {
	putPolicy := newClientUploadPolicy(Q.Bucket, endUser)
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	return putPolicy.UploadToken(mac)
}

// newClientUploadPolicy intentionally keeps bucket-level scope for
// compatibility with deployed clients, which currently choose their own key.
// The policy still removes the reported arbitrary-file-upload vector by
// enforcing content detection, an image/PDF/DOCX allowlist, a size limit,
// short expiry and insert-only semantics.
func newClientUploadPolicy(bucket, endUser string) storage.PutPolicy {
	return storage.PutPolicy{
		Scope:      bucket,
		Expires:    uploadTokenExpires,
		InsertOnly: 1,
		EndUser:    endUser,
		FsizeLimit: maxUploadSize,
		DetectMime: 1,
		MimeLimit:  allowedClientUploadMimes,
	}
}

// Server-side uploads know the destination key in advance, so grant access to
// that exact object rather than the whole bucket.
func newServerImageUploadPolicy(bucket, key string) storage.PutPolicy {
	return storage.PutPolicy{
		Scope:        bucket + ":" + key,
		Expires:      uploadTokenExpires,
		InsertOnly:   1,
		ForceSaveKey: true,
		SaveKey:      key,
		FsizeLimit:   maxUploadSize,
		DetectMime:   1,
		MimeLimit:    allowedImageMimeTypes,
	}
}
