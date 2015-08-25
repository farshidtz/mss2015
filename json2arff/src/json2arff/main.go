package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Entry struct {
	Time       string   `json:"t"`
	V0         *float64 `json:"v0"`
	V1         *float64 `json:"v1"`
	V2         *float64 `json:"v2"`
	V3         *float64 `json:"v3"`
	V4         *float64 `json:"v4"`
	V5         *float64 `json:"v5"`
	SensorName string   `json:"n"`
	Position   string   `json:"p"`
}

func writeFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename+".arff", data, 0644)
}

func addAttrs(buffer *bytes.Buffer, e Entry) {
	if e.V0 != nil {
		buffer.WriteString("@attribute v0 numeric\n")
	}
	if e.V1 != nil {
		buffer.WriteString("@attribute v1 numeric\n")
	}
	if e.V2 != nil {
		buffer.WriteString("@attribute v2 numeric\n")
	}
	if e.V3 != nil {
		buffer.WriteString("@attribute v3 numeric\n")
	}
	if e.V4 != nil {
		buffer.WriteString("@attribute v4 numeric\n")
	}
	if e.V5 != nil {
		buffer.WriteString("@attribute v5 numeric\n")
	}
}

func formatAttrs(e Entry) string {
	var buffer bytes.Buffer
	if e.V0 != nil {
		buffer.WriteString(fmt.Sprintf("%v ", *e.V0))
	}
	if e.V1 != nil {
		buffer.WriteString(fmt.Sprintf("%v ", *e.V1))
	}
	if e.V2 != nil {
		buffer.WriteString(fmt.Sprintf("%v ", *e.V2))
	}
	if e.V3 != nil {
		buffer.WriteString(fmt.Sprintf("%v ", *e.V3))
	}
	if e.V4 != nil {
		buffer.WriteString(fmt.Sprintf("%v ", *e.V4))
	}
	if e.V5 != nil {
		buffer.WriteString(fmt.Sprintf("%v ", *e.V5))
	}
	return buffer.String()
}

//var path = flag.String("data", "results.json", "json data file")
var query = flag.String("query", "", "e.g. -query=\"list2?n=acc&start=2015-08-25T12:14:08.229Z&end=2015-09-25T12:14:08.229Z\"")

func main() {
	flag.Parse()

	//	file, err := ioutil.ReadFile(*path)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		bufio.NewReader(os.Stdin).ReadBytes('\n')
	//		return
	//	}

	if *query == "" {
		flag.Usage()
		return
	}

	response, err := http.Get("http://46.101.133.187:8529/_db/_system/sensors-data-collector/" + *query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var data []Entry
	err = json.Unmarshal(contents, &data)
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	if len(data) == 0 {
		fmt.Println("Error: Response was empty.")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	var buffer bytes.Buffer

	// relation
	buffer.WriteString("@relation " + fmt.Sprintf("%v_%v", time.Now().Unix(), data[0].SensorName) + "\n\n")

	// attributes
	//buffer.WriteString("@attribute time numeric\n")
	addAttrs(&buffer, data[0])
	//buffer.WriteString("@attribute sensor {acc,gyro,baro,light}\n")
	buffer.WriteString("@attribute position {BackPocket,InHand}\n")
	buffer.WriteString("\n")

	// data
	buffer.WriteString("@data\n")
	for _, d := range data {
		//buffer.WriteString(fmt.Sprintf("%v %s%s %s\n", d.Time, formatAttrs(d), d.SensorName, d.Position))
		buffer.WriteString(fmt.Sprintf("%s%s\n", formatAttrs(d), d.Position))
	}

	// write to file
	err = ioutil.WriteFile(fmt.Sprintf("%v_%v.arff", time.Now().Unix(), data[0].SensorName), buffer.Bytes(), 0644)
	if err != nil {
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	return
}

// Sample .arff file
/*
@relation weka.datagenerators.classifiers.classification.Agrawal-S_1_-n_100_-F_1_-P_0.05

@attribute salary numeric
@attribute commission numeric
@attribute age numeric
@attribute elevel {0,1,2,3,4}
@attribute car {1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
@attribute zipcode {0,1,2,3,4,5,6,7,8}
@attribute hvalue numeric
@attribute hyears numeric
@attribute loan numeric
@attribute group {0,1}

@data
110499.735409,0,54,3,15,4,135000,30,354724.18253,1
*/
