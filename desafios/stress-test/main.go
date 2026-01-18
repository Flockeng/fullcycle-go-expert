package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "", "URL do serviço a ser testado")
	requests := flag.Int("requests", 0, "Número total de requests")
	concurrency := flag.Int("concurrency", 1, "Número de chamadas simultâneas")
	flag.Parse()

	if *url == "" || *requests <= 0 || *concurrency <= 0 {
		fmt.Fprintf(os.Stderr, "Uso: --url=<url> --requests=<n> --concurrency=<m>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	start := time.Now()

	jobs := make(chan int, *requests)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 15 * time.Second}

	var totalDone int64
	var ok200 int64
	var errs int64
	statusCounts := make(map[int]int64)
	var scMu sync.Mutex

	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				req, err := http.NewRequest("GET", *url, nil)
				if err != nil {
					atomic.AddInt64(&errs, 1)
					atomic.AddInt64(&totalDone, 1)
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&errs, 1)
					atomic.AddInt64(&totalDone, 1)
					continue
				}

				scMu.Lock()
				statusCounts[resp.StatusCode]++
				scMu.Unlock()

				if resp.StatusCode == 200 {
					atomic.AddInt64(&ok200, 1)
				}
				resp.Body.Close()
				atomic.AddInt64(&totalDone, 1)
			}
		}()
	}

	for i := 0; i < *requests; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Println("\n--- Relatório ---")
	fmt.Printf("Tempo total gasto: %s\n", elapsed)
	fmt.Printf("Quantidade total de requests realizados: %d\n", totalDone)
	fmt.Printf("Quantidade de requests com status HTTP 200: %d\n", ok200)

	fmt.Println("Distribuição de códigos de status HTTP:")

	var codes []int
	scMu.Lock()
	for k := range statusCounts {
		codes = append(codes, k)
	}
	scMu.Unlock()
	sort.Ints(codes)
	for _, code := range codes {
		scMu.Lock()
		cnt := statusCounts[code]
		scMu.Unlock()
		fmt.Printf("  %d: %d\n", code, cnt)
	}

	if errs > 0 {
		fmt.Printf("Erros de requisição (conexões/falhas): %d\n", errs)
	}

	fmt.Println("--- Fim do relatório ---")
	_ = elapsed
}
