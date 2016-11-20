package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
	"strings"

	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	healthz "github.com/MEDIGO/go-healthz"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/lib/pq"
)

var usage = `
    PG Bridge: Send Postgres notifications to SNS or to a webhook.

    Usage:

      pg-bridge -c config.json

    Options:

      -c FILE, --conf FILE    Configuration file for PG Bridge
      -h, --help              Show this screen
      -v, --version           Get the version
`

// Health structure
type Health struct {
	Port int    `json:"port"`
	Path string `json:"path"`
}

// Postgres connection structure
type Postgres struct {
	URL string `json:"url"`
}

// Config struct for the JSON config file
type Config struct {
	Postgres `json:"postgres"`
	Routes   []string
	Health   `json:"health"`
}

var config string

func init() {
	flag.StringVar(&config, "c", "", "configuration file")
	flag.StringVar(&config, "conf", "", "configuration file")
}

func main() {
	log.SetHandler(text.New(os.Stderr))
	flag.Parse()

	if config == "" {
		println(usage)
		os.Exit(1)
	}

	conf, err := ioutil.ReadFile(config)
	if err != nil {
		log.WithError(err).Fatal("could not read config")
	}

	var mapping Config
	err = json.Unmarshal(conf, &mapping)
	if err != nil {
		log.WithError(err).Fatal("could not decode JSON")
	}

	routes := map[string][]string{}
	for _, v := range mapping.Routes {
		route := strings.Split(v, " ")
		if routes[route[0]] != nil {
			routes[route[0]] = append(routes[route[0]], route[1])
		} else {
			routes[route[0]] = []string{}
			routes[route[0]] = append(routes[route[0]], route[1])
		}
	}

	// setup Postgres
	pg := ConnectPostgres(mapping.Postgres, routes)
	defer pg.Close()

	// Setup SNS
	// @TODO figure out how to check that the required
	// variables are actually present in the session
	pub := sns.New(session.New())

	// setup a healthcheck on /health
	if mapping.Health.Port != 0 {
		go HTTP(mapping.Health, pg)
	}

	// route the notifications
	for {
		n := <-pg.Notify
		log.WithField("payload", n.Extra).Infof("notification from %s", n.Channel)

		// fetch the associated topic
		topics := routes[n.Channel]
		for _, topic := range topics {
			// publish in a separate goroutine
			if strings.HasPrefix(topic, "http") {
				go publishHTTP(n.Channel, topic, n.Extra)
			} else {
				go publishSNS(pub, n.Channel, topic, n.Extra)
			}
		}
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	// Send a response back to person contacting us.
	conn.Write([]byte("Message received."))
	// Close the connection when you're done with it.
	conn.Close()
}

// publish payload to SNS
func publishSNS(pub *sns.SNS, channel string, topic string, payload string) {
	SNSPayload := &sns.PublishInput{
		Message:  aws.String(payload),
		TopicArn: aws.String(topic),
	}

	_, err := pub.Publish(SNSPayload)
	if err != nil {
		log.WithError(err).WithField("channel", channel).WithField("payload", payload).Error("unable to send payload to SNS")
	}

	log.Infof("delivered notification from %s to SNS", channel)
	return
}

func publishHTTP(channel string, topic string, payload string) {
	body := []byte(payload)
	req, err := http.NewRequest("POST", topic, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("unable to POST")
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("cannot read body")
	}

	log.Infof("delivered notification from %s to %s with this response: %s", channel, topic, resp.Status)
}

// ConnectPostgres connect to postgres
func ConnectPostgres(postgres Postgres, routes map[string][]string) *pq.Listener {
	conninfo := postgres.URL
	if conninfo == "" {
		log.Fatal("postgres.url value required in the configuration")
	}

	log.Infof("connecting to postgres: %s...", conninfo)
	_, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.WithError(err).Fatal("could not connect to postgres")
	}
	log.Infof("connected to postgres")

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.WithError(err).Fatal("error listening for notifications")
		}
	}

	log.Infof("setting up a listener...")
	listener := pq.NewListener(conninfo, 10*time.Second, time.Minute, reportProblem)
	log.Infof("set up a listener")

	// listen on each channel
	for channel := range routes {
		log.Infof("listening on '%s'", channel)
		err := listener.Listen(channel)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	return listener
}

// HTTP Health simple healthcheck service
func HTTP(health Health, pg *pq.Listener) *http.ServeMux {

	healthz.Register("postgres", time.Second*5, func() error {
		return pg.Ping()
	})

	mux := http.NewServeMux()

	path := os.Getenv("HEALTH_PATH")
	if path == "" {
		path = "/health"
	}

	mux.Handle(path, healthz.Handler())
	http.ListenAndServe(":"+os.Getenv("HEALTH_PORT"), mux)
	return mux
}
