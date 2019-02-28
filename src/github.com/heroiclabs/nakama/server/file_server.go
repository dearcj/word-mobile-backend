package server

import (
	"go.uber.org/zap"
	"net/http"
	"path"
	"path/filepath"
)

type FileServer struct {
	fs http.Handler
	wx *WordX
}

const FS_PORT = ":7352"

func StartFileServer(logger *zap.Logger, globalPath string, wx *WordX) *FileServer {

	filePathFull := filepath.Join(globalPath, "client/build/")
	fs := http.FileServer(http.Dir(filePathFull))
	logger.Info("File server listening on port " + FS_PORT)

	go func() {
		http.HandleFunc("/image/", func(writer http.ResponseWriter, request *http.Request) {
			id := path.Base(request.URL.Path)
			err, im := wx.Cache.GetImage(id)
			if err == nil {
				writer.Write([]byte(*im))
			} else {
				str := "Wrong category id"
				writer.Write([]byte(str))
			}

		})
		http.Handle("/", fs)

		err := http.ListenAndServe(FS_PORT, nil)

		if err != nil {
			logger.Error("Can't start file server")
		}
	}()

	return &FileServer{
		fs: fs,
		wx: wx,
	}
}

func (fs *FileServer) Stop() {
	//	fs.fs.
}
