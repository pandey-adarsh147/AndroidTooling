package main

import (
	"bufio"
	"os"
	"fmt"
	"regexp"
	"strconv"
	"sort"
	"sync"
	"strings"
	"os/exec"
)

var myExp *regexp.Regexp

type UrlValue struct {
	totalByte int
	totalTime int
	maxTime int
	count int
}

func main() {

	arguments := os.Args
	var fileName string
	if len(arguments) > 2 {
		fileName = arguments[1]
	} else {
		var wg *sync.WaitGroup = new(sync.WaitGroup)
		wg.Add(1)
		go exe_cmd("adb pull sdcard/helpchat_api.txt", wg)
		wg.Wait()
		fileName = "helpchat_api.txt"
	}

	fmt.Println(fileName)
	file, err := os.Open(fileName)
	if err != nil {

	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	myExp = regexp.MustCompile(`(?P<time>\d+:\d+:\d+): (?P<type>RESPONSE) (?P<code>\d+) (?P<url>https?:\/\/[\w|\.|\S]*) \((?P<response_time>\d+)ms (?P<bytes>\d+) body\)`)

	responseCount := make(map[string]*UrlValue)

	for scanner.Scan() {
		text := scanner.Text()
		result := getResult(text)
		if len(result) != 0 {
			urlVal, ok := responseCount[result["url"]]
			if ok == true {
				responseTime, _ := strconv.Atoi(result["response_time"])
				bytes, _ := strconv.Atoi(result["bytes"])
				urlVal.totalByte += bytes
				urlVal.count += 1
				urlVal.totalTime += responseTime
			} else {
				urlVal = new(UrlValue)
				responseTime, _ := strconv.Atoi(result["response_time"])
				bytes, _ := strconv.Atoi(result["bytes"])
				urlVal.totalByte = bytes
				urlVal.count = 1
				urlVal.totalTime = responseTime
				responseCount[result["url"]] = urlVal
			}

		}
	}
	/*for k, v := range responseCount {
		fmt.Printf("%s %d bytes %dms %d count \n", k, v.totalByte, v.totalTime, v.count)
	}*/

	sortMap(responseCount)
}

func exe_cmd(cmd string, wg *sync.WaitGroup) {
	fmt.Println("command is ",cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
}

func sortMap(urlMap map[string]*UrlValue) {
	pl := make(PairList, len(urlMap))
	i := 0
	for k, v := range urlMap {
		pl[i] = Pair{k, *v}
		i++
	}
	sort.Sort(sort.Reverse(pl))

	file, _ := os.Create("result.csv")
	defer file.Close()

	//writer := csv.NewWriter(file)

	for _, val := range pl {

		fmt.Printf("%10d bytes %10dms %4d count %s\n", val.Value.totalByte, val.Value.totalTime, val.Value.count, val.Key)
	}

	/*for _, value := range pl {
		writer.Write(value)
		//checkError("Cannot write to file", err)
	}

	defer writer.Flush()*/
}

type Pair struct {
	Key string
	Value UrlValue
}

type PairList []Pair
func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value.count < p[j].Value.count }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }



func getResult(text string)  map[string]string {
	result := make(map[string]string)
	match := myExp.FindStringSubmatch(text)
	if myExp.MatchString(text) {
		for i, name := range myExp.SubexpNames() {
			result[name] = match[i]
		}
	}

	return result
}
