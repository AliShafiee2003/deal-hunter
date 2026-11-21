package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Snapp Shop API Constants
const (
	APIBaseURL     = "https://apix.snappshop.ir/search/v1"
	SortByDiscount = "5DzQq4" // بیشترین تخفیف (Highest Discount ID)
	DefaultLat     = "35.77331"
	DefaultLng     = "51.418591"
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Predefined known Category IDs
var PredefinedCategories = map[string]string{
	"mobile":        "AqZkDJ", // گوشی موبایل
	"hard_external": "50a6Zq", // هارد دیسک اکسترنال
	"ssd":           "nqPlzq", // حافظه SSD
	"laptop":        "w02XDY", // لپ تاپ و اولترابوک
	"tablet":        "rg3QqA", // تبلت
	"smartwatch":    "5Db1Lq", // ساعت هوشمند
	"powerbank":     "KqJkXD", // پاوربانک
	"console":       "ngywqR", // کنسول خانگی
	"accessories":   "GDLknq", // لوازم جانبی موبایل
	"storage_acc":   "w02jL0", // تجهیزات ذخیره‌سازی
}

type SearchRequestBody struct {
	Query       string `json:"query,omitempty"`
	Category    string `json:"category,omitempty"`
	Sort        string `json:"sort,omitempty"`
	IsAvailable bool   `json:"is_available"`
	HasDiscount bool   `json:"has_discount"`
	Page        int    `json:"page"`
	Skip        int    `json:"skip,omitempty"`
	Limit       int    `json:"limit"`
}

type APIResponse struct {
	Status bool    `json:"status"`
	Data   APIData `json:"data"`
}

type APIData struct {
	Structure []StructureSection `json:"structure"`
}

type StructureSection struct {
	ID         string     `json:"id"`
	Type       string     `json:"section_type"`
	Items      []Item     `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
	Count       int `json:"count"`
	PerPage     int `json:"per_page"`
}

type Item struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Href      string    `json:"href"`
	PriceInfo PriceInfo `json:"price"`
}

type PriceInfo struct {
	Price           int64 `json:"price"`            // Original price (before discount) in Toman
	DiscountedPrice int64 `json:"discounted_price"` // Actual selling price (after discount) in Toman
	Discount        int   `json:"discount"`
	DiscountPercent int   `json:"discount_percent"`
}

type ProductResult struct {
	Title           string `json:"title"`
	DiscountPercent int    `json:"discount_percent"`
	PriceToman      int64  `json:"price_toman"`
	OriginalToman   int64  `json:"original_toman"`
	ProductURL      string `json:"product_url"`
	Category        string `json:"category"`
}

type Crawler struct {
	client  *http.Client
	lat     string
	lng     string
	minDisc int
}

func NewCrawler(lat, lng string, minDisc int) *Crawler {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
	return &Crawler{
		client:  &http.Client{Transport: tr, Timeout: 12 * time.Second},
		lat:     lat,
		lng:     lng,
		minDisc: minDisc,
	}
}

func (c *Crawler) FetchPage(reqBody SearchRequestBody) (*StructureSection, error) {
	u, err := url.Parse(APIBaseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("lat", c.lat)
	q.Set("lng", c.lng)
	u.RawQuery = q.Encode()

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("User-Agent", UserAgent)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Origin", "https://snappshop.ir")
	httpReq.Header.Set("Referer", "https://snappshop.ir/")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(bodySnippet))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Data.Structure) == 0 {
		return nil, fmt.Errorf("no structure sections returned")
	}

	return &apiResp.Data.Structure[0], nil
}

func (c *Crawler) CrawlTarget(name, catID, query string, maxPages int) []ProductResult {
	var results []ProductResult
	seenURLs := make(map[string]bool)

	for page := 1; page <= maxPages; page++ {
		req := SearchRequestBody{
			Category:    catID,
			Query:       query,
			Sort:        SortByDiscount,
			IsAvailable: true,
			HasDiscount: true,
			Page:        page,
			Skip:        page - 1,
			Limit:       100,
		}

		sec, err := c.FetchPage(req)
		if err != nil {
			break
		}

		if len(sec.Items) == 0 {
			break
		}

		belowThresholdCount := 0
		newItemsCount := 0

		for _, it := range sec.Items {
			if it.Href == "" || seenURLs[it.Href] {
				continue
			}
			seenURLs[it.Href] = true
			newItemsCount++

			disc := it.PriceInfo.Discount
			if disc == 0 && it.PriceInfo.DiscountPercent > 0 {
				disc = it.PriceInfo.DiscountPercent
			}

			if disc < c.minDisc {
				belowThresholdCount++
				continue
			}

			var priceToman int64
			var origToman int64

			if it.PriceInfo.DiscountedPrice > 0 {
				priceToman = it.PriceInfo.DiscountedPrice
				if it.PriceInfo.Price > 0 {
					origToman = it.PriceInfo.Price
				} else if disc > 0 {
					origToman = int64(float64(priceToman) / (1.0 - float64(disc)/100.0))
				}
			} else {
				priceToman = it.PriceInfo.Price
				if disc > 0 {
					origToman = int64(float64(priceToman) / (1.0 - float64(disc)/100.0))
				}
			}

			results = append(results, ProductResult{
				Title:           it.Title,
				DiscountPercent: disc,
				PriceToman:      priceToman,
				OriginalToman:   origToman,
				ProductURL:      "https://snappshop.ir" + it.Href,
				Category:        name,
			})
		}

		if newItemsCount == 0 || (belowThresholdCount > 0 && belowThresholdCount == newItemsCount) {
			break
		}

		if page >= sec.Pagination.TotalPages {
			break
		}
	}

	return results
}

func openBrowser(url string) {
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "start", url).Start()
	}
}

func formatPrice(val int64) string {
	s := fmt.Sprintf("%d", val)
	n := len(s)
	if n <= 3 {
		return s
	}
	var res []string
	remainder := n % 3
	if remainder > 0 {
		res = append(res, s[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		res = append(res, s[i:i+3])
	}
	return strings.Join(res, ",")
}

func saveCSV(filename string, items []ProductResult) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("\xEF\xBB\xBF") // UTF-8 BOM
	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"Category", "Discount (%)", "Title", "Price (Toman)", "Original (Toman)", "URL"})
	for _, p := range items {
		w.Write([]string{
			p.Category,
			strconv.Itoa(p.DiscountPercent),
			p.Title,
			strconv.FormatInt(p.PriceToman, 10),
			strconv.FormatInt(p.OriginalToman, 10),
			p.ProductURL,
		})
	}
	return nil
}

func main() {
	targetPtr := flag.String("target", "all", "Target: mobile, hard, ssd, laptop, accessories, all, or custom query")
	minDiscPtr := flag.Int("min-discount", 50, "Minimum discount percent (e.g. 80)")
	maxPagesPtr := flag.Int("max-pages", 3, "Maximum pages per category (1 page = 100 products)")
	latPtr := flag.String("lat", DefaultLat, "Latitude for delivery location")
	lngPtr := flag.String("lng", DefaultLng, "Longitude for delivery location")
	outputPtr := flag.String("output", "", "Output filename (.json or .csv)")
	monitorPtr := flag.Int("monitor", 0, "Monitor mode: interval in seconds (e.g. 30). 0 = run once")
	autoOpenPtr := flag.Bool("auto-open", false, "Auto open new 80%+ deals in default browser")

	flag.Parse()

	crawler := NewCrawler(*latPtr, *lngPtr, *minDiscPtr)

	// Memory of seen products across monitor ticks to detect NEW deals
	knownDeals := make(map[string]int)
	isFirstRun := true

	runScan := func() {
		fmt.Printf("\n========================================================================================\n")
		fmt.Printf("🚀 Snapp Shop Deal Hunter | Min Discount: %d%% | Target: %s\n", *minDiscPtr, *targetPtr)
		fmt.Printf("⏰ %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("========================================================================================\n\n")

		var targets []struct {
			Name  string
			CatID string
			Query string
		}

		switch *targetPtr {
		case "mobile":
			targets = append(targets, struct{ Name, CatID, Query string }{"گوشی موبایل", PredefinedCategories["mobile"], ""})
		case "hard":
			targets = append(targets, struct{ Name, CatID, Query string }{"هارد اکسترنال", PredefinedCategories["hard_external"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"حافظه SSD", PredefinedCategories["ssd"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"تجهیزات ذخیره‌سازی", PredefinedCategories["storage_acc"], ""})
		case "ssd":
			targets = append(targets, struct{ Name, CatID, Query string }{"حافظه SSD", PredefinedCategories["ssd"], ""})
		case "accessories":
			targets = append(targets, struct{ Name, CatID, Query string }{"لوازم جانبی موبایل", PredefinedCategories["accessories"], ""})
		case "laptop":
			targets = append(targets, struct{ Name, CatID, Query string }{"لپ تاپ", PredefinedCategories["laptop"], ""})
		case "all":
			targets = append(targets, struct{ Name, CatID, Query string }{"گوشی موبایل", PredefinedCategories["mobile"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"هارد اکسترنال", PredefinedCategories["hard_external"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"حافظه SSD", PredefinedCategories["ssd"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"لپ تاپ", PredefinedCategories["laptop"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"ساعت هوشمند", PredefinedCategories["smartwatch"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"پاوربانک", PredefinedCategories["powerbank"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"لوازم جانبی موبایل", PredefinedCategories["accessories"], ""})
			targets = append(targets, struct{ Name, CatID, Query string }{"کل فروشگاه (بیشترین تخفیف)", "", ""})
		default:
			targets = append(targets, struct{ Name, CatID, Query string }{"جستجو: " + *targetPtr, "", *targetPtr})
		}

		var allResults []ProductResult
		var mu sync.Mutex
		var wg sync.WaitGroup

		start := time.Now()

		for _, t := range targets {
			wg.Add(1)
			go func(name, catID, query string) {
				defer wg.Done()
				res := crawler.CrawlTarget(name, catID, query, *maxPagesPtr)
				mu.Lock()
				allResults = append(allResults, res...)
				mu.Unlock()
			}(t.Name, t.CatID, t.Query)
		}

		wg.Wait()
		elapsed := time.Since(start)

		// Deduplicate across targets
		uniqueResults := make([]ProductResult, 0, len(allResults))
		seenGlobal := make(map[string]bool)
		for _, r := range allResults {
			if !seenGlobal[r.ProductURL] {
				seenGlobal[r.ProductURL] = true
				uniqueResults = append(uniqueResults, r)
			}
		}

		// Sort by discount percent descending
		sort.Slice(uniqueResults, func(i, j int) bool {
			if uniqueResults[i].DiscountPercent == uniqueResults[j].DiscountPercent {
				return uniqueResults[i].PriceToman < uniqueResults[j].PriceToman
			}
			return uniqueResults[i].DiscountPercent > uniqueResults[j].DiscountPercent
		})

		// Check for newly dropped deals if in monitor mode
		var newlyAdded []ProductResult
		for _, r := range uniqueResults {
			prevDisc, exists := knownDeals[r.ProductURL]
			if !exists || r.DiscountPercent > prevDisc {
				if !isFirstRun {
					newlyAdded = append(newlyAdded, r)
				}
				knownDeals[r.ProductURL] = r.DiscountPercent
			}
		}

		if isFirstRun {
			fmt.Printf("✅ Scanned %d categories in %v! Found %d deals matching >= %d%% discount.\n\n",
				len(targets), elapsed, len(uniqueResults), *minDiscPtr)

			limitShow := len(uniqueResults)
			if limitShow > 30 {
				limitShow = 30
			}

			for i := 0; i < limitShow; i++ {
				p := uniqueResults[i]
				badge := "🔥"
				if p.DiscountPercent >= 80 {
					badge = "💥💥 [SUPER DEAL 80%+]"
				} else if p.DiscountPercent >= 60 {
					badge = "⚡ [HOT 60%+]"
				}

				fmt.Printf("[%2d] %s %d%% OFF | %s | %s تومان\n",
					i+1, badge, p.DiscountPercent, p.Title, formatPrice(p.PriceToman))
				fmt.Printf("     دسته: %s | %s\n\n", p.Category, p.ProductURL)
			}

			if len(uniqueResults) > limitShow {
				fmt.Printf("... و %d مورد دیگر.\n", len(uniqueResults)-limitShow)
			}

			isFirstRun = false
		} else {
			// Monitoring subsequent checks
			if len(newlyAdded) > 0 {
				fmt.Print("\a") // Windows console alert beep!
				fmt.Printf("🚨🚨 [%d] مورد تخفیف جدید پیدا شد! 🚨🚨\n", len(newlyAdded))
				for _, p := range newlyAdded {
					fmt.Printf("🔔 [NEW %d%% OFF] %s | %s تومان\n   %s\n\n",
						p.DiscountPercent, p.Title, formatPrice(p.PriceToman), p.ProductURL)
					if *autoOpenPtr && p.DiscountPercent >= 80 {
						openBrowser(p.ProductURL)
					}
				}
			} else {
				fmt.Printf("💤 بررسی در %v انجام شد — آیتم جدیدی به تخفیف‌های بالای %d%% اضافه نشده است.\n",
					elapsed, *minDiscPtr)
			}
		}

		// Save output if flag is given
		if *outputPtr != "" && len(uniqueResults) > 0 {
			if strings.HasSuffix(*outputPtr, ".csv") {
				saveCSV(*outputPtr, uniqueResults)
			} else {
				fileData, _ := json.MarshalIndent(uniqueResults, "", "  ")
				os.WriteFile(*outputPtr, fileData, 0644)
			}
		}
	}

	runScan()

	if *monitorPtr > 0 {
		fmt.Printf("\n⏱️ حالت پایش زنده فعال است: هر %d ثانیه بررسی می‌شود. برای توقف Ctrl+C بزنید...\n", *monitorPtr)
		ticker := time.NewTicker(time.Duration(*monitorPtr) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			runScan()
		}
	}
}
