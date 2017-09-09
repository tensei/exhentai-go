package exhentai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
)

type (
	Exhentai struct {
		client    *http.Client
		loggedIn  bool
		Ratelimit time.Duration
	}
	apiResponse struct {
		Gmetadata []struct {
			ArchiverKey  string   `json:"archiver_key"`
			Category     string   `json:"category"`
			Expunged     bool     `json:"expunged"`
			Filecount    string   `json:"filecount"`
			Filesize     int      `json:"filesize"`
			Gid          int      `json:"gid"`
			Posted       string   `json:"posted"`
			Rating       string   `json:"rating"`
			Tags         []string `json:"tags"`
			Thumb        string   `json:"thumb"`
			Title        string   `json:"title"`
			TitleJpn     string   `json:"title_jpn"`
			Token        string   `json:"token"`
			Torrentcount string   `json:"torrentcount"`
			Uploader     string   `json:"uploader"`
		} `json:"gmetadata"`
	}
)

const (
	DefaultRatelimit = time.Second / 2
	CookieDomain     = ".exhentai.org"
	DefaultUrl       = "https://exhentai.org"
	ApiUrl           = "https://api.e-hentai.org/api.php"
)

func NewClient() (*Exhentai, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return &Exhentai{}, fmt.Errorf("%s", err)
	}
	client := &http.Client{
		Timeout: time.Second * 10,
		Jar:     jar,
	}
	return &Exhentai{client, false, DefaultRatelimit}, nil
}

func (ex *Exhentai) Login(memberid, passhash string) error {
	return ex.login(memberid, passhash)
}

func (ex *Exhentai) login(memberid, passhash string) error {
	cookies := []*http.Cookie{
		&http.Cookie{
			Name:   "ipb_member_id",
			Value:  memberid,
			Path:   "/",
			Domain: CookieDomain,
		},
		&http.Cookie{
			Name:   "ipb_pass_hash",
			Value:  passhash,
			Path:   "/",
			Domain: CookieDomain,
		},
	}
	url, err := url.Parse(DefaultUrl)
	if err != nil {
		return err
	}

	ex.client.Jar.SetCookies(url, cookies)

	resp, err := ex.client.Get(DefaultUrl)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if strings.Contains(string(body), "Favorites") {
		ex.loggedIn = true
		fmt.Println("Logged in")
		return nil
	}
	return fmt.Errorf("Couldn't login")
}

func (ex *Exhentai) Download(gallery, savepath string) error {
	return ex.download(gallery, savepath)
}

func (ex *Exhentai) download(gallery, savepath string) error {
	if !ex.loggedIn {
		return fmt.Errorf("Not logged in")
	}

	metadata, err := ex.metadata(gallery)
	if err != nil {
		fmt.Println("Couldn't get metadata")
		return err
	}

	files, err := strconv.Atoi(metadata.Gmetadata[0].Filecount)
	if err != nil {
		return err
	}

	resp, err := ex.client.Get(gallery)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}
	imagelinks := []string{}
	// linkfinder
	lf := func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && strings.Contains(href, fmt.Sprintf("%d-", metadata.Gmetadata[0].Gid)) {
			imagelinks = append(imagelinks, href)
		}
	}
	//use linkfinder func
	doc.Find("a").Each(lf)

	imagepages := []string{}
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && strings.Contains(href, gallery) && strings.Contains(href, "?p=") {
			imagepages = append(imagepages, href)
		}
	})

	imagepages = distinct(imagepages)

	if len(imagepages) > 0 {
		for _, page := range imagepages {
			resp, err := ex.client.Get(page)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromResponse(resp)
			if err != nil {
				return err
			}
			//use linkfinder func
			doc.Find("a").Each(lf)
		}
	}
	imagelinks = distinct(imagelinks)

	fmt.Println(len(imagelinks), files)

	for _, imglink := range imagelinks {
		resp, err := ex.client.Get(imglink)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return err
		}
		original := false
		fsplit := strings.Split(imglink, "-")
		folder := filepath.Join(savepath, fsplit[len(fsplit)-1])
		doc.Find("a").Each(func(index int, item *goquery.Selection) {
			if href, ok := item.Attr("href"); ok && strings.Contains(href, "https://exhentai.org/fullimg.php?gid=") {
				fmt.Println("Found Original", imglink)
				resp, err := ex.client.Get(href)
				if err != nil {
					fmt.Println(err)
					return
				}
				location := resp.Request.URL.Path
				defer resp.Body.Close()

				extension := strings.Split(location, ".")

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					return
				}

				err = ioutil.WriteFile(fmt.Sprintf("%s.%s", folder, extension[len(extension)-1]), body, 0755)
				if err != nil {
					fmt.Println(err)
				}
				original = true
			}
		})
		if !original {
			doc.Find("img").Each(func(index int, item *goquery.Selection) {
				if href, ok := item.Attr("src"); ok && strings.Contains(href, "/keystamp=") {
					fmt.Println("Found", imglink)
					resp, err := ex.client.Get(href)
					if err != nil {
						fmt.Println(err)
						return
					}
					location := resp.Request.URL.Path
					defer resp.Body.Close()

					extension := strings.Split(location, ".")

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println(err)
						return
					}
					err = ioutil.WriteFile(fmt.Sprintf("%s.%s", folder, extension[len(extension)-1]), body, 0755)
					if err != nil {
						fmt.Println(err)
					}
				}
			})
		}
		time.Sleep(ex.Ratelimit)
	}

	return nil
}

func distinct(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

func getGalleryIdToken(url string) (gallery_id, gallery_token string) {
	us := strings.Split(url, "/")
	return us[len(us)-3], us[len(us)-2]
}

func (ex *Exhentai) Metadata(url string) (apiResponse, error) {
	return ex.metadata(url)
}

func (ex *Exhentai) metadata(url string) (apiResponse, error) {
	metadata := apiResponse{}
	if !ex.loggedIn {
		return metadata, fmt.Errorf("Not logged in")
	}

	gallery_id, gallery_token := getGalleryIdToken(url)

	data := fmt.Sprintf("{\"method\": \"gdata\",\"gidlist\": [[%s,\"%s\"]],\"namespace\": 1}", gallery_id, gallery_token)
	apiresp, err := ex.client.Post(ApiUrl, "Application/json", bytes.NewBufferString(data))
	if err != nil {
		return metadata, err
	}

	defer apiresp.Body.Close()

	b, err := ioutil.ReadAll(apiresp.Body)
	if err != nil {
		return metadata, err
	}

	err = json.Unmarshal(b, &metadata)
	return metadata, err
}
