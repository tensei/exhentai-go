# exhentai-go
## Example
```go
package main

import (
	"encoding/json"
	exhentai "exhentai-go"
	"fmt"
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

	metadata, err := exhentai.Metadata("https://exhentai.org/g/{galleryid}/{gallerytoken}/")
	if err != nil {
		fmt.Println(err)
	}

	ms, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ms)

	err = exhentai.Download("https://exhentai.org/g/xxxxx/xxxxx/", "/mnt/e/Development/Go/src/exhentai-go")
	fmt.Println(err)
}
```
