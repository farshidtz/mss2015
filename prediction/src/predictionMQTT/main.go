package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/flyingsparx/wekago"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

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

type Session struct {
	entries []Entry
}

type NiceEntry struct {
	Attr     [10]*float64
	Position string
	Error    int
	Time     uint64
}

type NiceSession struct {
	StartTime uint64
	entries   []NiceEntry
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

const MQTT_TOPIC = "mss2015/sensors/data"

var (
	modelPath = flag.String("model", "", "Weka .model file path")
)

func main() {
	flag.Parse()
	if *modelPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("go-predictor")

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// Prediction handler
	var predict = func(niceEntry NiceEntry, submissionTime time.Time) {
		// Load the training model
		model := wekago.NewModel("functions.MultilayerPerceptron")
		//model := wekago.NewModel("lazy.IBk")
		model.LoadModel(*modelPath)

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
		outcome := wekago.NewFeature("position", "?", "{SidePocket,Idle,InHand}")

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
		test_instance1.AddFeature(outcome)

		model.AddTestingInstance(test_instance1)

		err := model.Test()
		if err != nil {
			fmt.Println("Test error:", err.Error())
			os.Exit(1)
		}

		for _, prediction := range model.Predictions {
			fmt.Printf("Prediction: %s\t Prob: %v\t Took: %v \n",
				prediction.Predicted_value,
				prediction.Probability,
				time.Since(submissionTime),
			)
		}
	}

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
		default:
			// system or unhandled sensor data
			//fmt.Println("Unrecognized sensor:", e.SensorName)
		}
		// Check whether a NiceEntry is fullfilled
		if niceEntry.Full() {
			go predict(niceEntry, time.Now())
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
	// Unsubscribe from topic
	if token := c.Unsubscribe(MQTT_TOPIC); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	c.Disconnect(250)
	fmt.Println("^C Stopped.")
	os.Exit(0)
}
