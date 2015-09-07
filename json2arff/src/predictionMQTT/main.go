package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	LinAcc   float64
	Light    float64
	Attr     [5]float64
	Position string
	Error    int
}

type NiceSession struct {
	StartTime uint64
	entries   []NiceEntry
}

func writeFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename+".arff", data, 0644)
}

func queryData(position, start, end string) []Entry {
	url := fmt.Sprintf("http://46.101.133.187:8529/_db/_system/sensors-data-collector/list3?p=%s&start=%s&end=%s", position, start, end)
	//fmt.Println("Query:", url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	}

	var data []Entry
	err = json.Unmarshal(contents, &data)
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	}
	if len(data) == 0 {
		fmt.Println("Error: Response was empty.")
		//bufio.NewReader(os.Stdin).ReadBytes('\n')
		//os.Exit(1)
	}

	return data
}

var (
	start = flag.String("start", "", "e.g. -start=\"2015-08-25T00:00Z\"")
	end   = flag.String("end", "", "e.g. -end=\"2016-01-01T00:00Z\"")
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client *MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {
	model := wekago.NewModel("functions.MultilayerPerceptron")
	model.LoadModel("3class.model")

	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("go-simple")
	opts.SetDefaultPublishHandler(f)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	ticker := time.NewTicker(time.Millisecond * 100)
	var wg sync.WaitGroup
	// Ctrl+C handling
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt, os.Kill)
	go func() {
		<-handler
		ticker.Stop()
		fmt.Println("^C Stopped.")
		fmt.Println("^C Waiting for remaining replies...")
		wg.Wait()

		fmt.Println("^C Done.")
		os.Exit(0)
	}()

	for t := range ticker.C {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			//Publish 5 messages at qos 1 and wait for the receipt
			//from the server after sending each message
			text := fmt.Sprintf("this is msg %v!", t)
			token := c.Publish("sensors/test", 0, false, text)
			token.Wait()
			wg.Done()
		}(&wg)
	}

	//time.Sleep(3 * time.Second)

	test_feature1 := wekago.NewFeature("linacc", "0.17700698435374612", "numeric")
	test_feature2 := wekago.NewFeature("rotation", "2.646118820703034", "numeric")
	test_feature3 := wekago.NewFeature("pressure", "2021.89797", "numeric")
	test_feature4 := wekago.NewFeature("light", "41", "numeric")
	test_feature5 := wekago.NewFeature("proximity", "5.000305", "numeric")
	outcome := wekago.NewFeature("position", "?", "{SidePocket,Idle,InHand}")

	test_instance1 := wekago.NewInstance()
	test_instance1.AddFeature(test_feature1)
	test_instance1.AddFeature(test_feature2)
	test_instance1.AddFeature(test_feature3)
	test_instance1.AddFeature(test_feature4)
	test_instance1.AddFeature(test_feature5)
	test_instance1.AddFeature(outcome)

	model.AddTestingInstance(test_instance1)

	err := model.Test()
	if err != nil {
		fmt.Println("Test error:", err.Error())
		return
	}

	for _, prediction := range model.Predictions {
		fmt.Printf("Index:%v Observed value:%s Prob:%v Prediction:%s\n",
			prediction.Index,
			prediction.Observed_value,
			prediction.Probability,
			prediction.Predicted_value,
		)
	}

	return
}
