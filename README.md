# getallurls (gau)
[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![Go ReportCard](https://goreportcard.com/badge/github.com/lc/gau)](https://goreportcard.com/report/github.com/lc/gau)

getallurls (gau) fetches known URLs from AlienVault's [Open Threat Exchange](https://otx.alienvault.com), the Wayback Machine, and Common Crawl for any given domain. Inspired by Tomnomnom's [waybackurls](https://github.com/tomnomnom/waybackurls).

# Resources
- [Usage](#usage)
- [Installation](#installation)
- [ohmyzsh note](#ohmyzsh-note)

## Usage:
Examples:

```bash
$ printf example.com | gau
$ cat domains.txt | gau
$ gau example.com
$ gau -o example-urls.txt example.com
```

To display the help for the tool use the `-h` flag:

```bash
$ gau -h
```

| Flag | Description | Example |
|------|-------------|---------|
| `-providers` | providers to fetch urls from (by default, all are used) | `gau -providers wayback,otx,commoncrawl example.com` |
| `-retries` | amount of retries for http client | `gau -retries 7 example.com` |
| `-subs` | include subdomains of target domain | `gau -subs example.com` |
| `-p` | http proxy to use | `gau -p http://localhost:8080 example.com` |
| `-random-agent` | use a random user-agent | `gau -random-agent example.com` |
| `-v` | enable verbose mode (show errors) | `gau -v` |
| `-o` | filename to write results to | `gau -o urls.txt example.com` | 
| `-json` | write output as json | `gau -json example.com` |
| `-version` | show gau version | `gau -version` |



## Installation:
### From source:
```
$ GO111MODULE=on go get -u -v github.com/lc/gau
```

### From binary:
You can download the pre-built binaries from the [releases](https://github.com/lc/gau/releases/) page and then move them into your $PATH.

```bash
$ tar xvf gau_1.0.3_linux_amd64.tar.gz
$ mv gau /usr/bin/gau
```

## ohmyzsh note:
ohmyzsh's [git plugin](https://github.com/ohmyzsh/ohmyzsh/tree/master/plugins/git) has an alias which maps `gau` to the `git add --update` command. This is problematic, causing a binary conflict between this tool "gau" and the zsh plugin alias "gau" (`git add --update`). There is currently a few workarounds which can be found in this Github [issue](https://github.com/lc/gau/issues/8). 


## Useful?

<a href="http://buymeacoff.ee/cdl" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
