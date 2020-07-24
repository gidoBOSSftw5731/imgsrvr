package selector

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"../sessions"
	moocowtools "./modules/ProjectMoocow/tools"
	moocow "./modules/ProjectMoocow/web"
	tb "./modules/ProjectTapebucket"
	"./modules/dwarfism-2.0/dwarfism"
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

var (
	empty = ""
)

//SwitchStatement is a function holding a switch statement. it is in its own file so that
// webmasters may edit it to fit their configuration
func SwitchStatement(config tools.Config, obj Caseable) {
	switch obj.URLSplit[obj.SwitchLen] {
	case "i", "I":
		// Checks for hash/element/thing
		if len(obj.URLSplit) != obj.I1+1 {
			tools.ErrorHandler(obj.Resp, obj.Req, 400, "No file, No result")
			return
		}
		if obj.URLSplit[obj.I1] == "" {
			tools.ErrorHandler(obj.Resp, obj.Req, http.StatusNotFound, "Try specifying something!")
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
			tools.ErrorHandler(obj.Resp, obj.Req, 404, "I hate it too, dont worry")
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
	case "verifycaptcha":
		if len(obj.URLSplit) != obj.I1+1 {
			tools.ErrorHandler(obj.Resp, obj.Req, 429, "recaptcha failed")
			return
		}

		valid, err := tools.CheckCaptcha(obj.URLSplit[obj.I1], config.RecaptchaPrivKey)
		if err != nil || !valid {
			log.Errorln("Captcha error: ", err)
			tools.ErrorHandler(obj.Resp, obj.Req, 500, "captcha error idk")
			return
		}

		fmt.Fprint(obj.Resp, valid)
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
			tools.ErrorHandler(obj.Resp, obj.Req, 401, "quien es?")
			return
		}
		fmt.Fprintln(obj.Resp, ok)
	case "directory":
		tools.Directory(obj.Resp, obj.Req, config)
	case "pinbot":
		wd, _ := os.Getwd()

		if strings.Contains(obj.Req.URL.Path, "css") {
			http.ServeFile(obj.Resp, obj.Req, path.Join(wd, "server/selector/modules/ProjectMoocow/web/templates/main.css"))
			return
		} else if obj.URLECount < 4 {
			tools.ErrorHandler(obj.Resp, obj.Req, 404, "Malformed URL")
			return
		}

		var moocowconfig moocowtools.Config
		moocowtools.Configor(&moocowconfig, "server/selector/modules/ProjectMoocow/config.yml")

		discord, err := moocowtools.DiscordSession(moocowconfig)
		if err != nil {
			log.Errorln(err)
			tools.ErrorHandler(obj.Resp, obj.Req, 500, "Check Discord's status or try again")
			return
		}

		sql := moocow.SQLInfo{moocowconfig.DB.User,
			moocowconfig.DB.Password,
			moocowconfig.DB.IP,
			moocowconfig.DB.Port}

		log.Debugf("Guild: %v Channel: %v", obj.URLSplit[2], obj.URLSplit[3])
		webpage, err := moocow.Webpage(obj.URLSplit[2], obj.URLSplit[3], discord, sql,
			path.Join(wd, "server/selector/modules/ProjectMoocow/web/templates"))
		if err != nil {
			log.Errorln(err)
			tools.ErrorHandler(obj.Resp, obj.Req, 500, "idk what you want, have you tried turning it off and on again?")
			return
		}

		tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
		if err != nil {
			log.Errorln(err)
			tools.ErrorHandler(obj.Resp, obj.Req, 500, "This probably wasnt your fault, for once")
			return
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write([]byte(webpage))
		if err != nil {
			log.Errorln(err)
			tools.ErrorHandler(obj.Resp, obj.Req, 500, "please spare me from your wrath")
			return
		}

		http.ServeFile(obj.Resp, obj.Req, tmpFile.Name())

		tmpFile.Close()
		discord.Close()

	case "dwarfism2.0":
		dwarfism.ShortPage(obj.Resp, obj.Req, config, "")
	case "dform":
		dwarfism.ShortResp(obj.Resp, obj.Req, config)
	case "d":
		// Checks for hash/element/thing
		if len(obj.URLSplit) != obj.I1+1 {
			http.Redirect(obj.Resp, obj.Req, "/dwarfism2.0/", http.StatusSeeOther)
			return
		}

		dwarfism.Biggify(obj.Resp, obj.Req, config, obj.URLSplit[obj.I1])
	case "submitpaste":
		if obj.Req.Method != "POST" {
			tools.ErrorHandler(obj.Resp, obj.Req, 400, "Must be post")
			return
		}

		// I realize now that I am not going to copy over this struct every time
		req := obj.Req
		resp := &obj.Resp

		db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(127.0.0.1:3306)/ImgSrvr", config.SQLAcc))
		if err != nil {
			tools.ErrorHandler(*resp, req, 500, "error adding to db")
			log.Error("Oh noez, could not connect to database")
			return
		}
		log.Debug("Oi, mysql did thing")
		defer db.Close()
		// end of SQL opening

		//paste := req.FormValue("paste") // Saving paste to memory
		type out struct {
			Paste string
		}
		var u out
		err = json.NewDecoder(req.Body).Decode(&u)
		if err != nil {
			tools.ErrorHandler(*resp, req, 500, "json error")
			log.Errorln(err)
			return
		}

		var user string
		sessions.Verify(*resp, req, config.SQLAcc, &user)

		if user == "" {
			user = "anon"
		}

		url, err := tb.AddToDB(&u.Paste, &user, db)
		if err != nil {
			tools.ErrorHandler(*resp, req, 500, "error adding to db")
			log.Errorln(err)
			return
		}
		fmt.Fprintf(*resp, "/p/%v", url)
		log.Debugf("added paste %v", url)
	case "p", "P":
		// I realize now that I am not going to copy over this struct every time
		req := obj.Req
		resp := &obj.Resp

		db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(127.0.0.1:3306)/ImgSrvr", config.SQLAcc))
		if err != nil {
			tools.ErrorHandler(*resp, req, 500, "error adding to db")
			log.Error("Oh noez, could not connect to database")
			return
		}
		log.Debug("Oi, mysql did thing")
		defer db.Close()
		// end of SQL opening

		templatePath := "./server/selector/modules/ProjectTapebucket/index.html"
		templateContents, _ := ioutil.ReadFile(templatePath)

		paste := &empty
		if obj.URLECount != 2 {
			// look for the paste
			pasteurl := obj.URLSplit[2]
			if pasteurl == "" {
				break
			}
			paste, err = tb.ReturnFromDB(pasteurl, db)
			if err != nil {
				tools.ErrorHandler(*resp, req, 500, "error getting from db, is the URL correct?")
				log.Error(err)
				return
			}
		}

		type tmpl struct {
			Paste string
		}
		t := template.New("paste")
		t, _ = t.Parse(string(templateContents))
		if err := t.Execute(*resp, tmpl{*paste}); err != nil {
			tools.ErrorHandler(*resp, req, 500, "templating error")
			log.Error(err)
			return
		}

	case "":
		//raven.RecoveryHandler(appPage(obj.Resp, obj.Req, config))
		tools.AppPage(obj.Resp, obj.Req, config)
	default:
		tools.ErrorHandler(obj.Resp, obj.Req, 404, "idk what you want, I dont get paid enough to read minds")
	}

}
