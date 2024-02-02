package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Fuzzer struct {
	HedefURL     string
	Wordlist     []string
	Eşzamanlılık int
}

func main() {
	hedefURL := flag.String("url", "", "Hedef URL'yi belirtir")
	wordlistYolu := flag.String("wordlist", "common.txt", "Wordlist dosyasının yolunu belirtir")
	eşzamanlılık := flag.Int("eşzamanlılık", 10, "Eşzamanlı çalışan işçi sayısını belirtir")
	help := flag.Bool("help", false, "Kullanım bilgilerini gösterir")

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	if *hedefURL == "" {
		fmt.Println("Hedef URL gereklidir.")
		return
	}

	wordlist, err := wordlistiOku(*wordlistYolu)
	if err != nil {
		fmt.Println("Wordlist okuma hatası:", err)
		return
	}

	fuzzer := Fuzzer{
		HedefURL:     *hedefURL,
		Wordlist:     wordlist,
		Eşzamanlılık: *eşzamanlılık,
	}

	fuzzer.Calistir()
}

func (f *Fuzzer) Calistir() {
	fmt.Printf("%s hedefi üzerinde %d işçi ile fuzzing yapılıyor\n", f.HedefURL, f.Eşzamanlılık)

	tarama := make(chan string, len(f.Wordlist))
	sonuclar := make(chan string, len(f.Wordlist))

	var wg sync.WaitGroup

	for i := 0; i < f.Eşzamanlılık; i++ {
		wg.Add(1)
		go f.isci(&wg, tarama, sonuclar)
	}

	for _, kelime := range f.Wordlist {
		tarama <- kelime
	}
	close(tarama)

	go func() {
		wg.Wait()
		close(sonuclar)
	}()

	for sonuc := range sonuclar {
		fmt.Println(sonuc)
	}
}

func (f *Fuzzer) isci(wg *sync.WaitGroup, isler <-chan string, sonuclar chan<- string) {
	defer wg.Done()

	for is := range isler {
		url := strings.Replace(f.HedefURL, "FUZZ", is, 1)

		resp, err := http.Get(url)
		if err != nil {
			sonuclar <- fmt.Sprintf("%s adresine ulaşılırken hata oluştu: %v", url, err)
		} else {
			sonuclar <- fmt.Sprintf("[%d] %s", resp.StatusCode, url)
			resp.Body.Close()
		}
	}
}

func wordlistiOku(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		x := fmt.Errorf("dosya açılırken hata oluştu (%s): %v", path, err)
		return nil, x
	}
	defer file.Close()

	var wordlist []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wordlist = append(wordlist, scanner.Text())
	}

	return wordlist, nil
}
