package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Url struct {
	Url             string
	Alias           string
	Selector        string
	Date            time.Time
	Hash            string
	Disable_Message bool
}

type Setting struct {
	Links    []Url
	Telegram struct {
		Token  string
		ChatId string
	}
}

func setting(s *Setting) error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(s)

	return err
}

func loadLinks(setting *Setting) error {
	fmt.Print("get", time.Now())
	var wg sync.WaitGroup
	timer := make([]time.Duration, len(setting.Links))
	for i := range setting.Links {
		wg.Add(1)
		go (func(i int) {
			defer wg.Done()
			startTime := time.Now()
			err := setting.LoadLink(&setting.Links[i])
			if err != nil {
				fmt.Println(err)
			}
			timer[i] = time.Since(startTime)
		})(i)
		wg.Wait()
	}
	fmt.Println(timer)
	return nil
}

func saveCache(setting *Setting) error {
	f, err := os.OpenFile("config.yaml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(setting)
	if err != nil {
		return err
	}
	err = encoder.Close()
	return err
}

func main() {
	s := Setting{}

	for {
		err := setting(&s)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		err = loadLinks(&s)
		if err != nil {
			fmt.Println(err)
		}
		err = saveCache(&s)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		time.Sleep(25 * time.Second)
	}
}
