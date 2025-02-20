package operations

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

func Upload(bucketName string, logger *zap.Logger) (http.Handler, error) {
	// Verify the bucket exists and is setup as a public, otherwise fail.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, h, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url, err := UploadIO(r.Context(), h.Filename, f)
		if err != nil {
			logger.Error("Unable to upload", zap.String("file", h.Filename), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(url))
	}), nil
}

func UploadIO(ctxt context.Context, name string, f io.Reader) (string, error) {

	os.MkdirAll("assets", os.ModePerm)
	ext := filepath.Ext(name)
	ext = strings.ToLower(ext)
	ext = strings.TrimPrefix(ext, ".")

	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(f); err != nil {
		return "", err
	}
	hash := md5.Sum(buffer.Bytes())
	filename := hex.EncodeToString(hash[:]) + "." + ext
	os.WriteFile(filepath.Join("assets", filename), buffer.Bytes(), os.ModePerm)
	return filename, nil
}

func formatForExtension(ext string) imaging.Format {
	switch ext {
	case "jpg", "jpeg":
		return imaging.JPEG
	case "png":
		return imaging.PNG
	case "gif":
		return imaging.GIF
	case "tiff":
		return imaging.TIFF
	case "bmp":
		return imaging.BMP
	default:
		return imaging.JPEG
	}
}

func resizedImage(buffer *bytes.Buffer, w, h int, ext string) error {
	image, err := imaging.Decode(buffer, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	if w > 0 && h > 0 {
		newImage := imaging.Fill(image, w, h, imaging.Center, imaging.Lanczos)
		buffer.Reset()
		imaging.Encode(buffer, newImage, formatForExtension(ext))
	} else if w > 0 || h > 0 {
		newImage := imaging.Resize(image, w, h, imaging.Lanczos)
		buffer.Reset()
		imaging.Encode(buffer, newImage, formatForExtension(ext))
	} else {
		buffer.Reset()
		imaging.Encode(buffer, image, formatForExtension(ext))
	}
	return nil
}

func HashFromUrl(u string) (string, error) {
	uu, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return strings.Split(filepath.Base(uu.Path), ".")[0], nil
}
