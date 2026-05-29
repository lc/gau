# getallurls (gau)
[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)

getallurls (gau) fetches known URLs from AlienVault's [Open Threat Exchange](https://otx.alienvault.com), the Wayback Machine, Common Crawl, and URLScan for any given domain. Inspired by Tomnomnom's [waybackurls](https://github.com/tomnomnom/waybackurls).

# Resources
- [Usage](#usage)
- [Installation](#installation)
- [ohmyzsh note](#ohmyzsh-note)

## Usage:
Examples:

```bash
$ printf example.com | gau
$ cat domains.txt | gau --threads 5
$ gau example.com google.com
$ gau --o example-urls.txt example.com
$ gau --blacklist png,jpg,gif example.com
```

To display the help for the tool use the `-h` flag:

```bash
$ gau -h
```

| Flag | Description | Example |
|------|-------------|---------|
|`--blacklist`| list of extensions to skip | gau --blacklist ttf,woff,svg,png|
|`--config` | Use alternate configuration file (default `$HOME/config.toml` or `%USERPROFILE%\.gau.toml`) | gau --config $HOME/.config/gau.toml|
|`--fc`| list of status codes to filter | gau --fc 404,302 |
|`--from`| fetch urls from date (format: YYYYMM) | gau --from 202101 |
|`--ft`| list of mime-types to filter | gau --ft text/plain|
|`--fp`| remove different parameters of the same endpoint | gau --fp|
|`--json`| output as json | gau --json |
|`--mc`| list of status codes to match | gau --mc 200,500 |
|`--mt`| list of mime-types to match |gau --mt text/html,application/json|
|`--o`| filename to write results to | gau --o out.txt |
|`--providers`| list of providers to use (wayback,commoncrawl,otx,urlscan) | gau --providers wayback|
|`--proxy`| http proxy to use (socks5:// or http:// | gau --proxy http://proxy.example.com:8080 |
|`--retries`| retries for HTTP client | gau --retries 10 |
|`--timeout`| timeout (in seconds) for HTTP client | gau --timeout 60 |
|`--subs`| include subdomains of target domain | gau example.com --subs |
|`--threads`| number of workers to spawn | gau example.com --threads |
|`--to`| fetch urls to date (format: YYYYMM) | gau example.com --to 202101 |
|`--verbose`| show verbose output | gau --verbose example.com |
|`--version`| show gau version | gau --version|


## Configuration Files
gau automatically looks for a configuration file at `$HOME/.gau.toml` or`%USERPROFILE%\.gau.toml`. You can point to a different configuration file using the `--config` flag. **If the configuration file is not found, gau will still run with a default configuration, but will output a message to stderr**.

You can specify options and they will be used for every subsequent run of gau. Any options provided via command line flags will override options set in the configuration file.

An example configuration file can be found [here](https://github.com/lc/gau/blob/master/.gau.toml)

## Installation:
### From source:
```
$ go install github.com/lc/gau/v2/cmd/gau@latest
```
### From github :
```
git clone https://github.com/lc/gau.git; \
cd gau/cmd; \
go build; \
sudo mv gau /usr/local/bin/; \
gau --version;
```
### From binary:
You can download the pre-built binaries from the [releases](https://github.com/lc/gau/releases/) page and then move them into your $PATH.

```bash
$ tar xvf gau_2.0.6_linux_amd64.tar.gz
$ mv gau /usr/bin/gau
```

### From Docker:
You can run gau via docker like so:
```bash
docker run --rm sxcurity/gau:latest --help
```


You can also build a docker image with the following command
```bash
docker build -t gau .
```
and then run it
```bash
docker run gau example.com
```
Bear in mind that piping command (echo "example.com" | gau) will not work with the docker container


## ohmyzsh note:
ohmyzsh's [git plugin](https://github.com/ohmyzsh/ohmyzsh/tree/master/plugins/git) has an alias which maps `gau` to the `git add --update` command. This is problematic, causing a binary conflict between this tool "gau" and the zsh plugin alias "gau" (`git add --update`). There is currently a few workarounds which can be found in this Github [issue](https://github.com/lc/gau/issues/8). 


## Useful?

<a href="http://buymeacoff.ee/cdl" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>

<a href="https://commoncrawl.org/donate/">Donate to CommonCrawl</a><br>
<a href="https://archive.org/donate">Donate to the InternetArchive</a>
