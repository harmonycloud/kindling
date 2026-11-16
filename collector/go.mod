module github.com/Kindling-project/kindling/collector

go 1.19

require (
	github.com/DataDog/ebpf v0.0.0-20220301203322-3fc9ab3b8daf
	github.com/florianl/go-conntrack v0.3.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/mdlayher/netlink v1.7.1
	github.com/olivere/elastic v6.2.37+incompatible // indirect
	github.com/olivere/elastic/v6 v6.2.1
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.17.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.8.1
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.25.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.7.0
	go.opentelemetry.io/otel/exporters/prometheus v0.25.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.25.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.2.0
	go.opentelemetry.io/otel/metric v0.30.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/sdk/export/metric v0.25.0
	go.opentelemetry.io/otel/sdk/metric v0.25.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.uber.org/atomic v1.10.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.23.0
	golang.org/x/sys v0.18.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	k8s.io/client-go v0.24.2
)

require (
	github.com/mitchellh/mapstructure v1.5.0
	golang.org/x/sync v0.3.0
)

require (
	connectrpc.com/connect v1.14.0
	github.com/google/pprof v0.0.0-20220608213341-c488b8fa1db3
	github.com/grafana/dskit v0.0.0-20240104111617-ea101a3b86eb
	github.com/grafana/pyroscope/api v0.4.0
	github.com/klauspost/compress v1.17.3
	github.com/prometheus/common v0.44.0
	github.com/prometheus/prometheus v0.37.9
	google.golang.org/protobuf v1.32.0
)

// Pin versions that prometheus/pyroscope would otherwise upgrade past the collector baseline.
replace (
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace => go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc => go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.2.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v0.25.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.2.0
	go.opentelemetry.io/proto/otlp => go.opentelemetry.io/proto/otlp v0.10.0
	k8s.io/api => k8s.io/api v0.21.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.5
	k8s.io/client-go => k8s.io/client-go v0.21.5
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mdlayher/socket v0.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.25.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.7.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.25.0 // indirect
	go.opentelemetry.io/proto/otlp v0.16.0 // indirect
	golang.org/x/oauth2 v0.11.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/grpc v1.59.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.70.0 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
