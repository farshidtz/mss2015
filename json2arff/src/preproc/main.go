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
	LinAcc   float64
	Light    float64
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

func minInt(v1 uint64, vn ...uint64) (m uint64) {
	m = v1
	for i := 0; i < len(vn); i++ {
		if vn[i] < m {
			m = vn[i]
		}
	}
	return
}

func maxInt(v1 uint64, vn ...uint64) (m uint64) {
	m = v1
	for i := 0; i < len(vn); i++ {
		if vn[i] > m {
			m = vn[i]
		}
	}
	return
}

func filter(in []Entry, n string) []Entry {
	var slice []Entry
	for _, e := range in {
		if e.SensorName == n {
			slice = append(slice, e)
		}
	}
	return slice
}

func fillContinuousData(niceSession *NiceSession, in []Entry, n, p string) {
	slice := filter(in, n)

	for _, e := range slice {
		value := math.Sqrt(math.Pow(e.V0, 2) + math.Pow(e.V1, 2) + math.Pow(e.V2, 2))
		time := e.Time / 1000

		niceIndex := time - niceSession.StartTime
		if niceIndex >= uint64(len(niceSession.entries)) {
			fmt.Println(n, "out of bound:", niceIndex, ">=", len(niceSession.entries))
			continue
		}
		niceSession.entries[niceIndex].LinAcc = niceSession.entries[niceIndex].LinAcc + value
		niceSession.entries[niceIndex].Position = p
		niceSession.entries[niceIndex].Error = e.ErrorRate
		fmt.Println(n, time, niceSession.entries[niceIndex].LinAcc)
	}
}

func fillOnchangeData(niceSession *NiceSession, in []Entry, n, p string) {
	slice := filter(in, n)

	//var tempIndex uint64 = 0
	var totals = make([]float64, len(niceSession.entries))
	for _, e := range slice {
		value := e.V0
		time := e.Time / 1000

		niceIndex := time - niceSession.StartTime
		if niceIndex >= uint64(len(niceSession.entries)) {
			fmt.Println(n, "out of bound:", niceIndex, ">=", len(niceSession.entries))
			continue
		}
		niceSession.entries[niceIndex].Light = niceSession.entries[niceIndex].Light + value

		totals[niceIndex]++
		//tempIndex = niceIndex

		fmt.Println(n, time, niceSession.entries[niceIndex].Light, totals[niceIndex])
	}
	fmt.Println("--- Post ")
	var latestValue float64 = niceSession.entries[0].Light / totals[0]
	for i := range niceSession.entries {

		fmt.Println(n, uint64(i)+niceSession.StartTime, niceSession.entries[i].Light, totals[i])
		if totals[i] == 0 {
			niceSession.entries[i].Light = latestValue
		} else {
			niceSession.entries[i].Light = niceSession.entries[i].Light / totals[i]
			latestValue = niceSession.entries[i].Light
		}
		niceSession.entries[i].Position = p
		//niceSession.entries[i].Error = e.ErrorRate
	}
}

// Fill gaps with zeros
func fillContinuousGaps(in []Entry, n string) []Entry {
	slice := filter(in, n)

	start := slice[0].Time
	end := slice[len(slice)-1].Time
	var out = make([]Entry, end-start)
	fmt.Println("fillContinuousGaps", len(out))

	//var i uint64
	for i := range out {
		fmt.Println(i)
		out[i].Time = start + uint64(i)
		for j, e := range slice {
			fmt.Println("-", j)
			if e.Time == start+uint64(i) {
				out[i] = e
			} else if e.Time > start+uint64(i) {
				break
			}
		}
	}

	return out
}

// Fill gaps with most recent reading
func fillOnchangeGaps(in []Entry, n string) []Entry {
	//start := slice[0].Time
	//end := slice[len(slice)-1].Time
	//var out = make([]Entry, end-start)
	//fmt.Println(start, end)

	slice := filter(in, n)

	var out []Entry

	//var outHead uint64
	var i uint64
	for i = 0; i < uint64(len(slice))-1; i++ {

		if slice[i].ErrorRate != 0 {
			continue
		}
		if slice[i+1].ErrorRate != 0 {
			continue
		}
		time := slice[i].Time
		var head uint64
		for head = 0; head < slice[i+1].Time-time; head++ {
			out = append(out, Entry{
				Time: time + head,
				V0:   slice[i].V0,
			})
			fmt.Println("fill", i, slice[i].V0, head, out[len(out)-1].V0, slice[i+1].Time-time)
			//out[outHead].V0 = slice[i].V0
			//fmt.Println("fill", i, *slice[i].V0, outHead, slice[i+1].Time-start, *out[outHead].V0)
			//outHead++
			//head++
		}
	}

	return out
}

// Split sessions
func split(entries []Entry) []Session {
	var sessions []Session
	appendHead := 1
	for i := 1; i < len(entries); i++ {
		fmt.Println("split", i)
		if entries[i].ErrorRate == 100 {
			sessions = append(sessions, Session{entries[appendHead:i]})
			appendHead = i + 1
			fmt.Println("head", appendHead)
		}
		if i == len(entries)-1 { // EOF
			sessions = append(sessions, Session{entries[appendHead:i]})
		}
	}
	return sessions
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
	positions := []string{"SidePocket", "Idle"}

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

			//var entry NiceEntry { Time: s.entries[0].Time/1000 }
			//fmt.Println("filling", entry)

			fillContinuousData(&niceSession, s.entries, "linacc", p)
			fillOnchangeData(&niceSession, s.entries, "light", p)

			niceData = append(niceData, niceSession.entries...)

			//			offset := 0
			//			for {
			//				var totalAcc float64 = 0
			//				for i := offset; i < len(linacc); i++ {
			//					totalAcc = totalAcc +
			//						math.Sqrt(math.Pow(linacc[i].V0, 2)+math.Pow(linacc[i].V1, 2)+math.Pow(linacc[i].V2, 2))

			//					if i-offset >= 1000 {
			//						//offset = offset + i
			//						break
			//					}
			//				}

			//				var avgLight float64 = 0
			//				for i := offset; i < len(light); i++ {
			//					avgLight = avgLight + light[i].V0

			//					if i-offset >= 1000 {
			//						avgLight = avgLight / 1000
			//						//offset = offset + i
			//						break
			//					}
			//				}
			//				// global offset
			//				offset = offset + 1000

			//				niceEntry := NiceEntry{
			//					LinAcc:   totalAcc,
			//					Light:    avgLight,
			//					Position: p,
			//				}
			//				niceData = append(niceData, niceEntry)
			//			}
			//			break
		}
		allEntries = append(allEntries, niceData...)

		//		//from := maxInt(linacc[0].Time, light[0].Time)
		//		//to := minInt(linacc[len(linacc)-1].Time, light[len(light)-1].Time)

		//		linaccHead := 0
		//		lightHead := 0
		//		//totalSeconds := uint64((to - from) / 1000)
		//		//fmt.Println(from, to, totalSeconds)
		//		//var niceData = make([]NiceEntry, totalSeconds)
		//		//fmt.Println(len(niceData))
		//		var niceData []NiceEntry

		//		light = fillGaps(light)

		//		for {
		//			var niceEntry NiceEntry
		//			niceEntry.Position = p

		//			// linacc
		//			var totalAcc float64 = 0
		//			//			for i := linaccHead; i < len(linacc); i++ {
		//			//				if linacc[i].ErrorRate != 0 {
		//			//					continue
		//			//				}
		//			//				fmt.Println("linacc", T, i, linaccHead, linacc[i].Time)
		//			//				if linacc[i].Time > uint64(T*1000)+from && linacc[i].Time <= uint64((T+1)*1000)+from {
		//			//					// record it
		//			//					totalAcc = totalAcc +
		//			//						math.Sqrt(math.Pow(*linacc[i].V0, 2)+math.Pow(*linacc[i].V1, 2)+math.Pow(*linacc[i].V2, 2))
		//			//					fmt.Println("mag+", totalAcc)
		//			//				} else if linacc[i].Time > uint64((T+1)*1000)+from {
		//			//					linaccHead = i - 1
		//			//					break
		//			//				}
		//			//				linaccHead++
		//			//			}
		//			//fmt.Println("Recorded linacc", T, totalAcc)
		//			niceEntry.LinAcc = totalAcc

		//			// light
		//			var averageLight float64 = 0
		//			total := 0
		//			fmt.Println("fuck", lightHead, "to", lightHead+1000, len(light))

		//			offset := lightHead
		//			for j := lightHead; j < offset+1000 && j < len(light); j++ {

		//				averageLight = averageLight + *light[j].V0
		//				fmt.Println("avg+", averageLight, j, *light[j].V0)
		//				lightHead++
		//				total++
		//			}
		//			averageLight = averageLight / float64(total)
		//			//fmt.Println("Recorded light", T, averageLight)
		//			niceEntry.Light = averageLight

		//			if linaccHead >= len(linacc) ||
		//				lightHead >= len(light) {
		//				break
		//			}
		//			niceData = append(niceData, niceEntry)
		//		}
		//		allEntries = append(allEntries, niceData...)
	}

	//	linacc := queryData("linacc", "SidePocket", *start, *end)
	//	light := queryData("light", "SidePocket", *start, *end)

	//	from := maxInt(linacc[0].Time, light[0].Time)
	//	to := minInt(linacc[len(linacc)-1].Time, light[len(light)-1].Time)

	//	linaccHead := 0
	//	lightHead := 0
	//	totalSeconds := uint64((to - from) / 1000)
	//	fmt.Println(from, to, totalSeconds)
	//	var niceData = make([]NiceEntry, totalSeconds)
	//	fmt.Println(len(niceData))
	//	for T := range niceData {
	//		// linacc
	//		var totalAcc float64 = 0
	//		for i := linaccHead; i < len(linacc); i++ {
	//			fmt.Println("linacc", T, i, linaccHead, linacc[i].Time)
	//			if linacc[i].Time > uint64(T*1000)+from && linacc[i].Time <= uint64((T+1)*1000)+from {
	//				// record it
	//				totalAcc = totalAcc +
	//					math.Sqrt(math.Pow(*linacc[i].V0, 2)+math.Pow(*linacc[i].V1, 2)+math.Pow(*linacc[i].V2, 2))
	//				fmt.Println("mag+", totalAcc)
	//			} else if linacc[i].Time > uint64((T+1)*1000)+from {
	//				linaccHead = i - 1
	//				break
	//			}
	//		}
	//		fmt.Println("Recorded linacc", T, totalAcc)
	//		niceData[T].LinAcc = totalAcc

	//		// light
	//		var averageLight float64 = 0
	//		for i := lightHead; i < len(light); i++ {
	//			fmt.Println("light", T, i, lightHead, light[i].Time)
	//			if light[i].Time > uint64(T*1000)+from && light[i].Time <= uint64((T+1)*1000)+from {
	//				// record it
	//				averageLight = averageLight + *light[i].V0
	//				fmt.Println("avg+", averageLight, *light[i].V0)
	//			} else if light[i].Time > uint64((T+1)*1000)+from {
	//				lightHead = i - 1
	//				averageLight = averageLight / float64(i)
	//				break
	//			}
	//		}
	//		fmt.Println("Recorded light", T, averageLight)
	//		niceData[T].Light = averageLight

	//	}

	// Shuffle
	//	for i := range allEntries {
	//		j := rand.Intn(i + 1)
	//		allEntries[i], allEntries[j] = allEntries[j], allEntries[i]
	//	}

	var buffer bytes.Buffer

	// relation
	buffer.WriteString("@relation " + fmt.Sprintf("%v_%v", time.Now().Unix(), "complex_features") + "\n\n")

	// attributes
	buffer.WriteString("@attribute linacc numeric\n")
	buffer.WriteString("@attribute light numeric\n")
	buffer.WriteString("@attribute position {SidePocket,InHand}\n")
	buffer.WriteString("\n")

	// data
	buffer.WriteString("@data\n")
	for _, d := range allEntries {
		if d.Error == 0 {
			buffer.WriteString(fmt.Sprintf("%v %v %v\n", d.LinAcc, d.Light, d.Position))
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