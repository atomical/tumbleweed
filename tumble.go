package main

import (
  "fmt"
  "flag"
  "net/http"
  "errors"
  "os"
  "io/ioutil"
  "encoding/json"
  "path/filepath"
  "github.com/cheggaaa/pb"
)

const (
  NUMBER_OF_IMAGES_PER_REQUEST = 50
  HTTP_USER_AGENT = "tumbleweed 0.0.1"
)

type APIResponse struct {
  PostsTotal int `json:"posts-total"`
  Posts []Post
}

type Post struct {
  // Id uint64 `json:"id"` - sometimes returned  as string, sometimes returned as int. 
  //                         the only way to get around this is to write your own JSON parser.
  PhotoURL1280 string `json:"photo-url-1280"`
}

var accountName string
var userAgent string

func init(){
  var helpFlag bool

  flag.StringVar(&userAgent, "u", HTTP_USER_AGENT, "http user agent")
  flag.BoolVar(&helpFlag, "help", false, "help")

  flag.Parse()

  accountName = flag.Arg(0)

  if helpFlag || len(accountName) == 0 {
    fmt.Println("Usage: tumblweed [accountname]")
    flag.PrintDefaults()
    os.Exit(0)
  }

  // convention to save files to /account name/file.jpg
  err := os.Mkdir(accountName, 0777)

  if err != nil {
    if ! os.IsExist(err){
      panic(err)
    }
  }

}

func main(){

  var posts []Post
  postsTotal := -1
  var bar *pb.ProgressBar

  fmt.Printf("Building index for %v.\n", accountName )

  for currentPos := 0; postsTotal == -1 || currentPos < postsTotal; currentPos += NUMBER_OF_IMAGES_PER_REQUEST {
    
    url := fmt.Sprintf("http://%s.tumblr.com/api/read/json?type=photo&debug=1&start=%v&num=%v", 
            accountName, currentPos, NUMBER_OF_IMAGES_PER_REQUEST) 
    
    data, err := fetchURL(url)
    
    apiResponse := &APIResponse{}

    err = json.Unmarshal(data, apiResponse)

    if err != nil {
      panic(err) 
    }

    if currentPos == 0 {
      postsTotal = apiResponse.PostsTotal
      bar = pb.StartNew(apiResponse.PostsTotal / NUMBER_OF_IMAGES_PER_REQUEST)
      if apiResponse.PostsTotal == 0 {
        fmt.Println("%s has zero posts.", accountName)
        return
      }
    }

    posts = append(posts, apiResponse.Posts...) 
    
    bar.Increment()

  }

  fmt.Println("Downloading images.")

  bar = pb.StartNew(postsTotal)

  for i := range(posts) {
    
    outPath := filepath.Join(accountName,filepath.Base(posts[i].PhotoURL1280))

    if _, err := os.Stat(outPath); ! os.IsNotExist(err) {
      bar.Increment()
      continue
    }

    contents, err := fetchURL(posts[i].PhotoURL1280)
    
    if err != nil {
      fmt.Printf("error downloading photo: %v\n", err)
      continue
    }

    go func( outPath string, contents []byte ){
      err = ioutil.WriteFile(outPath, contents, 0644)
      check(err)
    }( outPath, contents )

    bar.Increment()

  }

}

func fetchURL( url string ) ( []byte, error ) {

  req, err := http.NewRequest("GET", url, nil)
  
  if err != nil {
    return nil, errors.New(fmt.Sprintf("Could not handle request: %v", err))
  }

  req.Header.Set("User-Agent", HTTP_USER_AGENT)

  var httpClient http.Client

  resp, err := httpClient.Do(req)
  
  if err != nil {
    return nil, errors.New(fmt.Sprintf("Could not handle request: %v", err))
  }

  defer resp.Body.Close()
  
  body, err := ioutil.ReadAll(resp.Body)
  
  if resp.StatusCode == http.StatusNotFound {
    return nil, errors.New(fmt.Sprintf("Status Not Found: %v", resp.StatusCode ))
  }

  if err != nil {
    return nil, errors.New(fmt.Sprintf("Unknown error: %v", err ))
  }

  return body, nil 

}
  
func check(e error) {
    if e != nil {
        panic(e)
    }
}