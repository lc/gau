# getallurls (gau)
Fetch known URLs from AlienVault's [Open Threat Exchange](https://otx.alienvault.com), the Wayback Machine, and Common Crawl. Inspired by @tomnomnom's [waybackurls](https://github.com/tomnomnom/waybackurls).

## usage:
```
▻ printf 'example.com' | gau
```

or

```
▻ gau example.com
```

## flags:
```
  -providers string
    	providers to fetch urls for (default "wayback,otx,commoncrawl")
  -retries uint
    	amount of retries for http client (default 5)
  -subs
    	include subdomains of target domain
```

## installation:
### from source:
```
▻ GO111MODULE=on go get -u -v github.com/lc/gau
```

### from binary:
You can download the pre-built binaries from the [releases](https://github.com/lc/gau/releases/) page and then move them into your $PATH.

```bash
▻ tar xvf gau-linux-amd64.tar
▻ mv gau-linux-amd64 /usr/bin/gau
```

## Useful?

<a href="http://buymeacoff.ee/cdl" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;-webkit-box-shadow: 0px 3px 2px 0px rgba(190, 190, 190, 0.5) !important;" ></a>
