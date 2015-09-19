package main

import "text/template"

var tmpl = template.Must(template.New("").Parse(`

{{define "main"}}
<!DOCTYPE html>
<html>
<head>
	<title>Pkg Critic</title>
	<style type="text/css">
		body {
			line-height: 20px;
			font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
			font-size: 14px;
			margin: 10px auto;
			width: 50%;
		}
		span {
			border-color: black;
			border-style: solid;
			border-width: 1px;
			border-radius: 5px;
			padding: 0 5px;
		}
		a {
			color: #375eab;
			text-decoration: none;
		}
		a:hover {
			color: #23527c;
			text-decoration: underline;
		}
		h1 {
			box-sizing: border-box;
			color: rgb(55, 94, 171);
			display: inline;
			font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
			font-weight: bold;
		    display: inline-block;
		}
		ul {
			padding: 0px;
			list-style: none;
		}
		li {
			border-top: 1px solid #DDD;
		}
		form input {
			width: 80%;
		}
		form button {
			width: 18%;
		}
	</style>
</head>
<body>
	<h1>Pkg Critic</h1>
	<div>
		<form action="/">
			<input name="query" value="{{.Query}}">
			<button>Search</button>
		</form>
	</div>
	{{if .GitHubs}}
		<h2>GitHub Packages ({{len .GitHubs}})</h2>
		{{template "critiques" .GitHubs}}
	{{end}}
	{{if .NonGitHubs}}
		<h2>Non-GitHub Packages ({{len .NonGitHubs}})</h2>
		{{template "critiques" .NonGitHubs}}
	{{end}}
</body>
</html>
{{end}}

{{define "critiques"}}
	<ul>
		{{range .}}
			<li>
				<div class="left-panel">
					<p><a href="https://godoc.org/{{.Path}}">{{.Path}}</a></p>
				</div>

				<div class="right-panel clear">
					{{if .Repository}}
						<p>
							<span><a href="https://github.com/{{.GitHubFullName}}">Repo</a></span>
							<span>Stars: {{.StargazersCount}}</span>
							<span>Forks: {{.ForksCount}}</span>
							<span>Updated At: {{.UpdatedAt.Format "2006-01-02"}}</span>
							<span>Created At: {{.CreatedAt.Format "2006-01-02"}}</span>
						</p>
					{{end}}
					<p>{{.Synopsis}}</p>

					<div style="padding-left:50px;">
						{{if .Forks}}{{template "critiques" .Forks}}{{end}}
					</div>
				</div>
			</li>
		{{end}}
	</ul>
{{end}}
`))
