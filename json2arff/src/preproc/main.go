package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
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
	Attr     [10]float64
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
	fmt.Println("Query:", url)
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
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		//os.Exit(1)
	}

	return data
}

func filterBySensor(in []Entry, n string) []Entry {
	var slice []Entry
	for _, e := range in {
		if e.SensorName == n {
			slice = append(slice, e)
		}
	}
	return slice
}

func fillContinuousData(niceSession *NiceSession, in []Entry, n, p string, attrNo, totalAttr int, magnitude bool) {
	slice := filterBySensor(in, n)

	if magnitude && totalAttr != 3 {
		fmt.Println(n, "Magnitute needs 3 attributes!")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		os.Exit(1)
	}

	fmt.Println(n, "Pre Proc...")
	var totals = make([]float64, len(niceSession.entries))
	for _, e := range slice {
		time := e.Time / 1000

		niceIndex := time - niceSession.StartTime
		if niceIndex >= uint64(len(niceSession.entries)) {
			fmt.Println(n, "out of bound:", niceIndex, ">=", len(niceSession.entries))
			continue
		}
		if totalAttr == 1 {
			niceSession.entries[niceIndex].Attr[attrNo] = niceSession.entries[niceIndex].Attr[attrNo] + e.V0
		} else if totalAttr == 3 {
			niceSession.entries[niceIndex].Attr[attrNo] = niceSession.entries[niceIndex].Attr[attrNo] + e.V0
			niceSession.entries[niceIndex].Attr[attrNo+1] = niceSession.entries[niceIndex].Attr[attrNo+1] + e.V1
			niceSession.entries[niceIndex].Attr[attrNo+2] = niceSession.entries[niceIndex].Attr[attrNo+2] + e.V2
		} else {
			fmt.Println(n, "Total attribute", totalAttr, "not supported!")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(1)
		}
		if magnitude {
			magnitude := math.Sqrt(math.Pow(e.V0, 2) + math.Pow(e.V1, 2) + math.Pow(e.V2, 2))
			niceSession.entries[niceIndex].Attr[attrNo+3] = niceSession.entries[niceIndex].Attr[attrNo+3] + magnitude
		}

		totals[niceIndex]++
		niceSession.entries[niceIndex].Position = p
		niceSession.entries[niceIndex].Error = e.ErrorRate

		fmt.Println(n, time, niceSession.entries[niceIndex].Attr[attrNo])
	}
	fmt.Println(n, "Post Proc...")
	for i := range niceSession.entries {
		for t := 0; t < totalAttr; t++ {
			niceSession.entries[i].Attr[attrNo+t] = niceSession.entries[i].Attr[attrNo+t] / totals[i]
		}
		if magnitude {
			niceSession.entries[i].Attr[attrNo+3] = niceSession.entries[i].Attr[attrNo+3] / totals[i]
		}
	}
}

func fillOnchangeData(niceSession *NiceSession, in []Entry, n, p string, attrNo int) {
	slice := filterBySensor(in, n)

	fmt.Println(n, "Pre Proc...")
	var totals = make([]float64, len(niceSession.entries))
	for _, e := range slice {
		time := e.Time / 1000

		niceIndex := time - niceSession.StartTime
		if niceIndex >= uint64(len(niceSession.entries)) {
			fmt.Println(n, "out of bound:", niceIndex, ">=", len(niceSession.entries))
			continue
		}
		niceSession.entries[niceIndex].Attr[attrNo] = niceSession.entries[niceIndex].Attr[attrNo] + e.V0
		totals[niceIndex]++

		fmt.Println(n, time, niceSession.entries[niceIndex].Attr[attrNo], totals[niceIndex])
	}
	fmt.Println(n, "Post Proc...")
	var latestValue float64 = niceSession.entries[0].Attr[attrNo] / totals[0]
	for i := range niceSession.entries {
		//fmt.Println(n, uint64(i)+niceSession.StartTime, niceSession.entries[i].Attr[attrNo], totals[i])
		if totals[i] == 0 {
			niceSession.entries[i].Attr[attrNo] = latestValue
		} else {
			niceSession.entries[i].Attr[attrNo] = niceSession.entries[i].Attr[attrNo] / totals[i]
			latestValue = niceSession.entries[i].Attr[attrNo]
		}
		niceSession.entries[i].Position = p
		//niceSession.entries[i].Error = e.ErrorRate
	}
}

// Split sessions
func split(entries []Entry) []Session {
	var sessions []Session
	fmt.Println("Splitting sessions...")
	appendHead := 1
	for i := 1; i < len(entries); i++ {
		//fmt.Println("split", i)
		if entries[i].ErrorRate == 100 {
			sessions = append(sessions, Session{entries[appendHead:i]})
			appendHead = i + 1
			//fmt.Println("head", appendHead)
		}
		if i == len(entries)-1 { // EOF
			sessions = append(sessions, Session{entries[appendHead:i]})
		}
	}
	return sessions
}

func attributes(e *NiceEntry) string {
	var buffer bytes.Buffer
	for _, a := range e.Attr {
		buffer.WriteString(fmt.Sprintf("%v ", a))
	}
	return buffer.String()
}

var (
	start = flag.String("start", "", "e.g. -start=\"2015-08-25T00:00Z\"")
	end   = flag.String("end", "", "e.g. -end=\"2016-01-01T00:00Z\"")
)

func main() {
	flag.Parse()
	if *start == "" || *end == "" {
		flag.Usage()
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	//sensors := []string{"linacc", "rotation", "light", "proximity", "pressure"}
	positions := []string{"SidePocket", "Idle", "InHand"}

	var allEntries []NiceEntry
	for _, p := range positions {
		sessions := split(queryData(p, *start, *end))
		fmt.Println(p, len(sessions))

		var niceData []NiceEntry

		for _, s := range sessions {
			fmt.Println(len(sessions))
			fmt.Println(len(s.entries))
			first := s.entries[0].Time
			last := s.entries[len(s.entries)-1].Time
			totalSeconds := uint64((last - first) / 1000)
			fmt.Println(first, last, totalSeconds)
			fmt.Println(len(niceData))
			niceSession := NiceSession{
				entries:   make([]NiceEntry, totalSeconds),
				StartTime: first / 1000,
			}

			fillContinuousData(&niceSession, s.entries, "linacc", p, 0, 3, true)    // 0,1,2,3
			fillContinuousData(&niceSession, s.entries, "rotation", p, 4, 3, false) // 4,5,6
			fillContinuousData(&niceSession, s.entries, "pressure", p, 7, 1, false) // 7
			fillOnchangeData(&niceSession, s.entries, "light", p, 8)
			fillOnchangeData(&niceSession, s.entries, "proximity", p, 9)

			niceData = append(niceData, niceSession.entries...)

		}
		allEntries = append(allEntries, niceData...)
	}

	//// Shuffle
	//	for i := range allEntries {
	//		j := rand.Intn(i + 1)
	//		allEntries[i], allEntries[j] = allEntries[j], allEntries[i]
	//	}

	var buffer bytes.Buffer

	// relation
	buffer.WriteString("@relation " + fmt.Sprintf("%v_%v", time.Now().Unix(), "complex_features") + "\n\n")

	// attributes
	buffer.WriteString("@attribute linaccX numeric\n")
	buffer.WriteString("@attribute linaccY numeric\n")
	buffer.WriteString("@attribute linaccZ numeric\n")
	buffer.WriteString("@attribute linaccMag numeric\n")
	buffer.WriteString("@attribute rotationX numeric\n")
	buffer.WriteString("@attribute rotationY numeric\n")
	buffer.WriteString("@attribute rotationZ numeric\n")
	buffer.WriteString("@attribute pressure numeric\n")
	buffer.WriteString("@attribute light numeric\n")
	buffer.WriteString("@attribute proximity numeric\n")
	buffer.WriteString("@attribute position {" + strings.Join(positions, ",") + "}\n")
	buffer.WriteString("\n")

	// data
	buffer.WriteString("@data\n")
	for _, d := range allEntries {
		if d.Error == 0 {
			buffer.WriteString(fmt.Sprintf("%v%v\n",
				attributes(&d),
				d.Position,
			))
		}
	}

	// write to file
	err := ioutil.WriteFile(fmt.Sprintf("%v_%v.arff", time.Now().Unix(), "complex_features"), buffer.Bytes(), 0644)
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
