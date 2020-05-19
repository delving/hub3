module github.com/delving/hub3

go 1.14

require (
	code.gitea.io/gitea v1.11.5
	github.com/DataDog/zstd v1.3.5 // indirect
	github.com/OneOfOne/xxhash v1.2.2
	github.com/Sereal/Sereal v0.0.0-20190226181601-237c2cca198f // indirect
	github.com/allegro/bigcache v1.1.0
	github.com/antzucaro/matchr v0.0.0-20191224151129-ab6ba461ddec
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/asdine/storm v1.1.0
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/deiu/gon3 v0.0.0-20170627184619-f84eb1e0bd62
	github.com/die-net/lrucache v0.0.0-20190707192454-883874fe3947
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/elastic/go-elasticsearch/v8 v8.0.0-20200428082342-063ba81e9d1b
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/elazarl/goproxy v0.0.0-20181111060418-2ce16c963a8a // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/gammazero/workerpool v0.0.0-20180103203609-079e51c30502
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/cors v1.0.0
	github.com/go-chi/docgen v1.0.2
	github.com/go-chi/render v0.0.0-20171231234154-8c8c7a43d054
	github.com/go-git/go-git/v5 v5.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.4.0
	github.com/google/gofuzz v1.0.0
	github.com/gorilla/schema v1.0.2
	github.com/gosimple/slug v1.9.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/jinzhu/configor v1.2.0 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/justinas/alice v1.2.0
	github.com/kiivihal/goharvest v0.0.0-20190502201718-d93ace331ed0
	github.com/kiivihal/rdf2go v0.1.2
	github.com/klauspost/compress v1.10.5 // indirect
	github.com/knakk/digest v0.0.0-20160404164910-fd45becddc49 // indirect
	github.com/knakk/rdf v0.0.0-20171130200148-b6ee24f8f40f
	github.com/knakk/sparql v0.0.0-20170625101756-3de19ad6a5dc
	github.com/labstack/gommon v0.0.0-20170925052817-57409ada9da0
	github.com/linkeddata/gojsonld v0.0.0-20170418210642-4f5db6791326
	github.com/mailgun/groupcache v1.3.0
	github.com/matryer/is v1.3.0
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moul/http2curl v0.0.0-20170919181001-9ac6cf4d929b // indirect
	github.com/muesli/smartcrop v0.3.0 // indirect
	github.com/nats-io/nats-server/v2 v2.1.6 // indirect
	github.com/nats-io/nats-streaming-server v0.17.0 // indirect
	github.com/nats-io/stan.go v0.6.0
	github.com/olivere/elastic/v7 v7.0.10
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/parnurzeal/gorequest v0.2.15
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/common v0.10.0 // indirect
	github.com/qor/admin v0.0.0-20200315024928-877b98a68a6f // indirect
	github.com/qor/assetfs v0.0.0-20170713023933-ff57fdc13a14 // indirect
	github.com/qor/audited v0.0.0-20171228121055-b52c9c2f0571 // indirect
	github.com/qor/media v0.0.0-20191022071353-19cf289e17d4 // indirect
	github.com/qor/middlewares v0.0.0-20170822143614-781378b69454 // indirect
	github.com/qor/oss v0.0.0-20191031055114-aef9ba66bf76 // indirect
	github.com/qor/qor v0.0.0-20200224122013-457d2e3f50e1 // indirect
	github.com/qor/responder v0.0.0-20171031032654-b6def473574f // indirect
	github.com/qor/roles v0.0.0-20171127035124-d6375609fe3e // indirect
	github.com/qor/serializable_meta v0.0.0-20180510060738-5fd8542db417 // indirect
	github.com/qor/session v0.0.0-20170907035918-8206b0adab70 // indirect
	github.com/qor/transition v0.0.0-20190608002025-f17b56902e4b
	github.com/qor/validations v0.0.0-20171228122639-f364bca61b46 // indirect
	github.com/rs/xid v1.2.1
	github.com/rs/zerolog v1.18.0
	github.com/rychipman/easylex v0.0.0-20160129204217-49ee7767142f // indirect
	github.com/sajari/fuzzy v1.0.0
	github.com/segmentio/ksuid v1.0.2
	github.com/sosedoff/gitkit v0.2.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/src-d/go-git v4.7.0+incompatible
	github.com/stretchr/testify v1.4.0
	github.com/testcontainers/testcontainers-go v0.3.1
	github.com/theplant/cldr v0.0.0-20190423050709-9f76f7ce4ee8 // indirect
	github.com/theplant/htmltestingutils v0.0.0-20190423050759-0e06de7b6967 // indirect
	github.com/theplant/testingutils v0.0.0-20190603093022-26d8b4d95c61 // indirect
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/urfave/negroni v0.3.0
	github.com/valyala/fasthttp v1.12.0
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4 // indirect
	github.com/yosssi/gohtml v0.0.0-20200512035251-dd92a3e0d30d // indirect
	go.elastic.co/apm/module/apmchi v1.5.0
	go.elastic.co/apm/module/apmhttp v1.7.2 // indirect
	golang.org/x/crypto v0.0.0-20200420201142-3c4aac89819a // indirect
	golang.org/x/image v0.0.0-20200430140353-33d19683fad8 // indirect
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200515095857-1151b9dac4a9 // indirect
	golang.org/x/text v0.3.2
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
	google.golang.org/grpc v1.21.1
	google.golang.org/protobuf v1.23.0
	gopkg.in/VividCortex/ewma.v1 v1.1.1 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.0-20171129124420-c112833d014c
	gopkg.in/cheggaaa/pb.v2 v2.0.7 // indirect
	gopkg.in/fatih/color.v1 v1.7.0 // indirect
	gopkg.in/mattn/go-colorable.v0 v0.1.0 // indirect
	gopkg.in/mattn/go-isatty.v0 v0.0.4 // indirect
	gopkg.in/mattn/go-runewidth.v0 v0.0.4 // indirect
	gopkg.in/olivere/elastic.v5 v5.0.79
	gopkg.in/vmihailenco/msgpack.v2 v2.9.1 // indirect
	willnorris.com/go/imageproxy v0.10.0
)
