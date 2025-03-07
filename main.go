package main

import (
	"flag"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen-address", ":9104", "Address to listen on for web interface and telemetry.")
	metricPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	use_cee       = flag.Bool("format.cee", false, "Use CEE format instead of JSON as input.")
)

func main() {
	logwriter, e := syslog.New(syslog.LOG_NOTICE|syslog.LOG_SYSLOG, "rsyslog_exporter")
	if e == nil {
		log.SetOutput(logwriter)
	}

	flag.Parse()
	exporter := newRsyslogExporter()
	exporter.use_cee = *use_cee

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Print("interrupt received, exiting")
		os.Exit(0)
	}()

	go func() {
		exporter.run()
	}()

	prometheus.MustRegister(exporter)
	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>Rsyslog exporter</title></head>
<body>
<h1>Rsyslog exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`))
	})

	log.Printf("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
