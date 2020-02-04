package img

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/768bit/vutils"
)

func EstablishPathToSource(path string) (bool, string) {
	_, err := url.ParseRequestURI(path)
	if err != nil {
		//perhaps it is missing the first parts.. perhaps it is a file path.. this needs to be established

		if len(path) > 1 && (path[0] == '/' || path[0:1] == "./" || (len(path) > 2 && path[0:2] == "../")) {
			//this is definitely a local path
			return checkLocalPath(path)
		} else if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
			//lets try using http/https to lookup path, however, we should see if the path that is supplied exists locally, i.e. does the first portion of the path exist in CWD

			cwd, _ := os.Getwd()

			spl := strings.Split(path, "/")

			fullPath := filepath.Join(cwd, spl[0])

			if vutils.Files.CheckPathExists(fullPath) {
				return checkLocalPath(filepath.Join(cwd, path))
			}

			//otherwise it is a case of rerunning this function but using https then http...

			if ex, opath := EstablishPathToSource("https://" + path); opath == "" {
				return EstablishPathToSource("http://" + path)
			} else {
				return ex, opath
			}

		}

		return false, ""

	}

	u, err := url.Parse(path)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false, ""
	}

	return false, path

}

func checkLocalPath(path string) (exists bool, fullPath string) {
	exists = false
	var err error = nil
	cwd, _ := os.Getwd()
	if path[0] == '/' {
		fullPath = path
	} else if path[0:1] == "./" {
		fullPath, err = filepath.Abs(filepath.Join(cwd, path[2:]))
		if err != nil {
			return
		}
	} else if path[0:2] == "../" {
		fullPath, err = filepath.Abs(filepath.Join(cwd, path))
		if err != nil {
			return
		}
	}
	if vutils.Files.CheckPathExists(fullPath) {
		exists = true
	} else {
		fullPath = ""
	}
	return
}
