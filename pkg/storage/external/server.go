package external

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gorilla/mux"
	"golang.org/x/mod/zip"
)

// NewServer takes a storage.Backend implementation of your
// choice, and returns a new http.Handler that Athens can
// reach out to for storage operations
func NewServer(strg storage.Backend) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc(download.PathList, func(w http.ResponseWriter, r *http.Request) {
		mod := mux.Vars(r)["module"]
		list, err := strg.List(r.Context(), mod)
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
			return
		}
		fmt.Fprintf(w, "%s", strings.Join(list, "\n"))
	})
	r.HandleFunc(download.PathVersionInfo, func(w http.ResponseWriter, r *http.Request) {
		params, err := paths.GetAllParams(r)
		if err != nil {
			return
		}
		info, err := strg.Info(r.Context(), params.Module, params.Version)
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
			return
		}
		w.Write(info)
	})
	r.HandleFunc(download.PathVersionModule, func(w http.ResponseWriter, r *http.Request) {
		params, err := paths.GetAllParams(r)
		if err != nil {
			return
		}
		mod, err := strg.GoMod(r.Context(), params.Module, params.Version)
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
			return
		}
		w.Write(mod)
	})
	r.HandleFunc(download.PathVersionZip, func(w http.ResponseWriter, r *http.Request) {
		params, err := paths.GetAllParams(r)
		if err != nil {
			return
		}
		zip, err := strg.Zip(r.Context(), params.Module, params.Version)
		if err != nil {
			http.Error(w, err.Error(), errors.Kind(err))
			return
		}
		defer zip.Close()
		io.Copy(w, zip)
	})
	r.HandleFunc("/{module:.+}/@v/{version}.upload", func(w http.ResponseWriter, r *http.Request) {
		params, err := paths.GetAllParams(r)
		if err != nil {
			return
		}
		err = r.ParseMultipartForm(zip.MaxZipFile + zip.MaxGoMod)
		if err != nil {
			fmt.Printf("parse: %v\n", err)
			return
		}
		infoFile, header, err := r.FormFile("mod.info")
		if err != nil {
			fmt.Printf("info: %v\n", err)
			return
		}
		defer infoFile.Close()
		info, err := ioutil.ReadAll(infoFile)
		if err != nil {
			return
		}
		modReader, header, err := r.FormFile("mod.mod")
		if err != nil {
			return
		}
		defer modReader.Close()
		modFile, err := ioutil.ReadAll(modReader)
		if err != nil {
			return
		}
		modZ, header, err := r.FormFile("mod.zip")
		if err != nil {
			fmt.Printf("mod.zip: %v\n", err)
			return
		}
		defer modZ.Close()
		err = strg.Save(r.Context(), params.Module, params.Version, modFile, modZ, info)
		if err != nil {
			fmt.Printf("save: %v\n", err)
		}
	})
	return r
}
