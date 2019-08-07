package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/WinPakin/ackpb"

	"github.com/gonum/stat"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// Ack appends to the end of the prev Msg
type ackmsg struct {
	Msg string `json:"msg"`
}

// APIStat corresponds to the ApiStat module in the Angular frontend
type APIStat struct {
	TestName string    `json:"testName"`
	APIType  string    `json:"apiType"`
	Mean     float64   `json:"mean"`
	Median   float64   `json:"median"`
	Max      float64   `json:"max"`
	Min      float64   `json:"min"`
	Std      float64   `json:"std"`
	Lst      []float64 `json:"lst"`
}

// NodeReq Request type that will be sent from the Node.js server
type nodeReq struct {
	TestName string `json:"testName"`
	APIType  string `json:"apiType"`
}

// SNDSERV host:port
// var SNDSERV string = "http://localhost:5000/ack";
var SNDSERV = "http://sndserv-service:80/ack"

// post makes one post request to the sndserv and checks for ack.
func post() {
	fstmsg := ackmsg{
		Msg: "from-testserv",
	}
	fmjson, err := json.Marshal(fstmsg)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := http.Post(SNDSERV, "application/json", bytes.NewBuffer(fmjson))
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var ack ackmsg
	err = json.Unmarshal(body, &ack)
	if err != nil {
		fmt.Println(err)
	}
	if ack.Msg != "from-testserv:ack-sndserv" {
		fmt.Println("WRONG MESSAGE: ", ack.Msg)
	}
}

// unaryCall makes a unaryCall to the gRPC server
func unaryCall(c ackpb.AckServiceClient) {
	req := &ackpb.AckReq{
		Msg: "from-testserv",
	}
	res, err := c.SendAck(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	} else if res.Msg != "from-testserv:rgpc-recv" {
		log.Panicln("wrong gRPC response message: ", res)
	}

}

// time records the number of milliseconds needed for each request
func timer(apitype string, c ackpb.AckServiceClient) float64 {
	var elapsed float64
	if apitype == "REST" {
		start := time.Now()
		post()
		elapsed = float64(time.Since(start)) / 1000000.0
	} else if apitype == "gRPC" {
		start := time.Now()
		unaryCall(c)
		elapsed = float64(time.Since(start)) / 1000000.0
	} else {
		log.Panicln("wrong API Type")
	}
	rounded := math.Round(elapsed*100) / 100
	return rounded

}

// findbuck, finds the index (bucket) for each time
func findbuck(x float64) int {
	if x < 9.0 {
		return int(math.Floor(x))
	}
	return 9

}

// sortedmedian finds the median of a sorted sorted slice
// if even number elements picks the one with a higher index
// 5 elements --> picks idx 2
// 6 element --> picks idx 3
func sortedmedian(lst []float64) float64 {
	mididx := len(lst) / 2
	return lst[mididx]
}

// allOnes creates []float64 of size x filled with ones
func allOnes(x int) []float64 {
	w := make([]float64, x, x)
	for i, _ := range w {
		w[i] = 1
	}
	return w
}

// roundTwoDeci rounds to two decimal places
func roundTwoDeci(x float64) float64 {
	return math.Round(x*100) / 100
}

// testAPI repeated tests the latency of either REST or gRPC apis
func testAPI(testName string, apiType string, c ackpb.AckServiceClient) APIStat {
	lstall := make([]float64, 30, 30)
	lst := make([]float64, 10, 10)
	min := math.MaxFloat64
	max := -math.MaxFloat64
	var t float64
	for i := 0; i < 30; i++ {
		t = timer(apiType, c)
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
		lstall[i] = t
		lst[findbuck(t)]++
	}
	sort.Float64s(lstall)
	median := sortedmedian(lstall)
	// weights
	w := allOnes(30)
	std := stat.StdDev(lstall, w)
	mean := stat.Mean(lstall, w)
	// data to be sent
	data := APIStat{
		TestName: testName,
		APIType:  apiType,
		Mean:     roundTwoDeci(mean),
		Median:   median,
		Max:      max,
		Min:      min,
		Std:      roundTwoDeci(std),
		Lst:      lst,
	}
	return data

}

// GRPCSERV host:port
// var GRPCSERV = "localhost:5001"
var GRPCSERV = "grpcserv-service:80"

func handleFunc(w http.ResponseWriter, r *http.Request) {
	// Set up Connection for gRPC
	opts := grpc.WithInsecure()
	cc, err := grpc.Dial(GRPCSERV, opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()
	c := ackpb.NewAckServiceClient(cc)
	// fmt.Println(testAPI("first test", "REST", c))
	// fmt.Println(testAPI("first test", "gRPC", c))

	// Responding to Node with stats
	fmt.Println("got req from Node!")
	w.Header().Set("Content-Type", "application/json")
	var reqmsg nodeReq
	_ = json.NewDecoder(r.Body).Decode(&reqmsg)
	testStat := testAPI(reqmsg.TestName, reqmsg.APIType, c)
	json.NewEncoder(w).Encode(testStat)
}

func main() {
	fmt.Println("go test server running ...")
	r := mux.NewRouter()
	r.HandleFunc("/gotest", handleFunc).Methods("POST")
	log.Fatal(http.ListenAndServe(":5002", r))

}
