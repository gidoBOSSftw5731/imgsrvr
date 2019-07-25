package selector

import (
	"fmt"
	"net/http"
	"path"

	"../sessions"
	"./tools"
	"github.com/gidoBOSSftw5731/log"
)

//Caseable is a struct for passing data to SwitchStatement
type Caseable struct {
	URLSplit                                        []string
	URLECount, NumberOfPrefixSlashes, SwitchLen, I1 int
	Resp                                            http.ResponseWriter
	Req                                             *http.Request
}

//SwitchStatement is a function holding a switch statement. it is in its own file so that
// webmasters may edit it to fit their configuration
func SwitchStatement(config tools.Config, obj Caseable) {
	switch obj.URLSplit[obj.SwitchLen] {
	case "i", "I":
		// Checks for hash/element/thing
		if obj.URLSplit[obj.I1] == "" {
			tools.ErrorHandler(obj.Resp, obj.Req, http.StatusNotFound)
			return
		}
		log.Tracef("urlECount of IMG: %d\n", obj.URLECount)
		log.Tracef("Split for image: %v\n", obj.URLSplit)
		tools.SendImg(obj.Resp, obj.Req, obj.URLSplit[obj.I1], config)
		//upload(obj.Resp, obj.Req)
	case "upload":
		log.Traceln("Upload selected")
		tools.Upload(obj.Resp, obj.Req, config)
	case "favicon.ico", "favicon-16x16.png", "favicon-32x32.png", "favicon-96x96.png", "favicon-256x256" +
		".png", "android-icon-192x192.png", "apple-icon-114x114.png", "apple-icon-120x120.png", "apple-icon-" +
		"144x144.png", "apple-icon-152x152.png", "apple-icon-180x180.png", "apple-icon-57x57.png", "apple-icon-" +
		"60x60.png", "apple-icon-72x72.png", "apple-icon-76x76.png", "ms-icon-144x144.png", "ms-icon-150x150" +
		".png", "ms-icon-310x310.png", "ms-icon-70x70.png": // case for favicons
		http.ServeFile(obj.Resp, obj.Req, "favicons/"+obj.URLSplit[obj.SwitchLen])
	case "robots.txt":
		http.ServeFile(obj.Resp, obj.Req, "robots.txt")

	case "css":
		http.ServeFile(obj.Resp, obj.Req, "server/"+obj.URLSplit[obj.SwitchLen+1])
	case "js":
		i := obj.SwitchLen + 2
		if i >= len(obj.URLSplit)+1 {
			tools.ErrorHandler(obj.Resp, obj.Req, 404)
			return
		}
		buf := "/"
		for i <= len(obj.URLSplit) {
			buf += obj.URLSplit[i-1]
			i++
		}
		http.ServeFile(obj.Resp, obj.Req, path.Join("js/", buf))

	case "minePageVar.css", "firstPage.css", "todoPageVar.css":
		http.ServeFile(obj.Resp, obj.Req, "server/"+obj.URLSplit[obj.SwitchLen])
	case "github", "git":
		github := "https://github.com/gidoBOSSftw5731"
		http.Redirect(obj.Resp, obj.Req, github, http.StatusSeeOther)
	case "signin", "login":
		tools.SignIn(obj.Resp, obj.Req, config)
	case "logout", "signout":
		sessions.DeleteKeySite(obj.Resp, obj.Req, config.SQLAcc)
		http.Redirect(obj.Resp, obj.Req, "/", 302)
	case "loginhandler":
		tools.LoginHandler(obj.Resp, obj.Req, config)
	case "verifysession":
		var user string
		ok, err := sessions.Verify(obj.Resp, obj.Req, config.SQLAcc, &user)
		if err != nil && err != fmt.Errorf("INVALID") {
			log.Errorln(err)
			tools.ErrorHandler(obj.Resp, obj.Req, 500)
			return
		}
		fmt.Fprintln(obj.Resp, ok)
	case "directory":
		tools.Directory(resp, req, config)
	case "":
		//raven.RecoveryHandler(appPage(obj.Resp, obj.Req, config))
		tools.AppPage(obj.Resp, obj.Req, config)
	default:
		tools.ErrorHandler(obj.Resp, obj.Req, 404)
	}

}
