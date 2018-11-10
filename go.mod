module github.com/gomods/athens

require (
	cloud.google.com/go v0.26.0
	contrib.go.opencensus.io/exporter/stackdriver v0.6.0
	github.com/Azure/azure-pipeline-go v0.0.0-20180607212504-7571e8eb0876
	github.com/Azure/azure-storage-blob-go v0.0.0-20180727221336-197d1c0aea1b
	github.com/BurntSushi/toml v0.3.1
	github.com/DataDog/datadog-go v0.0.0-20180822151419-281ae9f2d895
	github.com/DataDog/opencensus-go-exporter-datadog v0.0.0-20180917103902-e6c7f767dc57
	github.com/aws/aws-sdk-go v1.15.24
	github.com/cockroachdb/cockroach-go v0.0.0-20181001143604-e0a95dfd547c
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/globalsign/mgo v0.0.0-20180828104044-6f9f54af1356
	github.com/go-ini/ini v1.25.4
	github.com/go-playground/locales v0.12.1
	github.com/go-playground/universal-translator v0.16.0
	github.com/go-sql-driver/mysql v1.4.0
	github.com/gobuffalo/buffalo v0.13.1
	github.com/gobuffalo/buffalo-plugins v1.6.1
	github.com/gobuffalo/envy v1.6.5
	github.com/gobuffalo/events v1.1.1
	github.com/gobuffalo/fizz v1.0.12
	github.com/gobuffalo/flect v0.0.0-20181019110701-3d6f0b585514
	github.com/gobuffalo/genny v0.0.0-20181019144442-df0a36fdd146
	github.com/gobuffalo/github_flavored_markdown v1.0.5
	github.com/gobuffalo/httptest v1.0.2
	github.com/gobuffalo/makr v1.1.5
	github.com/gobuffalo/meta v0.0.0-20181018192820-8c6cef77dab3
	github.com/gobuffalo/mw-forcessl v0.0.0-20180802152810-73921ae7a130
	github.com/gobuffalo/mw-paramlogger v0.0.0-20181005191442-d6ee392ec72e
	github.com/gobuffalo/packr v1.13.7
	github.com/gobuffalo/plush v3.7.20+incompatible
	github.com/gobuffalo/pop v4.8.4+incompatible
	github.com/gobuffalo/suite v2.1.6+incompatible
	github.com/gobuffalo/tags v2.0.11+incompatible
	github.com/gobuffalo/uuid v2.0.4+incompatible
	github.com/gobuffalo/validate v2.0.3+incompatible
	github.com/gobuffalo/x v0.0.0-20181007152206-913e47c59ca7
	github.com/gofrs/uuid v3.1.0+incompatible
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/protobuf v1.2.0
	github.com/google/go-cmp v0.2.0
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/googleapis/gax-go v2.0.0+incompatible
	github.com/gopherjs/gopherjs v0.0.0-20180825215210-0210a2f0f73c // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/sessions v1.1.3
	github.com/hashicorp/go-multierror v1.0.0
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/jackc/pgx v3.2.0+incompatible
	github.com/jmespath/go-jmespath v0.0.0-20160202185014-0b12d6b521d8
	github.com/jmoiron/sqlx v1.2.0
	github.com/jtolds/gls v4.2.1+incompatible // indirect
	github.com/karrick/godirwalk v1.7.5
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/kr/pty v1.1.3 // indirect
	github.com/lib/pq v1.0.0
	github.com/markbates/going v1.0.2
	github.com/markbates/goth v1.46.0
	github.com/markbates/grift v1.0.4
	github.com/markbates/inflect v1.0.1
	github.com/markbates/oncer v0.0.0-20181014194634-05fccaae8fc4
	github.com/markbates/refresh v1.4.10
	github.com/markbates/willie v1.0.9
	github.com/mattn/go-colorable v0.0.9
	github.com/mattn/go-isatty v0.0.4
	github.com/mattn/go-sqlite3 v1.9.0
	github.com/minio/minio-go v6.0.5+incompatible
	github.com/mitchellh/go-homedir v1.0.0
	github.com/philhofer/fwd v1.0.0
	github.com/pkg/errors v0.8.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/rs/cors v1.5.0
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516
	github.com/sergi/go-diff v1.0.0
	github.com/sirupsen/logrus v1.1.1
	github.com/smartystreets/assertions v0.0.0-20180820201707-7c9eb446e3cf // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.2.2
	github.com/tinylib/msgp v1.0.2
	github.com/unrolled/secure v0.0.0-20181005190816-ff9db2ff917f
	go.opencensus.io v0.17.0
	golang.org/x/crypto v0.0.0-20181015023909-0c41d7ab0a0e
	golang.org/x/net v0.0.0-20181017193950-04a2e542c03f
	golang.org/x/oauth2 v0.0.0-20180620175406-ef147856a6dd
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20181019084534-8f1d3d21f81b
	golang.org/x/text v0.3.0
	google.golang.org/api v0.0.0-20180910000450-7ca32eb868bf
	google.golang.org/appengine v1.2.0
	google.golang.org/genproto v0.0.0-20180831171423-11092d34479b
	google.golang.org/grpc v1.14.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.3.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.20.2
	gopkg.in/yaml.v2 v2.2.1
)
