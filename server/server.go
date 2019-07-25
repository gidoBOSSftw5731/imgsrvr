package server

import (
	"net/http"
	"strings"

	//raven "github.com/getsentry/raven-go"<<<<<<< staging
	"./selector"
	"./selector/tools"

	"github.com/gidoBOSSftw5731/log"
)

//FastCGIServer is how the config constants get to the server package.
type FastCGIServer struct {
	config tools.Config
}

//NewFastCGIServer is an implementation of fastcgi server.
func NewFastCGIServer(URLPrefix, imgStore, baseURL, SQLAcc, recaptchaPrivKey, recaptchaPubKey string, imgHash int) *FastCGIServer {
	return &FastCGIServer{
		config: tools.Config{
			URLPrefix:        URLPrefix,
			ImgHash:          imgHash,
			ImgStore:         imgStore,
			BaseURL:          baseURL,
			SQLAcc:           SQLAcc,
			RecaptchaPrivKey: recaptchaPrivKey,
			RecaptchaPubKey:  recaptchaPubKey,
		}}
}

/*HTML meaning guide for my sanity:
<br>: page break
<html> and </html>: beginning and end of html section
<body> and </body>: main part with text
<!-- and -->: Comments... WHY
<a href="url"> and </a>: links
<p> and </p>: margin
*/

/*func prepareTemplate(source string) (*template.New, string, error) {

}*/

func (s FastCGIServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	log.Debug("the req arrived")
	if req.Body == nil {
		return
	}
	var caseobj selector.Caseable
	ip := req.RemoteAddr
	log.Debug("This request is being requested by:", ip)

	caseobj.URLSplit = strings.Split(req.URL.Path, "/")
	caseobj.URLECount = len(caseobj.URLSplit)
	log.Debug("The url is:", req.URL.Path)
	log.Debugf("urlECount: %d\n", caseobj.URLECount)
	// Checking amt of elements in url (else sends 404)
	//Check for prefix
	if caseobj.URLECount < 2 || !strings.HasPrefix(req.URL.Path, s.config.URLPrefix) {
		tools.ErrorHandler(resp, req, http.StatusNotFound)
		return
	}
	// Defining variables as if there was NO prefix (this will never happen)
	// INSERT FUCTION TO DO VARS

	/* URL Template
	0: usually null, or maybe its the url...?
	1: The prefix, if there is one, if it isnt then this list isnt applicable
	2: The part to ask for a part of the site
	3: data for the site (like file name)
	*/
	// Add your variables involving the URL here
	caseobj.NumberOfPrefixSlashes = strings.Count(s.config.URLPrefix, "/") - 1
	caseobj.SwitchLen = 1 + caseobj.NumberOfPrefixSlashes
	//test1 := 2 + numberOfPrefixSlashes
	//test2 := 1 + numberOfPrefixSlashes
	caseobj.I1 = 2 + caseobj.NumberOfPrefixSlashes

	caseobj.Resp = resp
	caseobj.Req = req

	selector.SwitchStatement(s.config, caseobj)

}
