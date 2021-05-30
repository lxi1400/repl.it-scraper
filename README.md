# Only working repl.it scraper (for discord tokens) as of 28/02/2021 (that doesn't just return the URL)


# Note to repl.it employees if they see this repository (doubt it)
Hi! 

# General useless nerdy information
```YAML
What it does (steps):
1) Scrapes repl.it URLS off bing
2) Downloads zip file of repl.it repository to a local folder
3) extracts files from zip
4) reads file data for a token
5) dumps tokens to a file
6) opens up the notepad containing all of the tokens
7) cleans up junk files
```

Code is ugly I know, but it works.

Not gonna be extremely fast. Deal with it.

I was too lazy to partition it into multiple files.


# How to build

```YAML
How to build:
go build -ldflags "-s -w"

Don't want to build it for whatever reason? (here's a way to run it through source):
go run main.go
```
