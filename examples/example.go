package main

import (
	"encoding/json"
	"log"

	exhentai "github.com/tensei/exhentai-go"
)

func main() {
	exhentai, err := exhentai.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	err = exhentai.Login("xxxxx", "xxxxx")
	if err != nil {
		log.Fatal(err)
	}

	gallery := "https://exhentai.org/g/{GalleryID}/{GalleryToken}/"
	metadata, err := exhentai.Metadata(gallery)
	if err != nil {
		log.Println(err)
	}

	ms, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		log.Println(err)
	}
	log.Println(ms)

	err = exhentai.Download(gallery, "/save/path")
	log.Println(err)
}
