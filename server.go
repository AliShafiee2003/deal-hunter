package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SnappAPIBaseURL = "https://apix.snappshop.ir/search/v1"
	SnappSortDisc   = "5DzQq4"
	DefaultLat      = "35.77331"
	DefaultLng      = "51.418591"
	UserAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type CrawlTarget struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	SnappCat   string `json:"snapp_cat,omitempty"`
	SnappQuery string `json:"snapp_query,omitempty"`
	DkCatID    int    `json:"dk_cat_id,omitempty"`
	DkQuery    string `json:"dk_query,omitempty"`
}

// Unified 27-category hierarchy covering both SnappShop & Digikala (Category IDs + Keyword searches)
var TargetConfigs = []CrawlTarget{
	// ۱. هارد اکسترنال و حافظه
	{Key: "hard", Name: "هارد اکسترنال", SnappCat: "50a6Zq", DkCatID: 68, DkQuery: "هارد اکسترنال"},
	{Key: "hard", Name: "هارد اکسترنال", SnappQuery: "هارد اکسترنال", DkCatID: 11140, DkQuery: "هارد دیسک اکسترنال"},
	{Key: "ssd", Name: "حافظه SSD", SnappCat: "nqPlzq", DkCatID: 69, DkQuery: "اس اس دی"},
	{Key: "ssd", Name: "حافظه SSD", SnappQuery: "اس اس دی", DkQuery: "حافظه ssd"},
	{Key: "ssd", Name: "حافظه SSD", SnappCat: "5gv2Bg", DkQuery: "فلش مموری"},

	// ۲. موبایل و تبلت
	{Key: "mobile", Name: "گوشی موبایل", SnappCat: "AqZkDJ", DkCatID: 11, DkQuery: "گوشی موبایل"},
	{Key: "mobile", Name: "گوشی موبایل", SnappQuery: "گوشی موبایل", DkQuery: "گوشی سامسونگ"},
	{Key: "mobile", Name: "گوشی موبایل", SnappQuery: "گوشی شیائومی", DkQuery: "گوشی شیائومی"},
	{Key: "mobile", Name: "گوشی موبایل", DkQuery: "گوشی آیفون"},
	{Key: "mobile", Name: "گوشی موبایل", SnappQuery: "قاب و گلس گوشی", DkQuery: "لوازم جانبی موبایل"},
	{Key: "tablet", Name: "تبلت و آیپد", SnappCat: "rg3QqA", DkCatID: 13, DkQuery: "تبلت"},
	{Key: "tablet", Name: "تبلت و آیپد", SnappQuery: "تبلت", DkQuery: "کیف و کاور تبلت"},

	// ۳. لپ‌تاپ، کامپیوتر و گیمینگ
	{Key: "laptop", Name: "لپ‌تاپ و کامپیوتر", SnappCat: "w02XDY", DkCatID: 18, DkQuery: "لپ تاپ"},
	{Key: "laptop", Name: "لپ‌تاپ و کامپیوتر", SnappQuery: "لپ تاپ", DkQuery: "مانیتور"},
	{Key: "laptop", Name: "لپ‌تاپ و کامپیوتر", SnappCat: "EqO91g", DkQuery: "کیبورد و ماوس"},
	{Key: "laptop", Name: "لپ‌تاپ و کامپیوتر", SnappCat: "nqPlRq", DkQuery: "قطعات کامپیوتر"},
	{Key: "gaming", Name: "کنسول و گیمینگ", SnappCat: "ngywqR", DkCatID: 216, DkQuery: "کنسول بازی"},
	{Key: "gaming", Name: "کنسول و گیمینگ", SnappCat: "ygGWgW", DkQuery: "دسته بازی"},
	{Key: "gaming", Name: "کنسول و گیمینگ", SnappQuery: "گیمینگ", DkQuery: "لوازم گیمینگ"},

	// ۴. صوتی، تصویری و گجت
	{Key: "headphone", Name: "هدفون و هندزفری", SnappQuery: "هدفون", DkCatID: 211, DkQuery: "هدفون"},
	{Key: "headphone", Name: "هدفون و هندزفری", SnappQuery: "هندزفری", DkQuery: "هندزفری بلوتوثی"},
	{Key: "headphone", Name: "هدفون و هندزفری", DkQuery: "هدست"},
	{Key: "speaker", Name: "اسپیکر و صوتی", SnappCat: "AqZPkg", DkCatID: 25, DkQuery: "اسپیکر"},
	{Key: "speaker", Name: "اسپیکر و صوتی", SnappQuery: "اسپیکر", DkQuery: "اسپیکر بلوتوثی"},
	{Key: "speaker", Name: "اسپیکر و صوتی", DkQuery: "ساندبار و میکروفون"},
	{Key: "camera", Name: "دوربین عکاسی", SnappCat: "Xqnm0b", DkCatID: 48, DkQuery: "دوربین عکاسی"},
	{Key: "camera", Name: "دوربین عکاسی", SnappQuery: "دوربین", DkQuery: "تجهیزات عکاسی"},
	{Key: "gadgets", Name: "ساعت هوشمند", SnappCat: "5Db1Lq", DkCatID: 5848, DkQuery: "ساعت هوشمند"},
	{Key: "gadgets", Name: "ساعت هوشمند", SnappQuery: "ساعت هوشمند", DkQuery: "مچ بند هوشمند"},
	{Key: "gadgets", Name: "ساعت هوشمند", SnappCat: "x0Qdog", DkCatID: 6944, DkQuery: "بند ساعت هوشمند"},
	{Key: "powerbank", Name: "پاوربانک", SnappCat: "KqJkXD", DkCatID: 1272, DkQuery: "پاوربانک"},
	{Key: "powerbank", Name: "پاوربانک", SnappQuery: "پاوربانک", DkQuery: "شارژر همراه"},
	{Key: "powerbank", Name: "پاوربانک", SnappCat: "KgN24g", DkQuery: "شارژر و کابل شارژ"},

	// ۵. لوازم خانگی و آشپزخانه
	{Key: "home", Name: "لوازم خانگی", SnappQuery: "لوازم خانگی", DkCatID: 5874, DkQuery: "لوازم خانگی برقی"},
	{Key: "home", Name: "لوازم خانگی", SnappQuery: "مبل", DkCatID: 6357, DkQuery: "تلویزیون"},
	{Key: "home", Name: "لوازم خانگی", SnappCat: "AqZpAg", DkQuery: "فرش و پتو و دکوراسیون"},
	{Key: "home", Name: "لوازم خانگی", DkQuery: "پنکه و کولر و بخاری"},
	{Key: "kitchen", Name: "خانه و آشپزخانه", SnappQuery: "لوازم آشپزخانه", DkCatID: 6346, DkQuery: "خانه آشپزخانه"},
	{Key: "kitchen", Name: "خانه و آشپزخانه", DkCatID: 6247, DkQuery: "ظروف پخت و پز"},
	{Key: "kitchen", Name: "خانه و آشپزخانه", DkQuery: "سرویس قابلمه و تابه"},
	{Key: "kitchen", Name: "خانه و آشپزخانه", DkQuery: "سرخ کن و پلوپز و چای ساز"},
	{Key: "fridge", Name: "یخچال و فریزر", SnappCat: "70KkK0", DkCatID: 5803, DkQuery: "یخچال"},
	{Key: "fridge", Name: "یخچال و فریزر", DkQuery: "فریزر"},
	{Key: "vacuum", Name: "جاروبرقی", SnappCat: "5DjPX0", DkCatID: 5773, DkQuery: "جارو برقی"},
	{Key: "vacuum", Name: "جاروبرقی", DkCatID: 11329, DkQuery: "جارو شارژی و رباتیک"},
	{Key: "washing", Name: "ماشین لباسشویی", SnappCat: "KgN6eg", DkCatID: 5804, DkQuery: "ماشین لباسشویی"},
	{Key: "washing", Name: "ماشین لباسشویی", DkQuery: "ماشین ظرفشویی"},
	{Key: "coffee", Name: "قهوه‌ساز و تمپر", SnappCat: "ODYY3D", DkCatID: 6232, DkQuery: "قهوه ساز"},
	{Key: "coffee", Name: "قهوه‌ساز و تمپر", SnappQuery: "قهوه ساز", DkCatID: 5929, DkQuery: "تمپر قهوه"},
	{Key: "coffee", Name: "قهوه‌ساز و تمپر", SnappCat: "50aVeq", DkQuery: "اسپرسوساز و موکاپات"},

	// ۶. زیبایی، سلامت و مکمل
	{Key: "beauty", Name: "زیبایی و سلامت", SnappQuery: "آرایشی بهداشتی", DkCatID: 5865, DkQuery: "آرایشی بهداشتی"},
	{Key: "beauty", Name: "زیبایی و سلامت", SnappQuery: "کرم", DkCatID: 6484, DkQuery: "کرم مرطوب کننده و ضد آفتاب"},
	{Key: "beauty", Name: "زیبایی و سلامت", SnappQuery: "عطر ادکلن", DkQuery: "عطر و ادکلن"},
	{Key: "beauty", Name: "زیبایی و سلامت", SnappCat: "J0W6Rg", DkCatID: 6085, DkQuery: "سشوار و ریش تراش و شامپو"},
	{Key: "health", Name: "سلامت و ورزش", SnappQuery: "مکمل", DkCatID: 10120, DkQuery: "مکمل سلامت"},
	{Key: "health", Name: "سلامت و ورزش", DkCatID: 8637, DkQuery: "فشارسنج و تجهیزات پزشکی"},
	{Key: "eyewear", Name: "عینک طبی", SnappCat: "XqnlYg", DkCatID: 11123, DkQuery: "عینک طبی"},
	{Key: "eyewear", Name: "عینک طبی", SnappQuery: "عینک", DkCatID: 9383, DkQuery: "عینک آفتابی"},
	{Key: "supplement", Name: "کپسول و مکمل", SnappQuery: "کپسول مکمل", DkCatID: 11211, DkQuery: "کپسول مکمل"},
	{Key: "supplement", Name: "کپسول و مکمل", DkCatID: 11197, DkQuery: "قرص ویتامین و پروتئین"},

	// ۷. مد، پوشاک، کفش و طلا
	{Key: "fashion", Name: "مد و پوشاک", SnappQuery: "لباس", DkCatID: 9345, DkQuery: "پوشاک"},
	{Key: "fashion", Name: "مد و پوشاک", SnappCat: "EqOv10", DkCatID: 9470, DkQuery: "تیشرت و پیراهن و شلوار"},
	{Key: "fashion", Name: "مد و پوشاک", DkCatID: 9326, DkQuery: "مانتو و هودی و لباس زنانه"},
	{Key: "fashion", Name: "مد و پوشاک", DkQuery: "کیف و کوله پشتی و ساعت عقربه ای"},
	{Key: "shoes", Name: "کفش", SnappCat: "LDxyQg", DkCatID: 9486, DkQuery: "کفش"},
	{Key: "shoes", Name: "کفش", SnappCat: "ODYnNg", DkCatID: 9477, DkQuery: "کفش ورزشی و کتانی"},
	{Key: "shoes", Name: "کفش", SnappQuery: "کفش", DkQuery: "دمپایی و صندل و بوت"},
	{Key: "gold", Name: "طلا و نقره", SnappCat: "lgwVBD", DkCatID: 9393, DkQuery: "طلا"},
	{Key: "gold", Name: "طلا و نقره", SnappQuery: "طلا", DkCatID: 9394, DkQuery: "گردنبند و انگشتر و دستبند نقره"},

	// ۸. خودرو، ابزار، سوپرمارکت، کتاب، ورزش و کودک
	{Key: "tools_auto", Name: "خودرو و ابزارآلات", SnappQuery: "ابزار", DkCatID: 10082, DkQuery: "ابزار خودرو"},
	{Key: "tools_auto", Name: "خودرو و ابزارآلات", SnappQuery: "خودرو", DkCatID: 8369, DkQuery: "لوازم جانبی خودرو"},
	{Key: "tools_auto", Name: "خودرو و ابزارآلات", SnappCat: "rg3r8D", DkCatID: 6165, DkQuery: "دریل و پیچ گوشتی و آچار"},
	{Key: "supermarket", Name: "سوپرمارکت", SnappQuery: "سوپرمارکت", DkQuery: "سوپرمارکت"},
	{Key: "supermarket", Name: "سوپرمارکت", DkQuery: "قهوه و چای و شکلات و تنقلات"},
	{Key: "supermarket", Name: "سوپرمارکت", DkQuery: "روغن و برنج و مواد غذایی و شوینده"},
	{Key: "book_art", Name: "کتاب و هنر", SnappQuery: "کتاب", DkCatID: 10557, DkQuery: "کتاب"},
	{Key: "book_art", Name: "کتاب و هنر", SnappCat: "x0Qaa0", DkCatID: 6014, DkQuery: "لوازم التحریر و نوشت افزار"},
	{Key: "sport", Name: "ورزش و سفر", SnappQuery: "ورزشی", DkCatID: 9428, DkQuery: "وسایل ورزشی"},
	{Key: "sport", Name: "ورزش و سفر", SnappCat: "DzbpwD", DkQuery: "چادر مسافرتی و کمپینگ و دمبل"},
	{Key: "toy", Name: "اسباب‌بازی و کودک", SnappQuery: "اسباب بازی", DkCatID: 6029, DkQuery: "اسباب بازی"},
	{Key: "toy", Name: "اسباب‌بازی و کودک", SnappCat: "kqE9Qq", DkCatID: 6030, DkQuery: "بازی فکری و لگو"},
	{Key: "all", Name: "کل فروشگاه (تخفیف‌های ویژه)", SnappQuery: "", DkQuery: "ALL_DISCOUNTS"},
}

// --- Snapp Shop API Structures ---
type SnappSearchBody struct {
	Query       string `json:"query,omitempty"`
	Category    string `json:"category,omitempty"`
	Sort        string `json:"sort,omitempty"`
	IsAvailable bool   `json:"is_available"`
	HasDiscount bool   `json:"has_discount"`
	Page        int    `json:"page"`
	Skip        int    `json:"skip"` // site uses skip = pageIndex (0-based); "page" is ignored by the API
	Limit       int    `json:"limit"`
}

type SnappAPIResponse struct {
	Status bool         `json:"status"`
	Data   SnappAPIData `json:"data"`
}

type SnappAPIData struct {
	Structure []SnappStructureSection `json:"structure"`
}

type SnappStructureSection struct {
	ID    string         `json:"id"`
	Type  string         `json:"section_type"`
	Items []SnappRawItem `json:"items"`
}

type SnappRawItem struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Href      string         `json:"href"`
	PriceInfo SnappPriceInfo `json:"price"`
	Image     struct {
		Src string `json:"src"`
		Alt string `json:"alt"`
	} `json:"image"`
}

type SnappPriceInfo struct {
	Price           int64 `json:"price"`            // Original price in Toman
	DiscountedPrice int64 `json:"discounted_price"` // Final price in Toman
	Discount        int   `json:"discount"`
	DiscountPercent int   `json:"discount_percent"`
}

// --- Digikala API Structures ---
type DkResponse struct {
	Status int    `json:"status"`
	Data   DkData `json:"data"`
}

type DkData struct {
	Widgets []DkWidget `json:"widgets"`
}

type DkWidget struct {
	Type string       `json:"type"`
	Data DkWidgetData `json:"data"`
}

type DkWidgetData struct {
	Widgets []DkProductItem `json:"widgets"`
}

type DkProductItem struct {
	Type string        `json:"type"`
	Data DkProductData `json:"data"`
}

type DkProductData struct {
	ID             int64  `json:"id"`
	TitleFa        string `json:"title_fa"`
	DataLayer      struct {
		Category      string `json:"category"`
		ItemCategory2 string `json:"item_category2"`
		ItemCategory3 string `json:"item_category3"`
		ItemCategory4 string `json:"item_category4"`
	} `json:"data_layer"`
	DefaultVariant struct {
		Price struct {
			SellingPrice    int64 `json:"selling_price"`
			RrpPrice        int64 `json:"rrp_price"`
			DiscountPercent int   `json:"discount_percent"`
		} `json:"price"`
	} `json:"default_variant"`
	Images struct {
		Main struct {
			URL []string `json:"url"`
		} `json:"main"`
	} `json:"images"`
	URL struct {
		URI string `json:"uri"`
	} `json:"url"`
}

// --- Unified Deal Item ---
type DealItem struct {
	ID              string    `json:"id"`
	Store           string    `json:"store"`      // "snapp" or "digikala"
	StoreName       string    `json:"store_name"` // "اسنپ‌شاپ" or "دیجی‌کالا"
	Title           string    `json:"title"`
	CategoryKey     string    `json:"category_key"`
	CategoryName    string    `json:"category_name"`
	DiscountPercent int       `json:"discount_percent"`
	PriceToman      int64     `json:"price_toman"`
	OriginalToman   int64     `json:"original_toman"`
	PriceFormatted  string    `json:"price_formatted"`
	OrigFormatted   string    `json:"orig_formatted"`
	ProductURL      string    `json:"product_url"`
	FastBuyURL      string    `json:"fast_buy_url"`
	ImageURL        string    `json:"image_url"`
	FoundAt         time.Time `json:"found_at"`
	FoundTimeStr    string    `json:"found_time_str"`
}

type ServerState struct {
	mu               sync.RWMutex
	deals            map[string]DealItem
	lastScanTime     time.Time
	scanCount        int
	intervalSec      int
	isScanning       bool
	lat              string
	lng              string
	snappClient      *http.Client
	snappReqs        int64
	snapp429         int64
	dkClient         *http.Client
	s26SniperActive  bool
	s26LastScanTime  time.Time
	s26ScanCount     int64
	s26FoundDeals    []DealItem
	s26TriggeredURLs map[string]bool
	s26MaxPriceToman int64
}

func NewServerState(intervalSec int, lat, lng string) *ServerState {
	tr := &http.Transport{
		MaxIdleConns:        120,
		MaxIdleConnsPerHost: 30,
		IdleConnTimeout:     90 * time.Second,
	}

	dkJar, _ := cookiejar.New(nil)
	dkTr := &http.Transport{
		MaxIdleConns:        120,
		MaxIdleConnsPerHost: 30,
		IdleConnTimeout:     90 * time.Second,
	}

	return &ServerState{
		deals:            make(map[string]DealItem),
		intervalSec:      intervalSec,
		lat:              lat,
		lng:              lng,
		snappClient:      &http.Client{Transport: tr, Timeout: 12 * time.Second},
		dkClient:         &http.Client{Transport: dkTr, Jar: dkJar, Timeout: 12 * time.Second},
		s26SniperActive:  true,
		s26FoundDeals:    make([]DealItem, 0),
		s26TriggeredURLs: make(map[string]bool),
		s26MaxPriceToman: 30000000,
	}
}

// --- Fetch from Snapp Shop ---
var snappSem = make(chan struct{}, 4)
var dkSem = make(chan struct{}, 4)

// FetchSnappTarget crawls multiple pages (sorted by discount desc) until discounts drop low.
func (s *ServerState) FetchSnappTarget(target CrawlTarget) ([]DealItem, error) {
	maxPages := 5
	var all []DealItem
	for page := 1; page <= maxPages; page++ {
		items, err := s.fetchSnappPage(target, page)
		if err != nil {
			if len(all) > 0 {
				return all, nil
			}
			return nil, err
		}
		all = append(all, items...)
		if len(items) < 50 {
			break
		}
	}
	return all, nil
}

var snappGate sync.Mutex
var snappLast time.Time

// snappThrottle enforces a minimum gap between Snapp requests (API rate-limits bursts).
func snappThrottle() {
	snappGate.Lock()
	defer snappGate.Unlock()
	if wait := 700*time.Millisecond - time.Since(snappLast); wait > 0 {
		time.Sleep(wait)
	}
	snappLast = time.Now()
}

func (s *ServerState) fetchSnappPage(target CrawlTarget, page int) ([]DealItem, error) {
	var items []DealItem
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		snappThrottle()
		items, err = s.fetchSnappPageOnce(target, page)
		if err == nil || !strings.Contains(err.Error(), "status 429") {
			return items, err
		}
		time.Sleep(time.Duration(4*(attempt+1)) * time.Second)
	}
	return items, err
}

func (s *ServerState) fetchSnappPageOnce(target CrawlTarget, page int) ([]DealItem, error) {
	u, err := url.Parse(SnappAPIBaseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("lat", s.lat)
	q.Set("lng", s.lng)
	u.RawQuery = q.Encode()

	reqBody := SnappSearchBody{
		Category:    target.SnappCat,
		Query:       target.SnappQuery,
		Sort:        SnappSortDisc,
		IsAvailable: true,
		HasDiscount: true,
		Page:        page,
		Skip:        page - 1,
		Limit:       100,
	}

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

	resp, err := s.snappClient.Do(httpReq)
	atomic.AddInt64(&s.snappReqs, 1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		atomic.AddInt64(&s.snapp429, 1)
	}
	if resp.StatusCode != http.StatusOK {
		bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return nil, fmt.Errorf("snapp status %d: %s", resp.StatusCode, string(bodySnippet))
	}

	var apiResp SnappAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Data.Structure) == 0 {
		return nil, nil
	}

	items := apiResp.Data.Structure[0].Items
	var deals []DealItem
	now := time.Now()

	for _, it := range items {
		if it.Href == "" {
			continue
		}

		disc := it.PriceInfo.Discount
		if disc == 0 && it.PriceInfo.DiscountPercent > 0 {
			disc = it.PriceInfo.DiscountPercent
		}

		if disc <= 0 {
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

		fullURL := "https://snappshop.ir" + it.Href
		fastBuyURL := fullURL + "?quick_buy=1"
		img := it.Image.Src
		if img == "" {
			img = "https://snappshop.ir/assets/images/placeholder.png"
		}

		catKey := target.Key
		catName := target.Name
		if catKey == "all" {
			catKey, catName = MapItemToUnifiedCategory(it.Title, "", "", "")
		}

		deals = append(deals, DealItem{
			ID:              "snp_" + it.ID,
			Store:           "snapp",
			StoreName:       "اسنپ‌شاپ",
			Title:           it.Title,
			CategoryKey:     catKey,
			CategoryName:    catName,
			DiscountPercent: disc,
			PriceToman:      priceToman,
			OriginalToman:   origToman,
			PriceFormatted:  formatNumber(priceToman),
			OrigFormatted:   formatNumber(origToman),
			ProductURL:      fullURL,
			FastBuyURL:      fastBuyURL,
			ImageURL:        img,
			FoundAt:         now,
			FoundTimeStr:    now.Format("15:04:05"),
		})
	}

	return deals, nil
}

// MapItemToUnifiedCategory infers the unified category key & Persian name from product title and metadata.
func MapItemToUnifiedCategory(title, c2, c3, cat string) (string, string) {
	txt := strings.ToLower(title + " " + c2 + " " + c3 + " " + cat)
	switch {
	case strings.Contains(txt, "هارد اکسترنال") || strings.Contains(txt, "هارد دیسک اکسترنال"):
		return "hard", "هارد اکسترنال"
	case strings.Contains(txt, "ssd") || strings.Contains(txt, "اس اس دی") || strings.Contains(txt, "فلش مموری") || strings.Contains(txt, "کارت حافظه"):
		return "ssd", "حافظه SSD"
	case strings.Contains(txt, "پاوربانک") || strings.Contains(txt, "پاور بانک") || strings.Contains(txt, "شارژر همراه"):
		return "powerbank", "پاوربانک"
	case strings.Contains(txt, "ساعت هوشمند") || strings.Contains(txt, "مچ بند هوشمند") || strings.Contains(txt, "مچ‌بند هوشمند") || strings.Contains(txt, "بند ساعت"):
		return "gadgets", "ساعت هوشمند"
	case strings.Contains(txt, "هدفون") || strings.Contains(txt, "هندزفری") || strings.Contains(txt, "هدست") || strings.Contains(txt, "ایرپاد"):
		return "headphone", "هدفون و هندزفری"
	case strings.Contains(txt, "اسپیکر") || strings.Contains(txt, "بلندگو") || strings.Contains(txt, "ساندبار") || strings.Contains(txt, "میکروفون"):
		return "speaker", "اسپیکر و صوتی"
	case strings.Contains(txt, "تبلت") || strings.Contains(txt, "آیپد") || strings.Contains(txt, "ipad"):
		return "tablet", "تبلت و آیپد"
	case strings.Contains(txt, "گوشی") || strings.Contains(txt, "موبایل") || strings.Contains(txt, "آیفون") || strings.Contains(txt, "iphone"):
		return "mobile", "گوشی موبایل"
	case strings.Contains(txt, "لپ تاپ") || strings.Contains(txt, "لپ‌تاپ") || strings.Contains(txt, "کامپیوتر") || strings.Contains(txt, "مانیتور") || strings.Contains(txt, "کیبورد") || strings.Contains(txt, "ماوس") || strings.Contains(txt, "مادربرد"):
		return "laptop", "لپ‌تاپ و کامپیوتر"
	case strings.Contains(txt, "کنسول") || strings.Contains(txt, "پلی استیشن") || strings.Contains(txt, "ایکس باکس") || strings.Contains(txt, "گیمینگ") || strings.Contains(txt, "دسته بازی"):
		return "gaming", "کنسول و گیمینگ"
	case strings.Contains(txt, "دوربین") || strings.Contains(txt, "عکاسی"):
		return "camera", "دوربین عکاسی"
	case strings.Contains(txt, "یخچال") || strings.Contains(txt, "فریزر"):
		return "fridge", "یخچال و فریزر"
	case strings.Contains(txt, "جاروبرقی") || strings.Contains(txt, "جارو برقی") || strings.Contains(txt, "جارو شارژی") || strings.Contains(txt, "بخارشو"):
		return "vacuum", "جاروبرقی"
	case strings.Contains(txt, "لباسشویی") || strings.Contains(txt, "ظرفشویی"):
		return "washing", "ماشین لباسشویی"
	case strings.Contains(txt, "قهوه") || strings.Contains(txt, "اسپرسو") || strings.Contains(txt, "تمپر") || strings.Contains(txt, "موکاپات") || strings.Contains(txt, "فرنچ پرس"):
		return "coffee", "قهوه‌ساز و تمپر"
	case strings.Contains(txt, "عینک"):
		return "eyewear", "عینک طبی"
	case strings.Contains(txt, "کپسول") || strings.Contains(txt, "قرص") || strings.Contains(txt, "مکمل") || strings.Contains(txt, "ویتامین") || strings.Contains(txt, "پروتئین"):
		return "supplement", "کپسول و مکمل"
	case strings.Contains(txt, "کفش") || strings.Contains(txt, "کتانی") || strings.Contains(txt, "صندل") || strings.Contains(txt, "دمپایی") || strings.Contains(txt, "بوت"):
		return "shoes", "کفش"
	case strings.Contains(txt, "طلا") || strings.Contains(txt, "نقره") || strings.Contains(txt, "جواهر") || strings.Contains(txt, "انگشتر") || strings.Contains(txt, "گردنبند") || strings.Contains(txt, "دستبند"):
		return "gold", "طلا و نقره"
	case strings.Contains(txt, "آشپزخانه") || strings.Contains(txt, "قابلمه") || strings.Contains(txt, "تابه") || strings.Contains(txt, "ظروف") || strings.Contains(txt, "چای ساز") || strings.Contains(txt, "سرخ کن") || strings.Contains(txt, "پلوپز"):
		return "kitchen", "خانه و آشپزخانه"
	case strings.Contains(txt, "لوازم خانگی") || strings.Contains(txt, "تلویزیون") || strings.Contains(txt, "مبل") || strings.Contains(txt, "فرش") || strings.Contains(txt, "پتو") || strings.Contains(txt, "کولر") || strings.Contains(txt, "پنکه"):
		return "home", "لوازم خانگی"
	case strings.Contains(txt, "آرایشی") || strings.Contains(txt, "بهداشتی") || strings.Contains(txt, "کرم") || strings.Contains(txt, "شامپو") || strings.Contains(txt, "عطر") || strings.Contains(txt, "ادکلن") || strings.Contains(txt, "رژ") || strings.Contains(txt, "سشوار") || strings.Contains(txt, "ریش تراش"):
		return "beauty", "زیبایی و سلامت"
	case strings.Contains(txt, "پزشکی") || strings.Contains(txt, "سلامت") || strings.Contains(txt, "ارتوپدی") || strings.Contains(txt, "فشارسنج") || strings.Contains(txt, "ماساژور"):
		return "health", "سلامت و ورزش"
	case strings.Contains(txt, "پوشاک") || strings.Contains(txt, "لباس") || strings.Contains(txt, "مانتو") || strings.Contains(txt, "شلوار") || strings.Contains(txt, "پیراهن") || strings.Contains(txt, "تیشرت") || strings.Contains(txt, "تی شرت") || strings.Contains(txt, "هودی") || strings.Contains(txt, "کیف") || strings.Contains(txt, "کوله") || strings.Contains(txt, "جوراب"):
		return "fashion", "مد و پوشاک"
	case strings.Contains(txt, "خودرو") || strings.Contains(txt, "موتورسیکلت") || strings.Contains(txt, "ابزار") || strings.Contains(txt, "دریل") || strings.Contains(txt, "آچار") || strings.Contains(txt, "پیچ گوشتی"):
		return "tools_auto", "خودرو و ابزارآلات"
	case strings.Contains(txt, "سوپرمارکت") || strings.Contains(txt, "غذایی") || strings.Contains(txt, "خوراکی") || strings.Contains(txt, "نوشیدنی") || strings.Contains(txt, "شوینده"):
		return "supermarket", "سوپرمارکت"
	case strings.Contains(txt, "کتاب") || strings.Contains(txt, "تحریر") || strings.Contains(txt, "دفتر") || strings.Contains(txt, "خودکار") || strings.Contains(txt, "هنر"):
		return "book_art", "کتاب و هنر"
	case strings.Contains(txt, "ورزش") || strings.Contains(txt, "سفر") || strings.Contains(txt, "کمپینگ") || strings.Contains(txt, "چادر") || strings.Contains(txt, "دمبل"):
		return "sport", "ورزش و سفر"
	case strings.Contains(txt, "اسباب بازی") || strings.Contains(txt, "اسباب‌بازی") || strings.Contains(txt, "کودک") || strings.Contains(txt, "نوزاد") || strings.Contains(txt, "عروسک") || strings.Contains(txt, "لگو"):
		return "toy", "اسباب‌بازی و کودک"
	}
	return "all", "کل فروشگاه (تخفیف‌های ویژه)"
}

func (s *ServerState) fetchDkEndpoint(baseURL string, target CrawlTarget, maxPages int) []DealItem {
	var deals []DealItem
	now := time.Now()

	for page := 1; page <= maxPages; page++ {
		reqURL := fmt.Sprintf("%s&page=%d", baseURL, page)
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			break
		}

		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "fa-IR,fa;q=0.9,en;q=0.8")
		req.Header.Set("Origin", "https://www.digikala.com")
		req.Header.Set("Referer", "https://www.digikala.com/")
		req.Header.Set("x-web-client", "desktop")
		req.Header.Set("x-web-client-id", "web")

		resp, err := s.dkClient.Do(req)
		if err != nil {
			break
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			break
		}

		var dkResp DkResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&dkResp)
		resp.Body.Close()
		if decodeErr != nil {
			break
		}

		pageDeals := 0
		for _, w := range dkResp.Data.Widgets {
			if w.Type == "vertical_product_listing" {
				for _, item := range w.Data.Widgets {
					if item.Type == "product" {
						p := item.Data
						if p.ID == 0 || p.TitleFa == "" {
							continue
						}
						pr := p.DefaultVariant.Price
						sellingToman := pr.SellingPrice / 10
						rrpToman := pr.RrpPrice / 10
						disc := pr.DiscountPercent
						if disc <= 0 && rrpToman > sellingToman && rrpToman > 0 {
							disc = int((float64(rrpToman-sellingToman) / float64(rrpToman)) * 100)
						}
						if disc <= 0 {
							continue
						}

						img := ""
						if len(p.Images.Main.URL) > 0 {
							img = p.Images.Main.URL[0]
						}
						if img == "" {
							img = "https://www.digikala.com/statics/img/png/placeholder.png"
						}

						catKey := target.Key
						catName := target.Name
						if catKey == "all" {
							catKey, catName = MapItemToUnifiedCategory(
								p.TitleFa,
								p.DataLayer.ItemCategory2,
								p.DataLayer.ItemCategory3,
								p.DataLayer.Category,
							)
						}

						fullURL := "https://www.digikala.com" + p.URL.URI
						fastBuyURL := fullURL + "?quick_buy=1"

						deals = append(deals, DealItem{
							ID:              fmt.Sprintf("dk_%d", p.ID),
							Store:           "digikala",
							StoreName:       "دیجی‌کالا",
							Title:           p.TitleFa,
							CategoryKey:     catKey,
							CategoryName:    catName,
							DiscountPercent: disc,
							PriceToman:      sellingToman,
							OriginalToman:   rrpToman,
							PriceFormatted:  formatNumber(sellingToman),
							OrigFormatted:   formatNumber(rrpToman),
							ProductURL:      fullURL,
							FastBuyURL:      fastBuyURL,
							ImageURL:        img,
							FoundAt:         now,
							FoundTimeStr:    now.Format("15:04:05"),
						})
						pageDeals++
					}
				}
			}
		}

		if pageDeals < 3 {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}
	return deals
}

// --- Fetch from Digikala (Category ID + Keyword + Global Unfiltered) ---
func (s *ServerState) FetchDigikalaTarget(target CrawlTarget) ([]DealItem, error) {
	var allDeals []DealItem

	// 1. Global unfiltered highest discounts across all of Digikala
	if target.Key == "all" && target.DkQuery == "ALL_DISCOUNTS" {
		u1 := "https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26"
		allDeals = append(allDeals, s.fetchDkEndpoint(u1, target, 25)...)
		u2 := "https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&has_ready_to_shipment=1&sort=26"
		allDeals = append(allDeals, s.fetchDkEndpoint(u2, target, 15)...)
		return allDeals, nil
	}

	// 2. Numeric Category ID search (categories[]=ID)
	if target.DkCatID > 0 {
		uCat := fmt.Sprintf("https://api.digikala.com/discovery/api/v2/search?categories[]=%d&sort=26&has_selling_stock=1", target.DkCatID)
		allDeals = append(allDeals, s.fetchDkEndpoint(uCat, target, 5)...)
	}

	// 3. Keyword search
	if target.DkQuery != "" {
		uQ := fmt.Sprintf("https://api.digikala.com/discovery/api/v2/search?q=%s&sort=26&has_selling_stock=1", url.QueryEscape(target.DkQuery))
		allDeals = append(allDeals, s.fetchDkEndpoint(uQ, target, 5)...)
	}

	return allDeals, nil
}

const dealsCacheFile = "d:\\My_Script\\web-scarper\\deals_cache.json"

func (s *ServerState) LoadCache() {
	data, err := os.ReadFile(dealsCacheFile)
	if err != nil {
		return
	}
	var cached map[string]DealItem
	if err := json.Unmarshal(data, &cached); err == nil && len(cached) > 0 {
		s.mu.Lock()
		for k, v := range cached {
			s.deals[k] = v
		}
		s.mu.Unlock()
		fmt.Printf("💾 %d کالای ذخیره‌شده از حافظه دیسک بازیابی شد.\n", len(cached))
	}
}

func (s *ServerState) SaveCache() {
	s.mu.RLock()
	data, err := json.Marshal(s.deals)
	s.mu.RUnlock()
	if err == nil {
		_ = os.WriteFile(dealsCacheFile, data, 0644)
	}
}

// mergeFetchedBatch merges newly fetched deals into state immediately so categories never disappear mid-scan.
func (s *ServerState) mergeFetchedBatch(items []DealItem) {
	if len(items) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		existing, ok := s.deals[item.ProductURL]
		if !ok {
			s.deals[item.ProductURL] = item
		} else {
			existing.DiscountPercent = item.DiscountPercent
			existing.PriceToman = item.PriceToman
			existing.OriginalToman = item.OriginalToman
			existing.PriceFormatted = item.PriceFormatted
			existing.OrigFormatted = item.OrigFormatted
			existing.FastBuyURL = item.FastBuyURL
			// Only upgrade "all" to a specific category; never overwrite an existing specific category!
			if existing.CategoryKey == "all" && item.CategoryKey != "all" {
				existing.CategoryKey = item.CategoryKey
				existing.CategoryName = item.CategoryName
			}
			s.deals[item.ProductURL] = existing
		}
	}
}

// --- Run Scan for Both Stores ---
func (s *ServerState) RunScan() {
	s.mu.Lock()
	if s.isScanning {
		s.mu.Unlock()
		return
	}
	s.isScanning = true
	s.mu.Unlock()

	defer func() {
		s.SaveCache()
		s.mu.Lock()
		s.isScanning = false
		s.lastScanTime = time.Now()
		s.scanCount++
		s.mu.Unlock()
	}()

	var wg sync.WaitGroup

	for _, target := range TargetConfigs {
		// Snapp goroutine (only if target has SnappCat, SnappQuery, or is the "all" target)
		if target.SnappCat != "" || target.SnappQuery != "" || target.Key == "all" {
			wg.Add(1)
			go func(t CrawlTarget) {
				defer wg.Done()
				snappSem <- struct{}{}
				defer func() { <-snappSem }()
				items, err := s.FetchSnappTarget(t)
				if err == nil && len(items) > 0 {
					s.mergeFetchedBatch(items)
				}
			}(target)
		}

		// Digikala goroutine (only if target has DkCatID or DkQuery)
		if target.DkCatID > 0 || target.DkQuery != "" {
			wg.Add(1)
			go func(t CrawlTarget) {
				defer wg.Done()
				dkSem <- struct{}{}
				defer func() { <-dkSem }()
				items, err := s.FetchDigikalaTarget(t)
				if err == nil && len(items) > 0 {
					s.mergeFetchedBatch(items)
				}
			}(target)
		}
	}

	wg.Wait()
}

// --- Galaxy S26 Dedicated Sniper Engine ---

var persianDigitReplacer = strings.NewReplacer(
	"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
	"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
	"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
	"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
)

func NormalizePersianDigits(s string) string {
	return persianDigitReplacer.Replace(s)
}

// IsGalaxyS26Phone determines whether a product is a genuine Samsung Galaxy S26 phone between 20M and 30M Tomans (e.g. 26M).
func IsGalaxyS26Phone(title string, priceToman int64) bool {
	// 1. Price check: Must be > 20,000,000 Toman and <= 30,000,000 Toman (specifically checks 26,000,000)
	actualToman := priceToman
	if actualToman > 100000000 {
		actualToman = actualToman / 10
	}
	if actualToman < 20000000 || actualToman > 30000000 {
		return false
	}

	norm := NormalizePersianDigits(title)
	normLower := strings.ToLower(norm)

	// 2. Reject accessories
	if strings.Contains(norm, "مناسب") {
		return false
	}
	words := strings.Fields(norm)
	if len(words) > 0 {
		firstWord := words[0]
		for _, acc := range []string{"کاور", "قاب", "گلس", "محافظ", "کیف", "بند", "شارژر", "کابل", "برچسب", "هولدر", "پایه", "لنز"} {
			if strings.HasPrefix(firstWord, acc) {
				return false
			}
		}
	}

	// 3. Must be a phone: contains "گوشی" or "موبایل" or "phone" or "smartphone"
	isPhone := false
	for _, pw := range []string{"گوشی", "موبایل", "phone", "smartphone"} {
		if strings.Contains(normLower, pw) {
			isPhone = true
			break
		}
	}
	if !isPhone {
		return false
	}

	// 4. Must relate to Samsung or Galaxy
	hasBrand := false
	for _, b := range []string{"samsung", "galaxy", "سامسونگ", "گلکسی"} {
		if strings.Contains(normLower, b) {
			hasBrand = true
			break
		}
	}
	if !hasBrand {
		return false
	}

	// 5. Match S26 model in English or Persian (s26, s-26, s 26, اس 26) - digits normalized so Persian S۲۶ is also matched!
	matched, _ := regexp.MatchString(`(?i)(?:\bs[- ]?26\b|اس\s*26|s\s*26)`, normLower)
	return matched
}

func (s *ServerState) StartS26Sniper() {
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("🎯 [S26 Sniper] Dedicated Galaxy S26 background monitor started (interval: 15s, target price: 20M - 30M Toman, e.g. 26M)...")
		for {
			s.ScanForGalaxyS26()
			time.Sleep(15 * time.Second)
		}
	}()
}

func (s *ServerState) ScanForGalaxyS26() {
	queries := []string{
		"galaxy s26",
		"galaxy s۲۶",
		"گلکسی s26",
		"گلکسی s۲۶",
		"سامسونگ s26",
		"سامسونگ s۲۶",
		"s26",
		"s۲۶",
		"گوشی s26",
		"گوشی s۲۶",
	}

	s.mu.Lock()
	s.s26ScanCount++
	s.s26LastScanTime = time.Now()
	s.mu.Unlock()

	for _, q := range queries {
		s.searchSnappForS26(q, false)
		time.Sleep(300 * time.Millisecond)
	}
}

func (s *ServerState) searchSnappForS26(query string, hasDiscount bool) {
	snappThrottle()

	u, err := url.Parse(SnappAPIBaseURL)
	if err != nil {
		return
	}
	urlQ := u.Query()
	urlQ.Set("lat", s.lat)
	urlQ.Set("lng", s.lng)
	u.RawQuery = urlQ.Encode()

	reqBody := SnappSearchBody{
		Query:       query,
		Sort:        "price_asc",
		IsAvailable: true,
		HasDiscount: hasDiscount,
		Page:        1,
		Skip:        0,
		Limit:       50,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	httpReq, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}

	httpReq.Header.Set("User-Agent", UserAgent)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Origin", "https://snappshop.ir")
	httpReq.Header.Set("Referer", "https://snappshop.ir/")

	resp, err := s.snappClient.Do(httpReq)
	atomic.AddInt64(&s.snappReqs, 1)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var apiResp SnappAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return
	}

	if len(apiResp.Data.Structure) == 0 {
		return
	}

	items := apiResp.Data.Structure[0].Items
	now := time.Now()

	for _, it := range items {
		if it.Href == "" || it.Title == "" {
			continue
		}

		rawPrice := it.PriceInfo.DiscountedPrice
		if rawPrice <= 0 {
			rawPrice = it.PriceInfo.Price
		}
		if rawPrice <= 0 {
			continue
		}

		var priceToman int64
		if rawPrice > 100000000 {
			priceToman = rawPrice / 10
		} else if rawPrice > 3000000 && rawPrice <= 30000000 {
			priceToman = rawPrice
		} else if rawPrice > 30000000 && rawPrice <= 300000000 {
			priceToman = rawPrice / 10
		} else {
			priceToman = rawPrice
		}

		disc := it.PriceInfo.Discount
		if disc == 0 && it.PriceInfo.DiscountPercent > 0 {
			disc = it.PriceInfo.DiscountPercent
		}

		fullURL := "https://snappshop.ir" + it.Href
		fastBuyURL := fullURL + "?quick_buy=1&sniper=s26"
		img := it.Image.Src
		if img == "" {
			img = "https://snappshop.ir/assets/images/placeholder.png"
		}

		deal := DealItem{
			ID:              "snp_" + it.ID,
			Store:           "snapp",
			StoreName:       "اسنپ‌شاپ",
			Title:           it.Title,
			CategoryKey:     "mobile",
			CategoryName:    "گوشی موبایل",
			DiscountPercent: disc,
			PriceToman:      priceToman,
			OriginalToman:   it.PriceInfo.Price / 10,
			PriceFormatted:  formatNumber(priceToman),
			OrigFormatted:   formatNumber(it.PriceInfo.Price / 10),
			ProductURL:      fullURL,
			FastBuyURL:      fastBuyURL,
			ImageURL:        img,
			FoundAt:         now,
			FoundTimeStr:    now.Format("15:04:05"),
		}

		if IsGalaxyS26Phone(it.Title, priceToman) {
			s.TriggerS26Sniper(deal)
		}
	}
}

func (s *ServerState) TriggerS26Sniper(deal DealItem) {
	s.mu.Lock()
	if s.s26TriggeredURLs[deal.ProductURL] {
		s.mu.Unlock()
		return
	}
	s.s26TriggeredURLs[deal.ProductURL] = true
	s.s26FoundDeals = append([]DealItem{deal}, s.s26FoundDeals...)
	s.deals[deal.ProductURL] = deal
	s.mu.Unlock()

	fmt.Println("\n=======================================================")
	fmt.Println("🚨🚨🚨 [S26 SNIPER ALERT] GALAXY S26 DETECTED UNDER 30M TOMAN! 🚨🚨🚨")
	fmt.Printf("📦 عنوان: %s\n", deal.Title)
	fmt.Printf("💰 قیمت: %s تومان\n", deal.PriceFormatted)
	fmt.Printf("🔗 لینک خرید سریع: %s\n", deal.FastBuyURL)
	fmt.Println("⚡ در حال باز کردن فوری مرورگر و افزودن به سبد خرید...")
	fmt.Println("=======================================================\n")

	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "start", deal.FastBuyURL).Start()
	}

	go func() {
		for i := 0; i < 5; i++ {
			if runtime.GOOS == "windows" {
				exec.Command("powershell", "-Command", "[console]::beep(1200, 300); [console]::beep(1800, 400)").Run()
			}
			time.Sleep(300 * time.Millisecond)
		}
	}()
}

func (s *ServerState) StartBackgroundCrawler() {
	s.LoadCache()
	s.StartS26Sniper()
	ticker := time.NewTicker(time.Duration(s.intervalSec) * time.Second)
	go func() {
		s.RunScan()
		for range ticker.C {
			s.RunScan()
		}
	}()
}

func (s *ServerState) GetDeals(minDisc int, catKey, storeKey string) []DealItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]DealItem, 0)
	for _, d := range s.deals {
		if d.DiscountPercent < minDisc {
			continue
		}
		if storeKey != "" && storeKey != "all" && d.Store != storeKey {
			continue
		}
		if catKey != "" && catKey != "all" {
			if catKey == "hard" {
				if d.CategoryKey != "hard" && d.CategoryKey != "ssd" {
					continue
				}
			} else if d.CategoryKey != catKey {
				continue
			}
		}
		result = append(result, d)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].DiscountPercent == result[j].DiscountPercent {
			return result[i].PriceToman < result[j].PriceToman
		}
		return result[i].DiscountPercent > result[j].DiscountPercent
	})

	return result
}

func formatNumber(val int64) string {
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
	var b bytes.Buffer
	for i, seg := range res {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(seg)
	}
	return b.String()
}

// --- HTML Dashboard Template ---
const indexHTML = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>شکارچی تخفیف‌های ویژه | اسنپ‌شاپ و دیجی‌کالا</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <style>
        :root {
            --bg-primary: #0f111a;
            --bg-secondary: #171a26;
            --bg-card: #1e2235;
            --text-primary: #ffffff;
            --text-secondary: #94a3b8;
            --accent-snapp: #a855f7;
            --accent-dk: #ef394e;
            --accent-green: #00e676;
            --border: #2b3048;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
            font-family: system-ui, -apple-system, "Vazirmatn", "Segoe UI", Roboto, sans-serif;
        }

        body {
            background-color: var(--bg-primary);
            color: var(--text-primary);
            min-height: 100vh;
            padding-bottom: 50px;
        }

        .header {
            background: linear-gradient(180deg, #181c2e 0%, var(--bg-primary) 100%);
            border-bottom: 1px solid var(--border);
            padding: 20px 24px;
            position: sticky;
            top: 0;
            z-index: 100;
            backdrop-filter: blur(10px);
        }

        .header-content {
            max-width: 1500px;
            margin: 0 auto;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 16px;
        }

        .logo-box {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .logo-icon {
            font-size: 32px;
            background: linear-gradient(135deg, #a855f7, #ef394e);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .logo-title {
            font-size: 20px;
            font-weight: 800;
            background: linear-gradient(90deg, #ffffff, #cbd5e1);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .logo-subtitle {
            font-size: 12px;
            color: var(--text-secondary);
            margin-top: 2px;
        }

        .stats-bar {
            display: flex;
            align-items: center;
            gap: 12px;
            flex-wrap: wrap;
        }

        .stat-badge {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            padding: 7px 14px;
            border-radius: 10px;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .stat-badge.snapp {
            border-color: rgba(168, 85, 247, 0.4);
            color: #d8b4fe;
        }

        .stat-badge.dk {
            border-color: rgba(239, 57, 78, 0.4);
            color: #fca5a5;
        }

        .stat-badge span.val {
            font-weight: 700;
            color: #fff;
        }

        .banner-fastbuy {
            max-width: 1500px;
            margin: 14px auto 0 auto;
            background: linear-gradient(90deg, rgba(168, 85, 247, 0.15), rgba(239, 57, 78, 0.15));
            border: 1px solid rgba(168, 85, 247, 0.3);
            border-radius: 12px;
            padding: 12px 18px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 10px;
        }

        .banner-fastbuy-text {
            font-size: 13px;
            color: #e2e8f0;
        }

        .banner-fastbuy-text strong {
            color: #f472b6;
        }

        .banner-fastbuy-text p {
            font-size: 12px;
            color: #cbd5e1;
            margin-top: 3px;
        }

        .btn-script {
            background: linear-gradient(135deg, #a855f7, #ef394e);
            color: #fff;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 8px;
            font-size: 12px;
            font-weight: 700;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            transition: 0.2s;
            cursor: pointer;
            border: none;
            box-shadow: 0 3px 12px rgba(239, 57, 78, 0.3);
        }

        .btn-script:hover {
            opacity: 0.9;
            transform: translateY(-1px);
        }

        .controls-bar {
            max-width: 1500px;
            margin: 14px auto 18px auto;
            display: flex;
            flex-direction: column;
            gap: 12px;
        }

        /* Store Filter Tabs */
        .store-selector {
            display: flex;
            gap: 10px;
            background: var(--bg-secondary);
            padding: 6px;
            border-radius: 12px;
            border: 1px solid var(--border);
            width: fit-content;
        }

        .store-btn {
            background: transparent;
            color: var(--text-secondary);
            border: none;
            padding: 8px 18px;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 700;
            cursor: pointer;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .store-btn:hover {
            color: #fff;
            background: rgba(255, 255, 255, 0.05);
        }

        .store-btn.active[data-store="all"] {
            background: linear-gradient(135deg, #3b82f6, #6366f1);
            color: #fff;
            box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
        }

        .store-btn.active[data-store="snapp"] {
            background: linear-gradient(135deg, #a855f7, #7c3aed);
            color: #fff;
            box-shadow: 0 4px 12px rgba(124, 58, 237, 0.4);
        }

        .store-btn.active[data-store="digikala"] {
            background: linear-gradient(135deg, #ef394e, #c0142b);
            color: #fff;
            box-shadow: 0 4px 12px rgba(239, 57, 78, 0.4);
        }

        .category-scroll-container {
            display: flex;
            gap: 8px;
            overflow-x: auto;
            padding-bottom: 6px;
            scrollbar-width: thin;
            scrollbar-color: #3b4263 transparent;
        }

        .category-scroll-container::-webkit-scrollbar {
            height: 6px;
        }

        .category-scroll-container::-webkit-scrollbar-thumb {
            background: #3b4263;
            border-radius: 4px;
        }

        .tab-btn {
            background: var(--bg-secondary);
            color: var(--text-secondary);
            border: 1px solid var(--border);
            padding: 8px 14px;
            border-radius: 10px;
            font-size: 13px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s ease;
            display: flex;
            align-items: center;
            gap: 6px;
            white-space: nowrap;
            flex-shrink: 0;
        }

        .tab-btn:hover {
            background: var(--bg-card);
            color: #fff;
            border-color: #4b5275;
        }

        .tab-btn.active {
            background: linear-gradient(135deg, #a855f7, #ef394e);
            color: #fff;
            border-color: transparent;
            box-shadow: 0 4px 15px rgba(239, 57, 78, 0.3);
        }

        .tab-btn.highlight-hard {
            border: 1px solid rgba(41, 121, 255, 0.6);
            background: rgba(41, 121, 255, 0.15);
            color: #82b1ff;
        }

        .tab-btn.highlight-hard.active {
            background: linear-gradient(135deg, #2979ff, #00b0ff);
            color: #fff;
            box-shadow: 0 4px 15px rgba(41, 121, 255, 0.4);
        }

        .filters-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 14px;
            background: var(--bg-secondary);
            padding: 12px 18px;
            border-radius: 12px;
            border: 1px solid var(--border);
        }

        .filter-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .filter-group label {
            font-size: 13px;
            font-weight: 600;
            color: var(--text-secondary);
        }

        .disc-select {
            background: var(--bg-card);
            color: #fff;
            border: 1px solid var(--border);
            padding: 7px 12px;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 600;
            outline: none;
            cursor: pointer;
        }

        .search-input {
            background: var(--bg-card);
            color: #fff;
            border: 1px solid var(--border);
            padding: 7px 14px;
            border-radius: 8px;
            font-size: 13px;
            width: 230px;
            outline: none;
        }

        .btn-action {
            background: var(--bg-card);
            color: #fff;
            border: 1px solid var(--border);
            padding: 7px 14px;
            border-radius: 8px;
            font-size: 12px;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 6px;
            transition: 0.2s;
        }

        .btn-action:hover {
            background: #2f3554;
        }

        .grid {
            max-width: 1500px;
            margin: 0 auto;
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
            gap: 18px;
        }

        .card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 14px;
            overflow: hidden;
            display: flex;
            flex-direction: column;
            transition: transform 0.2s, box-shadow 0.2s;
            position: relative;
        }

        .card:hover {
            transform: translateY(-4px);
            box-shadow: 0 10px 25px rgba(0,0,0,0.5);
            border-color: #444c70;
        }

        .card-img-box {
            width: 100%;
            height: 190px;
            background: #fff;
            padding: 14px;
            display: flex;
            justify-content: center;
            align-items: center;
            position: relative;
        }

        .card-img {
            max-width: 100%;
            max-height: 100%;
            object-fit: contain;
        }

        .badge-discount {
            position: absolute;
            top: 10px;
            right: 10px;
            background: linear-gradient(135deg, #ff1744, #d50000);
            color: #fff;
            font-size: 13px;
            font-weight: 800;
            padding: 3px 9px;
            border-radius: 16px;
            box-shadow: 0 3px 10px rgba(213, 0, 0, 0.4);
            z-index: 2;
        }

        .badge-super {
            background: linear-gradient(135deg, #d500f9, #651fff);
            box-shadow: 0 4px 12px rgba(101, 31, 255, 0.4);
            animation: pulse-super 1.5s infinite;
        }

        @keyframes pulse-super {
            0% { transform: scale(1); }
            50% { transform: scale(1.06); }
            100% { transform: scale(1); }
        }

        .badge-store {
            position: absolute;
            top: 10px;
            left: 10px;
            font-size: 11px;
            font-weight: 700;
            padding: 3px 8px;
            border-radius: 6px;
            z-index: 2;
        }

        .badge-store.snapp {
            background: linear-gradient(135deg, #a855f7, #7c3aed);
            color: #fff;
            box-shadow: 0 2px 8px rgba(124, 58, 237, 0.4);
        }

        .badge-store.digikala {
            background: linear-gradient(135deg, #ef394e, #c0142b);
            color: #fff;
            box-shadow: 0 2px 8px rgba(239, 57, 78, 0.4);
        }

        .card-content {
            padding: 14px;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            flex-grow: 1;
        }

        .category-tag {
            font-size: 11px;
            color: #93c5fd;
            background: rgba(59, 130, 246, 0.15);
            padding: 2px 8px;
            border-radius: 6px;
            display: inline-block;
            margin-bottom: 6px;
        }

        .card-title {
            font-size: 13px;
            font-weight: 600;
            line-height: 1.5;
            color: #e2e8f0;
            height: 40px;
            overflow: hidden;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
            margin-bottom: 12px;
        }

        .price-box {
            display: flex;
            align-items: baseline;
            justify-content: space-between;
            margin-bottom: 12px;
        }

        .curr-price {
            font-size: 16px;
            font-weight: 800;
            color: var(--accent-green);
        }

        .curr-price span {
            font-size: 12px;
            font-weight: 400;
            color: var(--text-secondary);
        }

        .orig-price {
            font-size: 12px;
            color: #64748b;
            text-decoration: line-through;
        }

        .btn-group {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .btn-fastbuy {
            background: linear-gradient(135deg, #a855f7, #d946ef);
            color: #fff;
            text-decoration: none;
            text-align: center;
            padding: 9px;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 700;
            transition: all 0.2s;
            display: block;
            box-shadow: 0 4px 14px rgba(168, 85, 247, 0.35);
        }

        .btn-fastbuy.dk-fastbuy {
            background: linear-gradient(135deg, #ef394e, #f43f5e);
            box-shadow: 0 4px 14px rgba(239, 57, 78, 0.35);
        }

        .btn-fastbuy:hover {
            opacity: 0.92;
            transform: translateY(-2px);
        }

        .btn-view {
            background: rgba(255, 255, 255, 0.05);
            color: #94a3b8;
            text-decoration: none;
            text-align: center;
            padding: 7px;
            border-radius: 8px;
            font-size: 12px;
            transition: 0.2s;
            border: 1px solid rgba(255, 255, 255, 0.08);
        }

        .btn-view:hover {
            background: rgba(255, 255, 255, 0.1);
            color: #fff;
        }

        .cat-count {
            background: rgba(255, 255, 255, 0.14);
            color: #e2e8f0;
            font-size: 11px;
            font-weight: 700;
            padding: 1px 6px;
            border-radius: 8px;
            margin-right: 4px;
        }
        .tab-btn.active .cat-count {
            background: rgba(0, 0, 0, 0.28);
            color: #fff;
        }

        .empty-state {
            grid-column: 1 / -1;
            text-align: center;
            padding: 70px 20px;
            background: var(--bg-secondary);
            border-radius: 16px;
            border: 1px dashed var(--border);
        }

        .loading-bar {
            height: 3px;
            width: 100%;
            background: linear-gradient(90deg, #38bdf8, #a855f7, #ef4444);
            background-size: 200% 100%;
            animation: moveGradient 1.5s infinite linear;
            display: none;
            position: sticky;
            top: 0;
            z-index: 100;
        }
        @keyframes moveGradient {
            0% { background-position: 100% 0; }
            100% { background-position: -100% 0; }
        }
        .grid.loading-dim {
            opacity: 0.5;
            transition: opacity 0.2s ease;
            pointer-events: none;
        }
        .btn-clear-filter {
            background: linear-gradient(135deg, #ef4444, #f43f5e);
            color: #fff;
            border: none;
            padding: 10px 22px;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 700;
            cursor: pointer;
            margin-top: 15px;
            transition: all 0.2s;
            box-shadow: 0 4px 14px rgba(239, 68, 68, 0.35);
        }
        .btn-clear-filter:hover {
            opacity: 0.92;
            transform: translateY(-2px);
        }
        .btn-suggestion {
            background: rgba(255, 255, 255, 0.08);
            color: #38bdf8;
            border: 1px solid rgba(56, 189, 248, 0.3);
            padding: 8px 16px;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 600;
            cursor: pointer;
            margin: 6px 4px;
            transition: all 0.2s;
        }
        .btn-suggestion:hover {
            background: rgba(56, 189, 248, 0.2);
            color: #fff;
        }
        .scroll-status {
            grid-column: 1 / -1;
            text-align: center;
            padding: 25px 15px;
            color: #94a3b8;
            font-size: 13px;
        }

        .banner-sniper {
            background: linear-gradient(135deg, rgba(220, 38, 38, 0.15), rgba(185, 28, 28, 0.25));
            border: 1px solid rgba(239, 68, 68, 0.4);
            border-radius: 12px;
            padding: 12px 18px;
            margin-top: 14px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 12px;
        }
        .sniper-left {
            display: flex;
            align-items: center;
            gap: 12px;
            flex-wrap: wrap;
        }
        .sniper-badge-pulse {
            background: #ef4444;
            color: #fff;
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 13px;
            font-weight: 700;
            box-shadow: 0 0 12px rgba(239, 68, 68, 0.6);
            animation: pulseRed 2s infinite;
        }
        @keyframes pulseRed {
            0% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7); }
            70% { box-shadow: 0 0 0 10px rgba(239, 68, 68, 0); }
            100% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0); }
        }
        .sniper-info {
            font-size: 13px;
            color: #fca5a5;
            display: flex;
            flex-direction: column;
            gap: 2px;
        }
        .sniper-actions {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .sniper-status-pill {
            background: rgba(0, 0, 0, 0.35);
            border: 1px solid rgba(255, 255, 255, 0.12);
            padding: 6px 12px;
            border-radius: 8px;
            font-size: 12px;
            color: #cbd5e1;
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .live-dot {
            width: 8px;
            height: 8px;
            background: #22c55e;
            border-radius: 50%;
            display: inline-block;
            box-shadow: 0 0 8px #22c55e;
        }
        .btn-sniper-test {
            background: linear-gradient(135deg, #ef4444, #dc2626);
            color: #fff;
            border: none;
            padding: 8px 16px;
            border-radius: 8px;
            font-size: 12px;
            font-weight: 700;
            cursor: pointer;
            transition: all 0.2s;
            box-shadow: 0 2px 10px rgba(239, 68, 68, 0.4);
        }
        .btn-sniper-test:hover {
            opacity: 0.9;
            transform: translateY(-1px);
        }
    </style>
</head>
<body>
    <div id="loadingBar" class="loading-bar"></div>

    <header class="header">
        <div class="header-content">
            <div class="logo-box">
                <div class="logo-icon">⚡</div>
                <div>
                    <h1 class="logo-title">شکارچی تخفیف‌های ویژه (اسنپ‌شاپ + دیجی‌کالا)</h1>
                    <div class="logo-subtitle">پایش مداوم کالاها با بیشترین درصد تخفیف هر ۱ دقیقه (تا ۵ صفحه در هر بخش)</div>
                </div>
            </div>

            <div class="stats-bar">
                <div class="stat-badge snapp">
                    <span>🟣 اسنپ‌شاپ:</span>
                    <span class="val" id="snappDealCount">...</span>
                </div>
                <div class="stat-badge dk">
                    <span>🔴 دیجی‌کالا:</span>
                    <span class="val" id="dkDealCount">...</span>
                </div>
                <div class="stat-badge">
                    <span>📦 مجموع کالاها:</span>
                    <span class="val" id="totalDealCount">...</span>
                </div>
                <div class="stat-badge">
                    <span>⏱️ آخرین اسکن:</span>
                    <span class="val" id="lastScan">درحال دریافت...</span>
                </div>
                <button class="btn-action" onclick="triggerManualScan()">⚡ اسکن فوری</button>
            </div>
        </div>

        <!-- Userscript Banner -->
        <div class="banner-fastbuy">
            <div class="banner-fastbuy-text">
                <strong>⚡ قابلیت خرید خودکار تا درگاه بانکی:</strong>
                اسکریپت تمپرمونکی نسخه ۳.۰ فعال شد! به صورت کاملاً خودکار کالا به سبد اضافه شده، ثبت سفارش زده می‌شود، اولین روز و ساعت فعال ارسال انتخاب و تایید می‌گردد و مستقیماً به درگاه پرداخت بانک منتقل می‌شوید.
            </div>
            <a href="/snapp_fast_buy.user.js" target="_blank" class="btn-script">
                📥 نصب / آپدیت اسکریپت خرید ۱-کلیکه (نسخه ۳.۰ مجهز به توالی کامل تا درگاه بانک)
            </a>
        </div>

        <!-- Galaxy S26 Dedicated Sniper Box -->
        <div class="banner-sniper">
            <div class="sniper-left">
                <div class="sniper-badge-pulse">🎯 اسنایپر هوشمند Galaxy S26 / S۲۶</div>
                <div class="sniper-info">
                    <strong>پایش اختصاصی اسنپ‌شاپ برای گوشی سامسونگ Galaxy S26 (شامل عدد فارسی S۲۶):</strong>
                    <span>فرکانس: هر ۱۵ ثانیه | شرط خودکار: <strong>قیمت بالای ۲۰ و زیر ۳۰ میلیون تومان (به‌ویژه ۲۶ میلیون)</strong> | عملکرد: باز کردن تب جدید + افزودن به سبد + انتخاب روز و ساعت ارسال + انتقال مستقیم به درگاه بانکی + آلارم صوتی</span>
                </div>
            </div>
            <div class="sniper-actions">
                <div class="sniper-status-pill">
                    <span class="live-dot"></span>
                    <span>اسکن S26: <strong id="s26ScanCount" style="color: #38bdf8;">0</strong> بار</span>
                </div>
                <button onclick="triggerS26Test()" class="btn-sniper-test">⚡ تست شبیه‌ساز خرید S26</button>
            </div>
        </div>
    </header>

    <div class="controls-bar">
        <!-- Store Selector -->
        <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 10px;">
            <div class="store-selector">
                <button class="store-btn active" data-store="all" onclick="selectStore('all')">🌟 همه فروشگاه‌ها</button>
                <button class="store-btn" data-store="snapp" onclick="selectStore('snapp')">🟣 فقط اسنپ‌شاپ</button>
                <button class="store-btn" data-store="digikala" onclick="selectStore('digikala')">🔴 فقط دیجی‌کالا</button>
            </div>
            <div style="font-size: 13px; color: var(--text-secondary);">
                بروزرسانی بعدی در <strong id="countdown" style="color: #38bdf8;">60</strong> ثانیه
            </div>
        </div>

        <!-- Categories Scroll -->
        <div class="category-scroll-container">
            <button class="tab-btn highlight-hard active" data-cat="hard" onclick="selectCategory('hard')">
                🔥 هارد اکسترنال (تخصصی)
            </button>
            <button class="tab-btn" data-cat="ssd" onclick="selectCategory('ssd')">💾 حافظه SSD</button>
            <button class="tab-btn" data-cat="mobile" onclick="selectCategory('mobile')">📱 گوشی موبایل</button>
            <button class="tab-btn" data-cat="laptop" onclick="selectCategory('laptop')">💻 لپ‌تاپ و کامپیوتر</button>
            <button class="tab-btn" data-cat="tablet" onclick="selectCategory('tablet')">📟 تبلت و آیپد</button>
            <button class="tab-btn" data-cat="headphone" onclick="selectCategory('headphone')">🎧 هدفون و هندزفری</button>
            <button class="tab-btn" data-cat="speaker" onclick="selectCategory('speaker')">🔊 اسپیکر و صوتی</button>
            <button class="tab-btn" data-cat="camera" onclick="selectCategory('camera')">📷 دوربین عکاسی</button>
            <button class="tab-btn" data-cat="gadgets" onclick="selectCategory('gadgets')">⌚ ساعت هوشمند</button>
            <button class="tab-btn" data-cat="powerbank" onclick="selectCategory('powerbank')">🔋 پاوربانک</button>
            <button class="tab-btn" data-cat="gaming" onclick="selectCategory('gaming')">🎮 گیمینگ و کنسول</button>
            <button class="tab-btn" data-cat="home" onclick="selectCategory('home')">🏠 لوازم خانگی</button>
            <button class="tab-btn" data-cat="kitchen" onclick="selectCategory('kitchen')">🍳 خانه و آشپزخانه</button>
            <button class="tab-btn" data-cat="fridge" onclick="selectCategory('fridge')">🧊 یخچال و فریزر</button>
            <button class="tab-btn" data-cat="vacuum" onclick="selectCategory('vacuum')">🧹 جاروبرقی</button>
            <button class="tab-btn" data-cat="washing" onclick="selectCategory('washing')">🧺 لباسشویی</button>
            <button class="tab-btn" data-cat="shoes" onclick="selectCategory('shoes')">👟 کفش</button>
            <button class="tab-btn" data-cat="gold" onclick="selectCategory('gold')">🥇 طلا و نقره</button>
            <button class="tab-btn" data-cat="coffee" onclick="selectCategory('coffee')">☕ قهوه‌ساز و تمپر</button>
            <button class="tab-btn" data-cat="beauty" onclick="selectCategory('beauty')">💄 زیبایی و سلامت</button>
            <button class="tab-btn" data-cat="health" onclick="selectCategory('health')">🏃 سلامت و ورزش</button>
            <button class="tab-btn" data-cat="eyewear" onclick="selectCategory('eyewear')">👓 عینک طبی</button>
            <button class="tab-btn" data-cat="supplement" onclick="selectCategory('supplement')">💊 کپسول و مکمل</button>
            <button class="tab-btn" data-cat="fashion" onclick="selectCategory('fashion')">👕 مد و پوشاک</button>
            <button class="tab-btn" data-cat="sport" onclick="selectCategory('sport')">⚽ ورزش و سفر</button>
            <button class="tab-btn" data-cat="toy" onclick="selectCategory('toy')">🧸 اسباب‌بازی و کودک</button>
            <button class="tab-btn" data-cat="book_art" onclick="selectCategory('book_art')">📚 کتاب و هنر</button>
            <button class="tab-btn" data-cat="tools_auto" onclick="selectCategory('tools_auto')">🔧 ابزار و خودرو</button>
            <button class="tab-btn" data-cat="supermarket" onclick="selectCategory('supermarket')">🛒 سوپرمارکت</button>
            <button class="tab-btn" data-cat="all" onclick="selectCategory('all')">🌐 کل فروشگاه‌ها (شگفت‌انگیز)</button>
        </div>

        <!-- Filters Row -->
        <div class="filters-row">
            <div class="filter-group">
                <label>حداقل درصد تخفیف:</label>
                <select id="discFilter" class="disc-select" onchange="applyFilters()">
                    <option value="70" selected>🔥 بالای ۷۰٪ تخفیف</option>
                    <option value="80">🚀 بالای ۸۰٪ تخفیف</option>
                    <option value="50">⚡ بالای ۵۰٪ تخفیف</option>
                    <option value="30">✨ بالای ۳۰٪ تخفیف</option>
                    <option value="1">همه موارد تخفیف‌دار</option>
                </select>
            </div>

            <div class="filter-group">
                <label>سقف قیمت (تومان):</label>
                <input type="number" id="maxPrice" class="search-input" style="width:150px" min="0" step="100000" placeholder="مثلا 5000000" oninput="applyFilters()">
            </div>

            <div class="filter-group">
                <label>مرتب‌سازی:</label>
                <select id="sortBy" class="disc-select" onchange="applyFilters()">
                    <option value="disc_desc" selected>بیشترین تخفیف</option>
                    <option value="price_asc">ارزان‌ترین</option>
                    <option value="price_desc">گران‌ترین</option>
                    <option value="newest">جدیدترین</option>
                </select>
            </div>

            <div class="filter-group">
                <input type="text" id="searchInput" class="search-input" placeholder="🔍 جستجو در نتایج..." oninput="applyFilters()">
            </div>
        </div>
    </div>

    <!-- Products Grid -->
    <main class="grid" id="productsGrid"></main>

    <script>
        let currentCategory = 'hard';
        let currentStore = 'all';
        let countdownSec = 60;
        let activeFetchId = 0;
        let currentAbortController = null;

        // Infinite Scroll & Pagination State
        let rawDeals = [];
        let filteredDeals = [];
        let renderedCount = 0;
        const BATCH_SIZE = 40;
        let isLoadingMore = false;

        function selectCategory(cat) {
            if (currentCategory === cat) return;
            currentCategory = cat;
            document.querySelectorAll('.tab-btn').forEach(btn => {
                if (btn.getAttribute('data-cat') === cat) {
                    btn.classList.add('active');
                } else {
                    btn.classList.remove('active');
                }
            });
            loadDeals(false);
        }

        function selectStore(store) {
            if (currentStore === store) return;
            currentStore = store;
            document.querySelectorAll('.store-btn').forEach(btn => {
                if (btn.getAttribute('data-store') === store) {
                    btn.classList.add('active');
                } else {
                    btn.classList.remove('active');
                }
            });
            loadDeals(false);
        }

        function setDiscountFilter(val) {
            document.getElementById('discFilter').value = val;
            loadDeals(false);
        }

        function setStoreFilter(store) {
            selectStore(store);
        }

        function clearFilters() {
            document.getElementById('searchInput').value = '';
            document.getElementById('maxPrice').value = '';
            applyFilters();
        }

        function applyFilters() {
            const searchQ = document.getElementById('searchInput').value.trim().toLowerCase();
            const maxPrice = parseInt(document.getElementById('maxPrice').value, 10) || 0;
            const sortBy = document.getElementById('sortBy').value;

            filteredDeals = rawDeals.filter(d => {
                if (searchQ && !d.title.toLowerCase().includes(searchQ)) return false;
                if (maxPrice > 0 && d.price_toman > maxPrice) return false;
                return true;
            });

            if (sortBy === 'price_asc') filteredDeals.sort((a, b) => a.price_toman - b.price_toman);
            else if (sortBy === 'price_desc') filteredDeals.sort((a, b) => b.price_toman - a.price_toman);
            else if (sortBy === 'newest') filteredDeals.sort((a, b) => (b.found_at || '').localeCompare(a.found_at || ''));
            else filteredDeals.sort((a, b) => b.discount_percent - a.discount_percent || a.price_toman - b.price_toman);

            resetAndRenderGrid();
        }

        async function updateStatus() {
            try {
                const minDisc = document.getElementById('discFilter').value;
                const res = await fetch('/api/status?min_discount=' + minDisc + '&store=' + currentStore);
                const data = await res.json();
                document.getElementById('lastScan').innerText = data.last_scan_time || '---';
                document.getElementById('snappDealCount').innerText = (data.snapp_deals || 0).toLocaleString();
                document.getElementById('dkDealCount').innerText = (data.dk_deals || 0).toLocaleString();
                document.getElementById('totalDealCount').innerText = (data.total_deals || 0).toLocaleString();
                if (data.s26_scan_count !== undefined) {
                    const s26El = document.getElementById('s26ScanCount');
                    if (s26El) s26El.innerText = data.s26_scan_count.toLocaleString();
                }
                if (data.cat_counts) {
                    document.querySelectorAll('.tab-btn').forEach(btn => {
                        const cat = btn.getAttribute('data-cat');
                        let badge = btn.querySelector('.cat-count');
                        if (!badge) {
                            badge = document.createElement('span');
                            badge.className = 'cat-count';
                            btn.appendChild(badge);
                        }
                        const cnt = data.cat_counts[cat] || 0;
                        badge.innerText = cnt.toLocaleString();
                    });
                }
            } catch (err) {
                console.error("Status update error:", err);
            }
        }

        async function loadDeals(isBackground = false) {
            const fetchId = ++activeFetchId;
            const loadingBar = document.getElementById('loadingBar');
            const grid = document.getElementById('productsGrid');

            if (!isBackground) {
                if (loadingBar) loadingBar.style.display = 'block';
                if (grid.children.length > 0) {
                    grid.classList.add('loading-dim');
                } else {
                    grid.innerHTML = ` + "`" + `
                        <div class="empty-state">
                            <div style="font-size: 32px; margin-bottom: 8px;">⏳</div>
                            <h3 style="color: #94a3b8;">در حال دریافت و آماده‌سازی تخفیف‌ها...</h3>
                        </div>
                    ` + "`" + `;
                }
            }

            if (currentAbortController) {
                currentAbortController.abort();
            }
            currentAbortController = new AbortController();

            const minDisc = document.getElementById('discFilter').value;
            const url = '/api/deals?min_discount=' + minDisc + '&category=' + currentCategory + '&store=' + currentStore;

            try {
                const res = await fetch(url, { signal: currentAbortController.signal });
                const deals = await res.json();

                if (fetchId !== activeFetchId) return;

                rawDeals = Array.isArray(deals) ? deals : [];

                const searchQ = document.getElementById('searchInput').value.trim().toLowerCase();
                const maxPrice = parseInt(document.getElementById('maxPrice').value, 10) || 0;
                const sortBy = document.getElementById('sortBy').value;

                filteredDeals = rawDeals.filter(d => {
                    if (searchQ && !d.title.toLowerCase().includes(searchQ)) return false;
                    if (maxPrice > 0 && d.price_toman > maxPrice) return false;
                    return true;
                });

                if (sortBy === 'price_asc') filteredDeals.sort((a, b) => a.price_toman - b.price_toman);
                else if (sortBy === 'price_desc') filteredDeals.sort((a, b) => b.price_toman - a.price_toman);
                else if (sortBy === 'newest') filteredDeals.sort((a, b) => (b.found_at || '').localeCompare(a.found_at || ''));
                else filteredDeals.sort((a, b) => b.discount_percent - a.discount_percent || a.price_toman - b.price_toman);

                if (isBackground && grid.children.length > 0 && !grid.querySelector('.empty-state')) {
                    // Update badges quietly without wiping the DOM or resetting scroll
                    updateStatus();
                    return;
                }

                resetAndRenderGrid();
                updateStatus();
            } catch (err) {
                if (err.name === 'AbortError') return;
                console.error("Error fetching deals:", err);
                if (fetchId === activeFetchId) {
                    rawDeals = [];
                    filteredDeals = [];
                    resetAndRenderGrid();
                }
            } finally {
                if (fetchId === activeFetchId) {
                    if (loadingBar) loadingBar.style.display = 'none';
                    grid.classList.remove('loading-dim');
                }
            }
        }

        function resetAndRenderGrid() {
            const grid = document.getElementById('productsGrid');
            grid.innerHTML = '';
            renderedCount = 0;

            if (filteredDeals.length === 0) {
                renderEmptyState(grid);
                return;
            }

            renderNextBatch();
        }

        function renderEmptyState(container) {
            const searchQ = document.getElementById('searchInput').value.trim();
            const maxPrice = parseInt(document.getElementById('maxPrice').value, 10) || 0;
            const minDisc = document.getElementById('discFilter').value;

            if (rawDeals.length > 0) {
                let filterInfo = [];
                if (searchQ) filterInfo.push('جستجو: "' + searchQ + '"');
                if (maxPrice > 0) filterInfo.push('سقف قیمت: ' + maxPrice.toLocaleString() + ' تومان');

                container.innerHTML = ` + "`" + `
                    <div class="empty-state">
                        <div style="font-size: 38px; margin-bottom: 10px;">🔍</div>
                        <h3 style="color: #f59e0b;">تعداد ` + "`" + ` + rawDeals.length.toLocaleString() + ` + "`" + ` کالا در این بخش موجود است، اما با فیلترهای شما مطابقت ندارند!</h3>
                        <p style="color: #94a3b8; font-size: 14px; margin-top: 10px;">
                            فیلترهای فعال: <strong>` + "`" + ` + filterInfo.join(' | ') + ` + "`" + `</strong>
                        </p>
                        <button onclick="clearFilters()" class="btn-clear-filter">
                            ❌ حذف فیلترها و مشاهده همه ` + "`" + ` + rawDeals.length.toLocaleString() + ` + "`" + ` کالا
                        </button>
                    </div>
                ` + "`" + `;
            } else {
                const storeText = currentStore === 'all' ? 'همه فروشگاه‌ها' : (currentStore === 'snapp' ? 'اسنپ‌شاپ' : 'دیجی‌کالا');
                const storeBtn = currentStore !== 'all' ? '<button onclick="setStoreFilter(\'all\')" class="btn-suggestion">🌟 مشاهده در همه فروشگاه‌ها</button>' : '';
                container.innerHTML = ` + "`" + `
                    <div class="empty-state">
                        <div style="font-size: 38px; margin-bottom: 10px;">📭</div>
                        <h3>هیچ کالایی با حداقل %` + "`" + ` + minDisc + ` + "`" + ` تخفیف در این بخش (` + "`" + ` + storeText + ` + "`" + `) یافت نشد.</h3>
                        <p style="color: #64748b; font-size: 13px; margin: 12px 0;">
                            می‌توانید درصد تخفیف را کاهش دهید یا فروشگاه دیگری را انتخاب کنید:
                        </p>
                        <div>
                            <button onclick="setDiscountFilter(50)" class="btn-suggestion">⚡ مشاهده با حداقل ۵۰٪ تخفیف</button>
                            <button onclick="setDiscountFilter(30)" class="btn-suggestion">✨ مشاهده با حداقل ۳۰٪ تخفیف</button>
                            ` + "`" + ` + storeBtn + ` + "`" + `
                        </div>
                    </div>
                ` + "`" + `;
            }
        }

        function createCardHTML(d) {
            const isSuper = d.discount_percent >= 80;
            const badgeClass = isSuper ? 'badge-discount badge-super' : 'badge-discount';
            const isDk = d.store === 'digikala';
            const storeBadgeClass = isDk ? 'badge-store digikala' : 'badge-store snapp';
            const fastBuyBtnClass = isDk ? 'btn-fastbuy dk-fastbuy' : 'btn-fastbuy';
            const fastBuyLink = (d.fast_buy_url || d.product_url + '?quick_buy=1');
            const normalLink = d.product_url;
            const storeActionText = isDk ? 'مشاهده در دیجی‌کالا ↗' : 'مشاهده در اسنپ‌شاپ ↗';

            return ` + "`" + `
                <div class="card">
                    <div class="card-img-box">
                        <span class="${badgeClass}">%${d.discount_percent} تخفیف</span>
                        <span class="${storeBadgeClass}">${d.store_name}</span>
                        <img src="${d.image_url}" alt="${d.title}" class="card-img" loading="lazy" onerror="this.src='https://snappshop.ir/assets/images/placeholder.png'">
                    </div>
                    <div class="card-content">
                        <div>
                            <span class="category-tag">${d.category_name}</span>
                            <h2 class="card-title" title="${d.title}">${d.title}</h2>
                        </div>
                        <div>
                            <div class="price-box">
                                <div class="curr-price">${d.price_formatted} <span>تومان</span></div>
                                ${d.orig_formatted !== '0' && d.orig_formatted !== d.price_formatted ? ` + "`" + `<div class="orig-price">${d.orig_formatted} تومان</div>` + "`" + ` : ''}
                            </div>
                            <div class="btn-group">
                                <a href="${fastBuyLink}" target="_blank" rel="noopener noreferrer" class="${fastBuyBtnClass}">
                                    ⚡ خرید ۱-کلیکه (انتقال به شیپینگ)
                                </a>
                                <a href="${normalLink}" target="_blank" rel="noopener noreferrer" class="btn-view">
                                    ${storeActionText}
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            ` + "`" + `;
        }

        function renderNextBatch() {
            if (isLoadingMore || renderedCount >= filteredDeals.length) return;
            isLoadingMore = true;

            const grid = document.getElementById('productsGrid');
            const oldStatus = document.getElementById('scrollStatusIndicator');
            if (oldStatus) oldStatus.remove();

            const nextSlice = filteredDeals.slice(renderedCount, renderedCount + BATCH_SIZE);
            let chunkHtml = '';
            for (let i = 0; i < nextSlice.length; i++) {
                chunkHtml += createCardHTML(nextSlice[i]);
            }

            grid.insertAdjacentHTML('beforeend', chunkHtml);
            renderedCount += nextSlice.length;

            const statusDiv = document.createElement('div');
            statusDiv.id = 'scrollStatusIndicator';
            statusDiv.className = 'scroll-status';
            if (renderedCount < filteredDeals.length) {
                statusDiv.innerHTML = 'در حال نمایش <strong>' + renderedCount.toLocaleString() + '</strong> از <strong>' + filteredDeals.length.toLocaleString() + '</strong> کالا (اسکرول کنید تا بقیه نمایش داده شوند...)';
            } else {
                statusDiv.innerHTML = '✅ تمام <strong>' + filteredDeals.length.toLocaleString() + '</strong> کالا بارگذاری شدند.';
            }
            grid.appendChild(statusDiv);

            isLoadingMore = false;
        }

        window.addEventListener('scroll', () => {
            if (isLoadingMore || renderedCount >= filteredDeals.length) return;
            const scrollBottom = window.innerHeight + window.scrollY;
            const docHeight = document.documentElement.scrollHeight;
            if (scrollBottom >= docHeight - 700) {
                renderNextBatch();
            }
        }, { passive: true });

        async function triggerManualScan() {
            const btn = event.currentTarget;
            btn.innerText = "⏳ در حال اسکن...";
            btn.disabled = true;
            try {
                await fetch('/api/trigger', { method: 'POST' });
                countdownSec = 60;
                await loadDeals(false);
            } finally {
                btn.innerText = "⚡ اسکن فوری";
                btn.disabled = false;
            }
        }

        async function triggerS26Test() {
            if (confirm("آیا مایلید تست اسنایپر S26 را اجرا کنید؟\nمرورگر با محصول تستی باز شده و افزونه تمپرمونکی آن را در کسری از ثانیه به سبد اضافه کرده و به صفحه شیپینگ می‌برد.")) {
                try {
                    await fetch('/api/sniper/test', { method: 'POST' });
                } catch(e) {
                    console.error(e);
                }
            }
        }

        setInterval(() => {
            countdownSec--;
            if (countdownSec <= 0) {
                countdownSec = 60;
                loadDeals(true);
            }
            document.getElementById('countdown').innerText = countdownSec;
        }, 1000);

        loadDeals(false);
    </script>
</body>
</html>
`

func main() {
	portPtr := flag.Int("port", 8080, "Web dashboard port")
	intervalPtr := flag.Int("interval", 60, "Background crawl interval in seconds")
	latPtr := flag.String("lat", DefaultLat, "Delivery latitude")
	lngPtr := flag.String("lng", DefaultLng, "Delivery longitude")
	noBrowserPtr := flag.Bool("no-browser", false, "Do not auto open dashboard in browser")

	flag.Parse()

	state := NewServerState(*intervalPtr, *latPtr, *lngPtr)

	fmt.Printf("\n========================================================================\n")
	fmt.Printf("🚀 سرور پایش دوگانه اسنپ‌شاپ + دیجی‌کالا راه‌اندازی شد!\n")
	fmt.Printf("🟣 پایش ۲۵ دسته‌بندی اسنپ‌شاپ\n")
	fmt.Printf("🔴 پایش ۲۵ دسته‌بندی دیجی‌کالا (تا ۵ صفحه، مرتب بر اساس بیشترین تخفیف)\n")
	fmt.Printf("⏱️ بازه پایش خودکار: هر %d ثانیه یک‌بار\n", *intervalPtr)
	fmt.Printf("🌐 داشبورد وب در آدرس: http://localhost:%d\n", *portPtr)
	fmt.Printf("========================================================================\n\n")

	state.StartBackgroundCrawler()

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl, _ := template.New("index").Parse(indexHTML)
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/snapp_fast_buy.user.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(w, r, "d:\\My_Script\\web-scarper\\snapp_fast_buy.user.js")
	})

	http.HandleFunc("/api/deals", func(w http.ResponseWriter, r *http.Request) {
		minDisc := 70
		if m := r.URL.Query().Get("min_discount"); m != "" {
			if v, err := strconv.Atoi(m); err == nil {
				minDisc = v
			}
		}
		cat := r.URL.Query().Get("category")
		store := r.URL.Query().Get("store")
		deals := state.GetDeals(minDisc, cat, store)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(deals)
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		minDisc := 70
		if m := r.URL.Query().Get("min_discount"); m != "" {
			if v, err := strconv.Atoi(m); err == nil {
				minDisc = v
			}
		}
		storeFilter := r.URL.Query().Get("store")

		state.mu.RLock()
		defer state.mu.RUnlock()

		snappCount := 0
		dkCount := 0
		catCounts := make(map[string]int)
		allFiltered := 0

		for _, d := range state.deals {
			if d.Store == "snapp" {
				snappCount++
			} else if d.Store == "digikala" {
				dkCount++
			}
			if d.DiscountPercent >= minDisc && (storeFilter == "" || storeFilter == "all" || d.Store == storeFilter) {
				allFiltered++
				catCounts[d.CategoryKey]++
				if d.CategoryKey == "ssd" {
					catCounts["hard"]++
				}
			}
		}
		catCounts["all"] = allFiltered

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"scan_count":        state.scanCount,
			"last_scan_time":    state.lastScanTime.Format("15:04:05"),
			"is_scanning":       state.isScanning,
			"total_deals":       len(state.deals),
			"snapp_deals":       snappCount,
			"dk_deals":          dkCount,
			"cat_counts":        catCounts,
			"interval_sec":      state.intervalSec,
			"snapp_requests":    atomic.LoadInt64(&state.snappReqs),
			"snapp_429":         atomic.LoadInt64(&state.snapp429),
			"s26_sniper_active": state.s26SniperActive,
			"s26_scan_count":    state.s26ScanCount,
			"s26_last_scan":     state.s26LastScanTime.Format("15:04:05"),
			"s26_found_count":   len(state.s26FoundDeals),
		})
	})

	http.HandleFunc("/api/sniper/test", func(w http.ResponseWriter, r *http.Request) {
		testDeal := DealItem{
			ID:             "snp_test_s26",
			Store:          "snapp",
			StoreName:      "اسنپ‌شاپ",
			Title:          "گوشی موبایل سامسونگ مدل galaxy S۲۶ دو سیم کارت ظرفیت 256 گیگابایت (تست اسنایپر)",
			PriceToman:     26000000,
			PriceFormatted: "26,000,000",
			ProductURL:     "https://snappshop.ir/product/snp-533616199",
			FastBuyURL:     "https://snappshop.ir/product/snp-533616199?quick_buy=1&sniper=s26",
			ImageURL:       "https://snappshop.ir/assets/images/placeholder.png",
			FoundAt:        time.Now(),
			FoundTimeStr:   time.Now().Format("15:04:05"),
		}
		go state.TriggerS26Sniper(testDeal)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"triggered_s26_test"}`))
	})

	http.HandleFunc("/api/trigger", func(w http.ResponseWriter, r *http.Request) {
		go state.RunScan()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"triggered"}`))
	})

	dashURL := fmt.Sprintf("http://localhost:%d", *portPtr)

	if !*noBrowserPtr && runtime.GOOS == "windows" {
		go func() {
			time.Sleep(1 * time.Second)
			exec.Command("cmd", "/c", "start", dashURL).Start()
		}()
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
