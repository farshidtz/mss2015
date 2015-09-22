package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/flyingsparx/wekago"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

// MQTT Subscription topic
const MQTT_TOPIC = "mss2015/sensors/data"

//const MQTT_SERVER = "tcp://iot.eclipse.org:1883"
const MQTT_SERVER = "tcp://192.168.1.42:1883"

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

type NiceEntry struct {
	Attr [13]*float64
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

func magnitude(V0, V1, V2 float64) float64 {
	return math.Sqrt(math.Pow(V0, 2) + math.Pow(V1, 2) + math.Pow(V2, 2))
}

// Global variables
var modelName, modelPath string

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: ./predict weka_model_name weka_model_file")
		fmt.Println("Example: ./predict functions.MultilayerPerceptron trained.model\n")
		os.Exit(1)
	}
	modelName = flag.Args()[0]
	modelPath = flag.Args()[1]

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
	fmt.Println("Done.")
	os.Exit(0)
}

// Prediction handler
func predict(wg *sync.WaitGroup, niceEntry NiceEntry, submissionTime time.Time) {
	wg.Add(1)
	defer wg.Done()

	// Load the training model
	model := wekago.NewModel(modelName)
	model.LoadModel(modelPath)

	test_feature0 := wekago.NewFeature("linaccX", fmt.Sprint(*niceEntry.Attr[0]), "numeric")
	test_feature1 := wekago.NewFeature("linaccY", fmt.Sprint(*niceEntry.Attr[1]), "numeric")
	test_feature2 := wekago.NewFeature("linaccZ", fmt.Sprint(*niceEntry.Attr[2]), "numeric")
	test_feature3 := wekago.NewFeature("linaccMag", fmt.Sprint(*niceEntry.Attr[3]), "numeric")
	test_feature4 := wekago.NewFeature("rotationX", fmt.Sprint(*niceEntry.Attr[4]), "numeric")
	test_feature5 := wekago.NewFeature("rotationY", fmt.Sprint(*niceEntry.Attr[5]), "numeric")
	test_feature6 := wekago.NewFeature("rotationZ", fmt.Sprint(*niceEntry.Attr[6]), "numeric")
	test_feature7 := wekago.NewFeature("pressure", fmt.Sprint(*niceEntry.Attr[7]), "numeric")
	test_feature8 := wekago.NewFeature("light", fmt.Sprint(*niceEntry.Attr[8]), "numeric")
	test_feature9 := wekago.NewFeature("proximity", fmt.Sprint(*niceEntry.Attr[9]), "numeric")
	test_feature10 := wekago.NewFeature("gravityX", fmt.Sprint(*niceEntry.Attr[10]), "numeric")
	test_feature11 := wekago.NewFeature("gravityY", fmt.Sprint(*niceEntry.Attr[11]), "numeric")
	test_feature12 := wekago.NewFeature("gravityZ", fmt.Sprint(*niceEntry.Attr[12]), "numeric")
	outcome := wekago.NewFeature("position", "?", "{SidePocket,Idle,InHand,Handbag}")

	test_instance1 := wekago.NewInstance()
	test_instance1.AddFeature(test_feature0)
	test_instance1.AddFeature(test_feature1)
	test_instance1.AddFeature(test_feature2)
	test_instance1.AddFeature(test_feature3)
	test_instance1.AddFeature(test_feature4)
	test_instance1.AddFeature(test_feature5)
	test_instance1.AddFeature(test_feature6)
	test_instance1.AddFeature(test_feature7)
	test_instance1.AddFeature(test_feature8)
	test_instance1.AddFeature(test_feature9)
	test_instance1.AddFeature(test_feature10)
	test_instance1.AddFeature(test_feature11)
	test_instance1.AddFeature(test_feature12)
	test_instance1.AddFeature(outcome)

	model.AddTestingInstance(test_instance1)

	err := model.Test()
	if err != nil {
		fmt.Println("[Java]", err.Error())
		return
	}

	for _, prediction := range model.Predictions {
		fmt.Printf("Prediction: %s\t Prob: %v\t Took: %v \n",
			prediction.Predicted_value,
			prediction.Probability,
			time.Since(submissionTime),
		)
	}
}
