package main

import (
	"os"
	"fmt"
	"go-binance/binance"
	"github.com/crackcomm/go-clitable"
	"time"
	"math"
	"syscall"
)

var (
	beepFunc = syscall.MustLoadDLL("user32.dll").MustFindProc("MessageBeep")
)

func main() {
	for {
		exchageCount()
		createTable()
		time.Sleep(time.Second * 30)
	}
}

var Init = 0
var Coins = make(map[string]Coin)
var ExchangeInit = 0
var Exchanges = make(map[string]float64)
const PriceThreshold = 300

type Coin struct {
	Amount float64
	CostEach float64
	Total float64
}

func createTable() {
	mytable := clitable.New([]string{"Name", "Coins Amount", "EACH (USD)", "TOTAL (USD)", "PRICE CHANGE"})
	client := binance.New(os.Getenv("BINANCE_KEY"), os.Getenv("BINANCE_SECRET"))
	positions, err := client.GetPositions()

	if err != nil {
		panic(err)
	}
	btcPrice := getBtc()
	total := 0.0
	//percentTotal := 0.0

	for _, p := range positions {
		//fmt.Println(p.Asset, p.Free, p.Locked)
		symbol := p.Asset+"BTC"
		query := binance.SymbolQuery {
			Symbol: symbol,
		}

		client := binance.New("", "")
		res, err := client.GetLastPrice(query)

		if err != nil {
			panic(err)
		}


		//Set baselines.
		if Init == 0 {
			Coins[p.Asset] = Coin{Amount: p.Free, CostEach: res.Price * btcPrice, Total: res.Price * btcPrice * p.Free}
		} else {
			if res.Price*btcPrice*p.Free > PriceThreshold {
				mytable.AddRow(map[string]interface{}{
					"Name":         p.Asset,
					"Coins Amount": p.Free,
					"EACH (USD)":   res.Price * btcPrice,
					"TOTAL (USD)":  res.Price * btcPrice * p.Free,
					"PRICE CHANGE": Round(((res.Price * btcPrice) - Coins[p.Asset].CostEach)/(res.Price * btcPrice)*100, .5, 2),
				})

				if ((res.Price * btcPrice) - Coins[p.Asset].CostEach)/(res.Price * btcPrice)*100 > 10 {
					fmt.Println("PRICE CHANGE for:", p.Asset)
					beepFunc.Call(0xffffffff)
					time.Sleep(time.Microsecond * 500)
					beepFunc.Call(0xffffffff)
				}
				if Coins[p.Asset].Amount != p.Free {
					//UPDATE THE RECORD
					//Make sure this works?? tested here... https://play.golang.org/p/shCfvFYZylc didnt work in real life though...
					Coins[p.Asset] = Coin{Amount: p.Free, CostEach: res.Price * btcPrice, Total: res.Price * btcPrice * p.Free}
				}
			}

		}

		if (p.Asset == "BTC"){
			total = total + btcPrice*p.Free
		} else {
			total = total + res.Price*btcPrice*p.Free
		}
	}

	//finished initializing everything
	Init = 1
	//fmt.Println(Coins)

	mytable.Print()

	fmt.Println("-----------------TOTAL---------------")
	fmt.Println("here is the total USD: $", Round(total, .5, 2))

}


func getBtc() float64 {
	query := binance.SymbolQuery {
		Symbol: "BTCUSDT",
	}

	client := binance.New("", "")
	res, err := client.GetLastPrice(query)

	if err != nil {
		panic(err)
	}
	return res.Price
}

func Round(val float64, roundOn float64, places int ) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func exchageCount() {
	client := binance.New("", "")
	results, err := client.GetAllPrices()
	if err != nil {
		fmt.Println("error with the Binance Client request: ", err)
		return
	}

	if len(results) != ExchangeInit {
		fmt.Println("There was a difference of: ", len(results) - ExchangeInit, " So we are adding them.")
		ExchangeInit = len(results)

		for _, resultExchange := range results {
			found := 0
			for name, _ := range Exchanges {
				if name == resultExchange.Symbol {
					found = 1
				}
			}

			if found == 0 {
				//fmt.Println("We did not find: ", resultExchange.Symbol, " adding to Map List")
				Exchanges[resultExchange.Symbol] = resultExchange.Price

			}

		}

	}
}
