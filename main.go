package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	healthz "github.com/MEDIGO/go-healthz"
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/lib/pq"
)

func main() {
	log.SetHandler(text.New(os.Stderr))

	routesEnv := os.Getenv("BRIDGE_ROUTES")
	if routesEnv == "" {
		log.Fatal("BRIDGE_ROUTES environment variable required to handle the routing")
	}

	routing := strings.Split(routesEnv, ";")
	routes := make(map[string]string)
	for _, route := range routing {
		r := strings.Split(route, ",")
		if len(r) == 2 {
			routes[r[0]] = r[1]
		}
	}

	// setup Postgres
	pg := Postgres(routes)
	defer pg.Close()

	// setup SNS
	pub := sns.New(session.New())

	// setup a healthcheck on /health
	go Health(pg)

	keys := make([]string, 0, len(routes))
	for k := range routes {
		keys = append(keys, k)
	}

	log.Infof("listening for notifications on %s", strings.Join(keys, ", "))

	// start waiting for notifications to come in
	Notifications(pg, pub, routes)
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

// Notifications processes
func Notifications(pg *pq.Listener, pub *sns.SNS, routes map[string]string) {
	for {
		n := <-pg.Notify
		log.WithField("payload", n.Extra).Infof("notification from %s", n.Channel)

		// fetch the associated topic
		topic := routes[n.Channel]
		if topic == "" {
			log.WithField("channel", n.Channel).Error("no sns topic for channel")
		}

		payload := &sns.PublishInput{
			Message:  aws.String(n.Extra),
			TopicArn: aws.String(topic),
		}

		// publish in a separate goroutine
		go publish(pub, payload, n)
	}
}

// publish payload to SNS
func publish(pub *sns.SNS, payload *sns.PublishInput, n *pq.Notification) {
	_, err := pub.Publish(payload)
	if err != nil {
		log.WithError(err).WithField("channel", n.Channel).WithField("payload", n.Extra).Error("unable to send payload to SNS")
	}

	log.Infof("delivered notification from %s to SNS", n.Channel)
	return
}

// Postgres connect to postgres
func Postgres(routes map[string]string) *pq.Listener {
	conninfo := os.Getenv("POSTGRES_URL")
	if conninfo == "" {
		log.Fatal("POSTGRES_URL environment variable required")
	}

	_, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.WithError(err).Fatal("could not connect to postgres")
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.WithError(err).Fatal("error listening for notifications")
		}
	}

	listener := pq.NewListener(conninfo, 10*time.Second, time.Minute, reportProblem)

	// listen on each channel
	for channel := range routes {
		listener.Listen(channel)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	return listener
}

// Health simple healthcheck service
func Health(pg *pq.Listener) *http.ServeMux {

	healthz.Register("postgres", time.Second*5, func() error {
		return pg.Ping()
	})

	mux := http.NewServeMux()
	mux.Handle("/health", healthz.Handler())
	http.ListenAndServe(":"+os.Getenv("HEALTH_PORT"), mux)
	return mux
}
