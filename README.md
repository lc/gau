# getallurls
Fetch known URLs from AlienVault's [Open Threat Exchange](https://otx.alienvault.com), the Wayback Machine, and Common Crawl. Originally built as a microservice.

### usage:
```
▻ printf 'example.com' | gau
```

or

```
▻ gau example.com
```

### install:

```
▻ git clone https://github.com/lc/gau && cd gau
▻ go build -o $GOPATH/bin/gau gau.go
```

or

```
▻ go get -u github.com/lc/gau
```

## Credits:
Thanks @tomnomom for [waybackurls](https://github.com/tomnomnom/waybackurls)!
