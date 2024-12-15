package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"

    "gonum.org/v1/plot"
    "gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
)

var apiKey string

type Data struct {
    Country     string  `json:"Country"`
    Category    string  `json:"Category"`
    LatestValue float64 `json:"LatestValue"`
}

func main() {
    // Read the API key from the external file
    key, err := ioutil.ReadFile("apikey.cfg")
    if err != nil {
        fmt.Println("Error reading API key:", err)
        return
    }
    apiKey = string(key)

    if len(os.Args) < 3 {
        fmt.Println("Usage: go run main.go <country1> <country2>")
        return
    }

    country1 := os.Args[1]
    country2 := os.Args[2]

    rawData1 := fetchRawData("https://api.tradingeconomics.com/country/" + country1 + "?c=" + apiKey + "&group=gdp")
    fmt.Println("Raw Data for", country1, ":", string(rawData1))

    // Wait for 5 seconds before making the second API request
    time.Sleep(5 * time.Second)

    rawData2 := fetchRawData("https://api.tradingeconomics.com/country/" + country2 + "?c=" + apiKey + "&group=gdp")
    fmt.Println("Raw Data for", country2, ":", string(rawData2))

    data1 := parseData(rawData1)
    data2 := parseData(rawData2)

    fmt.Println("Category\t\t\tLatest Value")
    fmt.Println("-------------------------------------------------")

    for _, item := range data1 {
        if item.Category == "Full Year GDP Growth" {
            fmt.Printf("%s (%s)\t\t%.2f\n", item.Category, country1, item.LatestValue)
        }
    }

    for _, item := range data2 {
        if item.Category == "Full Year GDP Growth" {
            fmt.Printf("%s (%s)\t\t%.2f\n", item.Category, country2, item.LatestValue)
        }
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        serveChart(w, data1, data2, country1, country2)
    })

    fmt.Println("Server started at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func fetchRawData(url string) []byte {
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return nil
    }

    return body
}

func parseData(rawData []byte) []Data {
    var data []Data
    var errorResponse ErrorResponse

    err := json.Unmarshal(rawData, &data)
    if err != nil {
        err = json.Unmarshal(rawData, &errorResponse)
        if err == nil {
            fmt.Println("Error:", errorResponse.Message)
        } else {
            fmt.Println("Error parsing data:", err)
        }
        return nil
    }

    return data
}

func serveChart(w http.ResponseWriter, data1, data2 []Data, country1, country2 string) {
    values := []float64{}
    labels := []string{}

    for _, item := range data1 {
        if item.Category == "Full Year GDP Growth" {
            values = append(values, item.LatestValue)
            labels = append(labels, country1)
        }
    }

    for _, item := range data2 {
        if item.Category == "Full Year GDP Growth" {
            values = append(values, item.LatestValue)
            labels = append(labels, country2)
        }
    }

    createBarChart(values, labels, w)
}

func createBarChart(values []float64, labels []string, w http.ResponseWriter) {
    p := plot.New()
    p.Title.Text = "Full Year GDP Growth"
    p.Y.Label.Text = "Latest Value"

    bars, err := plotter.NewBarChart(plotter.Values(values), vg.Points(20))
    if err != nil {
        panic(err)
    }

    p.Add(bars)
    p.NominalX(labels...)

    w.Header().Set("Content-Type", "image/png")
    writerTo, err := p.WriterTo(4*vg.Inch, 4*vg.Inch, "png")
    if err != nil {
        panic(err)
    }
    writerTo.WriteTo(w)
}

type ErrorResponse struct {
    Message string `json:"Message"`
}
