package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/metakeule/fmtdate"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Boletim struct {
	Data             int64  `json:"data"`
	DataDesc         string `json:"data_desc"`
	Temperatura      string `json:"temperatura"`
	TempoDesc        string `json:"tempo_desc"`
	TemperaturaAgua  string `json:"temperatura_agua"`
	AguaDesc         string `json:"agua_desc"`
	NivelMar         string `json:"nivel_mar"`
	IscasDisponiveis string `json:"iscas"`
	PescadoresAgora  int    `json:"pescadores_agora"`
	VentoDesc        string `json:"vento_desc"`
	Correnteza       string `json:"correnteza"`
	Lua              string `json:"lua"`
}

func PadLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) >= lenght {
			return str[0:lenght]
		}
	}
}

func extractBoletinsFromSite() (boletins []Boletim) {
	boletins = []Boletim{}
	resp, err := http.Get("http://plataformadecidreira.com.br/comuns/boletim.php")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.Table {
			return true
		}
		return false
	}

	nodes := scrape.FindAll(root, matcher)

	for _, t := range nodes {
		boletins = append(boletins, extractBoletim(t))
	}

	return boletins
}

func extractBoletim(root *html.Node) (boletim Boletim) {
	boletim = Boletim{}

	m := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.Td {
			return true
		}
		return false
	}

	tds := scrape.FindAll(root, m)

	for _, td := range tds {
		content := scrape.Text(td)
		regx, _ := regexp.Compile(`(\d{2}/\d{2}/\d{4})\W+(\d+)`)
		hasData := regx.MatchString(content)

		if hasData {
			s := regx.FindAllString(content, 2)
			s = strings.Split(s[0], ",")
			dateDesc := strings.Trim(s[0], " ") + " " + PadLeft(strings.Trim(s[1], " "), "0", 2) + ":00"
			date, _ := fmtdate.Parse("DD/MM/YYYY hh:mm", dateDesc)
			boletim.Data = date.Unix()
			boletim.DataDesc = dateDesc
			continue
		}

		s := strings.Split(content, ":")

		if len(s) <= 1 {
			continue
		}

		s[1] = strings.Trim(s[1], " ")

		if strings.Contains(s[0], "Temperatura do ar") {
			boletim.Temperatura = s[1]
		} else if strings.Contains(s[0], "Tempo") {
			boletim.TempoDesc = s[1]
		} else if strings.Contains(s[0], "Água") {
			boletim.AguaDesc = s[1]
		} else if strings.Contains(s[0], "Nível") {
			boletim.NivelMar = s[1]
		} else if strings.Contains(s[0], "Isca") {
			boletim.IscasDisponiveis = s[1]
		} else if strings.Contains(s[0], "Pescadores") {
			i, err := strconv.Atoi(s[1])
			if err != nil {
				i = 0
			}
			boletim.PescadoresAgora = i
		} else if strings.Contains(s[0], "Temperatura da água") {
			boletim.TemperaturaAgua = s[1]
		} else if strings.Contains(s[0], "Vento") {
			boletim.VentoDesc = s[1]
		} else if strings.Contains(s[0], "Correnteza") {
			boletim.Correnteza = s[1]
		} else if strings.Contains(s[0], "Lua") {
			boletim.Lua = s[1]
		}
	}

	return boletim
}

func GetBoletimEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(extractBoletinsFromSite())
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/boletim", GetBoletimEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":8800", router))
	// test, _ := fmtdate.Parse("DD/MM/YYYY hh:mm", "31/08/2017 08:00")
	// fmt.Println(test)
}
