package main

import (
	"encoding/json"
	"fmt"

	//"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type App struct {
	Router *mux.Router
	DB     *gorm.DB
	//tmpl   *template.Template
}

type speedResult struct {
	ID            int64  `gorm:"column:id"`
	TimeStamp     string `gorm:"column:TimeStamp"`
	DownloadSpeed string `gorm:"column:DownloadSpeed"`
	UploadSpeed   string `gorm:"column:UploadSpeed"`
	Latency       string `gorm:"column:Latency"`
	PublicIp      string `gorm:"column:PublicIp"`
	ISP           string `gorm:"column:ISP"`
	Peers         string `gorm:"column:Peers"`
}

type tabler interface {
	TableName() string
}

func (speedResult) TableName() string {
	return "speedtest"
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) Initialize(user, password, dbname, host string) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, dbname)
	var err error
	a.DB, err = gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":"+addr, a.Router))
}

func getResults(db *gorm.DB) []speedResult {
	var speedtest []speedResult

	if err := db.Find(&speedtest).Error; err != nil {
		log.Printf("Error querying results: %v", err)
		return speedtest
	}
	log.Printf("%d rows found.", len(speedtest))
	return speedtest
}

func (a *App) getResults(w http.ResponseWriter, r *http.Request) {
	results := getResults(a.DB)
	if len(results) < 1 {
		fmt.Println("No results found")
	}

	respondWithJSON(w, http.StatusOK, results)
}

func main() {
	a := App{}
	a.Initialize(
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_DATABASE"),
		os.Getenv("MYSQL_HOST"),
	)

	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/", a.getResults).Methods("GET")
	a.Run(os.Getenv("HOST_PORT"))
}
