package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type speedTestRecords struct {
	Timestamp string `json:"timestamp"`
	Ping      struct {
		Latency float64 `json:"latency"`
	} `json:"ping"`
	Download struct {
		Bandwidth float64 `json:"bandwidth"`
	} `json:"download"`
	Upload struct {
		Bandwidth float64 `json:"bandwidth"`
	} `json:"upload"`
	ISP       string `json:"isp"`
	Interface struct {
		ExternalIP string `json:"externalIp"`
	} `json:"interface"`
	Server struct {
		Name     string `json:"name"`
		Country  string `json:"country"`
		Location string `json:"location"`
	} `json:"server"`
}

func getSpeedTestResult() (string, float64, float64, float64, string, string, string, string, string) {

	flag.Parse()
	args := flag.Args()
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	var jsonLine string
	for _, line := range strings.Split(string(stdout), "\n") {
		if strings.HasPrefix(line, "{") {
			jsonLine = line
			break
		}
	}
	results := speedTestRecords{}
	json.Unmarshal([]byte(jsonLine), &results)

	upload := math.RoundToEven(results.Upload.Bandwidth / 125000)
	download := math.RoundToEven(results.Download.Bandwidth / 125000)
	name := results.Server.Name
	country := results.Server.Country
	sponsor := results.Server.Location
	latency := results.Ping.Latency
	ipAddr := results.Interface.ExternalIP
	isp := results.ISP
	t, err := time.Parse(time.RFC3339, results.Timestamp)
	if err != nil {
		fmt.Println("Timestamp parse error:", err.Error())
		panic(err.Error())
	}
	timestamp := t.Format("2006-01-02 15:04:05")

	return timestamp, download, upload, latency, ipAddr, isp, sponsor, name, country
}

func dbConnect(timestamp string, download float64, upload float64, latency float64, ipAddr string, isp string, peerServer string) {

	pswd := os.Getenv("MYSQL_PASSWORD")
	user := os.Getenv("MYSQL_USER")
	database := os.Getenv("MYSQL_DATABASE")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	dsn := user + ":" + pswd + "@tcp(" + host + ":" + port + ")/" + database

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Error while validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	conn := db.Ping()
	if conn != nil {
		fmt.Println("Error while connecting to database")
		panic(conn.Error())
	}

	insert, err := db.Prepare("INSERT INTO speedtest.speedtest (TimeStamp, DownloadSpeed, UploadSpeed, Latency, PublicIp, ISP, Peers) VALUES (? ,? ,? ,? ,? ,? ,?)")
	if err != nil {
		panic(err.Error())
	}
	result, err := insert.Exec(timestamp, download, upload, latency, ipAddr, isp, peerServer)
	if err != nil {
		fmt.Println("Insert error:", err.Error())
		panic(err.Error())
	}
	rows, _ := result.RowsAffected()
	fmt.Println("Rows affected:", rows)
	defer insert.Close()
}

func main() {
	timestamp, download, upload, latency, ipAddr, isp, sponsor, name, country := getSpeedTestResult()
	peerServer := sponsor + " " + name + " " + " " + country
	dbConnect(timestamp, download, upload, latency, ipAddr, isp, peerServer)
}
