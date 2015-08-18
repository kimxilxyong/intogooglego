package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Markdownifier using the wonderfull online api service at http://heckyesmarkdown.com/
// This example converts html provided as a string to markdown, the html does not need to be complete,
// just a part/snippet is enough
// Thanks to Brett for providing this service to the world!
func HtmlToMarkdown(htmlInput string) (markdownResult string, err error) {

	serviceEndPoint := "http://heckyesmarkdown.com/go/"
	postParams := url.Values{}
	postParams.Set("html", htmlInput) // the html input string
	postParams.Set("read", "0")       // turn readability off, default is on
	postParams.Set("md", "1")         // Run Markdownify, default on
	client := &http.Client{}
	resp, err := client.PostForm(serviceEndPoint, postParams)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	markdownResult = string(body)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("%s", resp.Status)
	}
	return
}

func main() {
	// The html for this demo is taken from http://dillinger.io/
	htmlForMarkdown := `
<h1><a id="Dillinger_0"></a>Dillinger</h1>
<p>Dillinger is a cloud-enabled, mobile-ready, offline-storage, AngularJS powered HTML5 Markdown editor.</p>
<ul>
<li>Type some Markdown on the left</li>
<li>See HTML in the right</li>
<li>Magic</li>
</ul>
<p>Markdown is a lightweight markup language based on the formatting conventions that people naturally use in email.  As [John Gruber] writes on the [Markdown site] [1]:</p>
<blockquote>
<p>The overriding design goal for Markdown’s
formatting syntax is to make it as readable
as possible. The idea is that a
Markdown-formatted document should be
publishable as-is, as plain text, without
looking like it’s been marked up with tags
or formatting instructions.</p>
</blockquote>
<p>This text you see here is <em>actually</em> written in Markdown! To get a feel for Markdown’s syntax, type some text into the left window and watch the results in the right.</p>
<h3><a id="Version_20"></a>Version</h3>
<p>3.0.2</p>
<h3><a id="Tech_23"></a>Tech</h3>
<p>Dillinger uses a number of open source projects to work properly:</p>
<ul>
<li>[AngularJS] - HTML enhanced for web apps!</li>
<li>[Ace Editor] - awesome web-based text editor</li>
<li>[Marked] - a super fast port of Markdown to JavaScript</li>
<li>[Twitter Bootstrap] - great UI boilerplate for modern web apps</li>
<li>[node.js] - evented I/O for the backend</li>
<li>[Express] - fast node.js network app framework [@tjholowaychuk]</li>
<li>[Gulp] - the streaming build system</li>
<li>[keymaster.js] - awesome keyboard handler lib by [@thomasfuchs]</li>
<li>[jQuery] - duh</li>
</ul>
<p>And of course Dillinger itself is open source with a <a href="https://github.com/joemccann/dillinger">public repository</a> on GitHub.</p>
<h3><a id="Installation_39"></a>Installation</h3>
<p>You need Gulp installed globally:</p>
<pre><code class="language-sh">$ npm i -g gulp
</code></pre>
<pre><code class="language-sh">$ git clone [git-repo-url] dillinger
$ <span class="hljs-built_in">cd</span> dillinger
$ npm i <span class="hljs-operator">-d</span>
$ mkdir -p public/files/{md,html,pdf}
$ gulp build --prod
$ NODE_ENV=production node app
</code></pre>
<h3><a id="Plugins_56"></a>Plugins</h3>
<p>Dillinger is currently extended with the following plugins</p>
<ul>
<li>Dropbox</li>
<li>Github</li>
<li>Google Drive</li>
<li>OneDrive</li>
</ul>
<p>Readmes, how to use them in your own application can be found here:</p>
<ul>
<li><a href="https://github.com/joemccann/dillinger/tree/master/plugins/dropbox/README.md">plugins/dropbox/README.md</a></li>
<li><a href="https://github.com/joemccann/dillinger/tree/master/plugins/github/README.md">plugins/github/README.md</a></li>
<li><a href="https://github.com/joemccann/dillinger/tree/master/plugins/googledrive/README.md">plugins/googledrive/README.md</a></li>
<li><a href="https://github.com/joemccann/dillinger/tree/master/plugins/onedrive/README.md">plugins/onedrive/README.md</a></li>
</ul>
<h3><a id="Development_72"></a>Development</h3>
<p>Want to contribute? Great!</p>
<p>Dillinger uses Gulp + Webpack for fast developing.
Make a change in your file and instantanously see your updates!</p>
<p>Open your favorite Terminal and run these commands.</p>
<p>First Tab:</p>
<pre><code class="language-sh">$ node app
</code></pre>
<p>Second Tab:</p>
<pre><code class="language-sh">$ gulp watch
</code></pre>
<p>(optional) Third:</p>
<pre><code class="language-sh">$ karma start
</code></pre>
<h3><a id="Todos_96"></a>Todos</h3>
<ul>
<li>Write Tests</li>
<li>Rethink Github Save</li>
<li>Add Code Comments</li>
<li>Add Night Mode</li>
</ul>
<h2><a id="License_103"></a>License</h2>
<p>MIT</p>
<p><strong>Free Software, Hell Yeah!</strong></p>
<ul>
<li><a href="http://daringfireball.net">john gruber</a></li>
<li><a href="http://twitter.com/thomasfuchs">@thomasfuchs</a></li>
<li><a href="http://daringfireball.net/projects/markdown/">1</a></li>
<li><a href="https://github.com/chjj/marked">marked</a></li>
<li><a href="http://ace.ajax.org">Ace Editor</a></li>
<li><a href="http://nodejs.org">node.js</a></li>
<li><a href="http://twitter.github.com/bootstrap/">Twitter Bootstrap</a></li>
<li><a href="https://github.com/madrobby/keymaster">keymaster.js</a></li>
<li><a href="http://jquery.com">jQuery</a></li>
<li><a href="http://twitter.com/tjholowaychuk">@tjholowaychuk</a></li>
<li><a href="http://expressjs.com">express</a></li>
<li><a href="http://angularjs.org">AngularJS</a></li>
<li><a href="http://gulpjs.com">Gulp</a></li>
</ul>
`

	mdResult, err := HtmlToMarkdown(htmlForMarkdown)
	if err != nil {
		fmt.Println(err)
		if mdResult != "" {
			fmt.Printf("Error Message: %s\n", mdResult)
		}
	} else {
		fmt.Printf("MARKDOWN RESULT --snip\n%s\nsnip-- MARKDOWN END\n", mdResult)
	}
}
