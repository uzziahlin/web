package web

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	lru "github.com/hashicorp/golang-lru"
)

type FileUploader struct {
	FileField string
	PathFunc  func(header *multipart.FileHeader) string
}

func (f *FileUploader) Handle() HandleFunc {

	if f.FileField == "" {
		f.FileField = "file"
	}

	if f.PathFunc == nil {
		f.PathFunc = func(h *multipart.FileHeader) string {
			return filepath.Join("testdata", h.Filename)
		}
	}

	return func(ctx *Context) {

		// 取出表单中的文件
		srcFile, header, err := ctx.Req.FormFile(f.FileField)

		if err != nil {
			ctx.RespStatus = http.StatusInternalServerError
			ctx.RespData = []byte("文件上传失败")
			return
		}

		// 调用用户传入函数，获取问价存储路径
		path := f.PathFunc(header)

		// 打开目标文件，不存在则创建，存在则清空
		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o666)

		if err != nil {
			ctx.RespStatus = http.StatusInternalServerError
			ctx.RespData = []byte("文件上传失败")
			return
		}

		// 将文件保存到指定路径
		_, err = io.Copy(destFile, srcFile)

		if err != nil {
			ctx.RespStatus = http.StatusInternalServerError
			ctx.RespData = []byte("文件上传失败")
			return
		}

	}
}

type FileDownloader struct {
	Dir       string
	FileField string
}

func (f *FileDownloader) Handle() HandleFunc {

	if f.FileField == "" {
		f.FileField = "file"
	}

	return func(ctx *Context) {
		query := ctx.GetQuery(f.FileField)

		if query.err != nil {
			ctx.RespStatus = http.StatusInternalServerError
			ctx.RespData = []byte("文件下载失败")
			return
		}

		destPath := filepath.Clean(query.data)

		destPath = filepath.Join(f.Dir, destPath)

		fn := filepath.Base(destPath)

		header := ctx.Resp.Header()

		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")

		http.ServeFile(ctx.Resp, ctx.Req, destPath)

	}
}

type StaticResourceHandlerOption func(handler *StaticResourceHandler)

func NewStaticResourceHandler(baseDir string, opts ...StaticResourceHandlerOption) (*StaticResourceHandler, error) {
	// 默认1000
	cache, err := lru.New(1000)

	if err != nil {
		return nil, err
	}

	handler := &StaticResourceHandler{
		BaseDir: baseDir,
		cache:   cache,
		maxSize: 1024 * 1024 * 10,
		ExtContentTypeMap: map[string]string{
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler, nil
}

func StaticWithMaxFileSize(maxSize int) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.maxSize = maxSize
	}
}

func StaticWithCache(c *lru.Cache) StaticResourceHandlerOption {
	return func(handler *StaticResourceHandler) {
		handler.cache = c
	}
}

func StaticWithMoreExtension(extMap map[string]string) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		for ext, contentType := range extMap {
			h.ExtContentTypeMap[ext] = contentType
		}
	}
}

type StaticResourceHandler struct {
	BaseDir           string
	SubDir            map[string]string
	FileParam         string
	ExtContentTypeMap map[string]string
	cache             *lru.Cache
	maxSize           int
}

func (s *StaticResourceHandler) Handle(ctx *Context) {
	fileName, ok := ctx.PathParams[s.FileParam]

	if !ok {
		ctx.RespStatus = http.StatusInternalServerError
		ctx.RespData = []byte("文件不存在")
		return
	}

	path := s.getPath(fileName)

	if value, ok := s.cache.Get(path); ok {
		s.setHeader(ctx.Resp, filepath.Ext(fileName), len(value.([]byte)))
		ctx.RespStatus = http.StatusOK
		ctx.RespData = value.([]byte)
		return
	}

	file, err := os.ReadFile(path)

	if err != nil {
		ctx.RespStatus = http.StatusInternalServerError
		ctx.RespData = []byte("文件不存在")
		return
	}

	if len(file) <= s.maxSize {
		s.cache.Add(path, file)
	}

	s.setHeader(ctx.Resp, filepath.Ext(fileName), len(file))

	ctx.RespStatus = http.StatusOK
	ctx.RespData = file
}

func (s *StaticResourceHandler) setHeader(w http.ResponseWriter, ext string, contentLength int) {
	header := w.Header()
	contentType := s.ExtContentTypeMap[ext]
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", strconv.Itoa(contentLength))
}

func (s *StaticResourceHandler) getPath(fileName string) string {
	// 获取文件后缀名
	ext := filepath.Ext(fileName)

	//根据文件类型，获取文件所在子目录
	subPath := s.SubDir[ext]

	return filepath.Join(s.BaseDir, subPath, fileName)
}
