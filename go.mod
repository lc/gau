module github.com/lc/gau/v2

go 1.20

require (
	github.com/bobesa/go-domain-util v0.0.0-20190911083921-4033b5f7dd89
	github.com/deckarep/golang-set/v2 v2.3.0
	github.com/json-iterator/go v1.1.12
	github.com/lynxsecurity/pflag v1.1.3
	github.com/lynxsecurity/viper v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/valyala/bytebufferpool v1.0.0
	github.com/valyala/fasthttp v1.31.0
)

require (
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/ini.v1 v1.64.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

retract (
	v2.0.7
	v2.0.3
	v2.0.2
	v2.0.1
)
