# Pkg Critic

(GitHub) Pkg Critic is a tool combining search results from GoDoc API and stats info from GitHub API.

Pkg Critic also uses indentation to show the fork relationship between packages.

The GoDoc search result is re-ordered by GitHub Stars.

# Usage

Cli Interface:


```
pkgcritic -q browser -github-token xxxxx
```

Web Interface:

```
pkgcritic -web -open -github-token xxxxx
```

Note: using `-github-token` is to increase github api rate limit.
Details cound be found here:
https://developer.github.com/v3/oauth/
https://github.com/settings/tokens

![web](https://raw.githubusercontent.com/bom-d-van/pkgcritic/master/intro/web.png)

![cli](https://raw.githubusercontent.com/bom-d-van/pkgcritic/master/intro/cli.png)
