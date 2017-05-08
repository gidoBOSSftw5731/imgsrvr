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

<form action="/app/test/" method="POST">
  First name: <input type="text" name="fn"><br>
  Last name: <input type="text" name="tn"><br>
  <input type="submit" value="Go!">
</form>

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
  And a testing graphic!!!<img src="./testingpics/Graphic1.jpg" alt="Testing Graphic" style="height:500;width:500">

  <p>
  field from request: {{ .Tn}}
  </p>

  <form action="/app/main/" method="POST">
    <input type="submit" value="back to main!">
  </form>

  </body> </html>`
)
