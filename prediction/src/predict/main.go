package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/fatih/color"
)

const (
	// Confidence above this value is excellent
	CONFIDENCE_BOUNDARY = 0.9
	// Confidence below this value is noise
	NOISE_BOUNDARY = 0.5
)

// MQTT
const (
	MQTT_TOPIC  = "mss2015/sensors/data"
	MQTT_PUBLIC = "tcp://iot.eclipse.org:1883"
	MQTT_SERVER = "tcp://localhost:1883"
)

// Weka Remote Server
const (
	WEKA_SERVER_JAR  = "../WekaRemote-3.7.12.jar"
	WEKA_SERVER_PORT = 9100
)

// Raw entry
type Entry struct {
	Time       uint64  `json:"t"`
	V0         float64 `json:"v0"`
	V1         float64 `json:"v1"`
	V2         float64 `json:"v2"`
	V3         float64 `json:"v3"`
	V4         float64 `json:"v4"`
	V5         float64 `json:"v5"`
	SensorName string  `json:"n"`
	Position   string  `json:"p"`
	ErrorRate  int     `json:"e"`
}

func magnitude(V0, V1, V2 float64) float64 {
	return math.Sqrt(math.Pow(V0, 2) + math.Pow(V1, 2) + math.Pow(V2, 2))
}

// Entry containing all features
type NiceEntry struct {
	Attr [13]*float64 `json:"attributes"`
}

// Returns true of all attributes are set
func (niceEntry *NiceEntry) Full() bool {
	for i := range niceEntry.Attr {
		if niceEntry.Attr[i] == nil {
			return false
		}
	}
	return true
}

// Prediction results
type Prediction struct {
	Label        string    `json:"label"`
	Distribution []float64 `json:"dist"`
	Index        int       `json:"index"`
}

// Global variables
var modelPath string
var conn net.Conn

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("Usage: ./predict weka_model_file")
		fmt.Println("Example: ./predict mlp_new.model\n")
		os.Exit(1)
	}
	modelPath = flag.Args()[0]

	// Run Weka server
	wekaServer := make(chan struct{}, 1)
	runWekaServer(wekaServer, WEKA_SERVER_PORT)

	// Connect to Weka Server socket
	var err error
	conn, err = net.Dial("tcp", "localhost:"+fmt.Sprint(WEKA_SERVER_PORT))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(conn, modelPath+"\n")
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if strings.Contains(reply, "ERROR") {
		fmt.Println(reply)
		os.Exit(1)
	}

	// Create MQTT Client
	opts := MQTT.NewClientOptions().AddBroker(MQTT_SERVER)
	opts.SetClientID("go-predictor")
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup

	// MQTT message handler (the sensor listener)
	var niceEntry NiceEntry
	var mqttHandler MQTT.MessageHandler = func(client *MQTT.Client, msg MQTT.Message) {
		//fmt.Printf("TOPIC: %s\n", msg.Topic())
		//fmt.Println(string(msg.Payload()))
		var e Entry
		err := json.Unmarshal(msg.Payload(), &e)
		if err != nil {
			fmt.Println(err.Error())
		}

		switch e.SensorName {
		case "linacc":
			niceEntry.Attr[0] = &e.V0
			niceEntry.Attr[1] = &e.V1
			niceEntry.Attr[2] = &e.V2
			mag := magnitude(e.V0, e.V2, e.V3)
			niceEntry.Attr[3] = &mag
		case "rotation":
			niceEntry.Attr[4] = &e.V0
			niceEntry.Attr[5] = &e.V1
			niceEntry.Attr[6] = &e.V2
		case "pressure":
			niceEntry.Attr[7] = &e.V0
		case "light":
			niceEntry.Attr[8] = &e.V0
		case "proximity":
			niceEntry.Attr[9] = &e.V0
		case "gravity":
			niceEntry.Attr[10] = &e.V0
			niceEntry.Attr[11] = &e.V1
			niceEntry.Attr[12] = &e.V2
		default:
			// system or unhandled sensor data
			//fmt.Println("Unrecognized sensor:", e.SensorName)
		}
		// Check whether a NiceEntry is fullfilled
		//fmt.Println(niceEntry)
		if niceEntry.Full() {
			go predict(&wg, niceEntry, time.Now())
		} else {
			fmt.Println("Warming up.")
		}
	}

	// Subscribe to te topic
	if token := c.Subscribe(MQTT_TOPIC, 1, mqttHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// Ctrl+C handling
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt, os.Kill)
	<-handler // block the thread
	fmt.Println("^C Waiting for threads to close...")
	wg.Wait()
	// Unsubscribe from topic
	if token := c.Unsubscribe(MQTT_TOPIC); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	c.Disconnect(250)
	close(wekaServer)
	fmt.Println("Done.")
	os.Exit(0)
}

// Run Java Weka server
func runWekaServer(wekaServer chan struct{}, port uint) {
	cmd := exec.Command("java", "-jar", WEKA_SERVER_JAR, fmt.Sprint(port))
	err := cmd.Start()
	if err != nil {
		log.Fatal("Weka: " + err.Error())
	}
	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()

		select {
		case <-wekaServer:
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("Weka: Failed to kill: ", err)
			}
			<-done // allow goroutine to exit
			fmt.Println("Weka: Process killed")
		case err := <-done:
			if err != nil {
				fmt.Println("Weka: Process done with error: %v", err)
			}
		}
	}()
}

// Prediction handler
func predict(wg *sync.WaitGroup, niceEntry NiceEntry, submissionTime time.Time) {
	wg.Add(1)
	defer wg.Done()

	jsonObj, err := json.Marshal(&niceEntry)
	//fmt.Println(string(json))
	fmt.Fprintf(conn, string(jsonObj)+"\n")

	// listen for reply
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	reply = strings.TrimSpace(reply)

	// Unmarshall from json
	var pred Prediction
	err = json.Unmarshal([]byte(reply), &pred)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Format the color of output
	format := color.New(color.FgGreen, color.Bold)
	if pred.Distribution[pred.Index] >= CONFIDENCE_BOUNDARY {
		format.EnableColor()
	} else if pred.Distribution[pred.Index] < NOISE_BOUNDARY {
		format = color.New(color.FgRed, color.Bold)
	} else {
		format.DisableColor()
	}

	format.Printf("Prediction: %s\t Confidence: %.3f\t Took: %v \n",
		pred.Label,
		pred.Distribution[pred.Index],
		time.Since(submissionTime),
	)

}
