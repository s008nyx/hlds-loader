package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"io"
)

type Item struct {
    Name string `json:"name"`
    Type string `json:"type"`
    Mtime string `json:"mtime"`
    Size uint64 `json:"size,omitempty"`
}

func main() {
	fmt.Println("----------------------------------------------------")
	fmt.Println("------------| (Re)HLDS Installer v0.4 |-------------")
	fmt.Println("----------------------------------------------------")

	routes := make(map[int]string)
	
	depth := 0
	routes[depth] = "http://dl.rehlds.ru"

	do(routes, Item{Name:"",Type:"",Mtime: ""}, depth)
}


func do(routes map[int]string, item Item, depth int) {
	if isFile(item) {
		err := load(routes, item, depth)
		if err == nil {
			if depth <= 1 {
				do(routes, Item{Name:"",Type:"",Mtime: ""}, 0)
			} else {
				do(routes, Item{Name:"",Type:"",Mtime: ""}, depth - 1)
			}
		}
	} else {
		list(routes, item, depth)
	}
}

func getUrl(routes map[int]string, item Item, depth int) string {
	if len(item.Name) > 0 {
		routes[depth] = routes[depth - 1] + "/" + item.Name
		return routes[depth]
	}

	return routes[depth]
}

func isFile(item Item) bool {
	return item.Type == "file"
}

func list(routes map[int]string, item Item, depth int) {
	res, err := http.Get(getUrl(routes, item, depth))
	if err != nil {
		panic(err)
	}

	var data []Item

	body, _ := ioutil.ReadAll(res.Body)

	err = json.Unmarshal([]byte(body), &data)

	if err != nil {
		panic(err)
	}

	fmt.Println("Choose item (enter number):")
	fmt.Printf("--------------------------------\n")
    for i, item := range data{
       if isFile(item) {
       	   fmt.Printf("%-3d | %-110s | %-9s |\n", i + 1, item.Name, ByteCountDecimal(item.Size))
       } else {
       	   fmt.Printf("%-3d | %-110s |    ---    |\n", i + 1, item.Name)
       }
    }

 	fmt.Printf("--------------------------------\n")
 	if depth <= 0 {
    	fmt.Printf("0: Exit\n")
	} else {
		fmt.Printf("0: Back\n")
	}

    var point int
	fmt.Scanf("%d\n", &point)

	if point == 0 {
		if depth <= 0 {
			os.Exit(0)
		} else {
			do(routes, Item{Name:"",Type:"",Mtime: ""}, depth - 1)
		}
		
	}

	if point - 1 <= len(data) {
		do(routes, data[point - 1], depth + 1)
	} 

	do(routes, item, depth)
}

type WriteCounter struct {
	All uint64
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	//fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %d%% (%d/%d) complete", (100 * wc.Total / wc.All), wc.Total, wc.All)
}

func load(routes map[int]string, item Item, depth int) error {
	out, err := os.Create(item.Name + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(getUrl(routes, item, depth))
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{All: item.Size}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	fmt.Print("\n")
	out.Close()

	if err = os.Rename(item.Name + ".tmp", item.Name); err != nil {
		return err
	}

	return nil
}

func ByteCountDecimal(b uint64) string {
    const unit = 1000
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := int64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
