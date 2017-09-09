# exhentai-go
## Example
```go
package main

import (
	"encoding/json"
	"fmt"

	exhentai "github.com/tensei/exhentai-go"
)

func main() {
	exhentai, err := exhentai.NewClient()
	if err != nil {
		fmt.Println(err)
	}
  	// ipb_member_id and ipb_pass_hash from exhentai site cookies
	err = exhentai.Login("xxxxx", "xxxxx")
	if err != nil {
		fmt.Println(err)
	}

	// the gallery you want do download
	gallery := "https://exhentai.org/g/{GalleryID}/{GalleryToken}/"
	
	metadata, err := exhentai.Metadata(gallery)
	if err != nil {
		fmt.Println(err)
	}

	ms, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ms)

	err = exhentai.Download(gallery, "/save/path")
	fmt.Println(err)
}
```
