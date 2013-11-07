// @TODO
// This is quite dirty. we have so many way to improve it:
//  * Use websocket for live reload instead of JSONP as currently
//  * Generate temp dir, multiple instance
//  * Pure-Go solution to watch file system
//  
package main

import (
  "log"
  // "github.com/howeyc/fsnotify"
  "flag"
  "fmt"
  "net/http"
  "os"
  "os/signal"
  "github.com/drone/routes"
  "path/filepath"
  "io/ioutil"
  "os/exec"
  "math/rand"
  "time"
  "bytes"
)

const CHANGE_LOG = "goreload.log.v01"
const DEFAULT_PORT = 51203

func Whoami(w http.ResponseWriter, r *http.Request) {
  params := r.URL.Query()
  lastName := params.Get(":last")
  firstName := params.Get(":first")
  fmt.Fprintf(w, "Hey, %s %s. Let include <script> tag to do live reload :-)", firstName, lastName)
}

func main() {
  os.Mkdir("tmp", 0777)

  // Open a channel for signal processing
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt, os.Kill)
  go func() {
  for sig := range c {
    fmt.Println("Signal received:", sig)
    //Clearn up 
    fmt.Println("Exiting...")
    os.Exit(0)
    }
  }()

  // Get the command-line arguments
  // fswatch ~/Sites/goreload "goreload -n $RANDOM"
  argNotice := flag.String("n", "none", "The port to run goreload on. Defaults to 8080.")
  argPort := flag.Int("p", DEFAULT_PORT, "The port to run goreload on. Defaults to 8080.")
  if *argPort < 1024 {
    log.Fatal("You should use port > 1024 to not require sudo perm.")
  }
  argPath := flag.String("d", "./", "The path you want goreload watch for the change. Any change inside this directory will trigger reload. Multiple directory separate by comma.")
  flag.Parse()

  fmt.Println(*argNotice)
  if "none" != *argNotice {
    // log.Fatal("We got new chance")
    //contents,_ := ioutil.ReadFile("plikTekstowy.txt")
    rand.Seed(time.Now().Unix())
    content := fmt.Sprintf("%v", rand.Int())
    log.Print(content)
    ioutil.WriteFile("/tmp/" + CHANGE_LOG, []byte(content), 0777)    
  } else {
    //Ok, so 
    //fswatch ~/Sites/goreload "goreload -n $RANDOM"
    // Watch the change
    path, _ := os.Getwd()
    watchCmd := exec.Command(path + "/fswatch", "~/Sites/goreload ", "\"" + path + "/goreload -n changed\"")
    // watchCmd.Stdin = strings.NewReader("some input")
    var out bytes.Buffer
    watchCmd.Stdout = &out
    err := watchCmd.Run()
    if err != nil {
      log.Fatal(err)
    }

    f := func(d string, info os.FileInfo, err error) error {
      if err != nil {
        return err
      }
      fmt.Println(d)
      return nil
    }
    filepath.Walk(".", f)

    // Give the user some kind of feedback
    fmt.Println(fmt.Sprintf("Starting static file server at %s on port %v", *argPath, *argPort))

    // Start the server on argPort, using FileServer at argPath as the handler
    // assetPath := "./assets"
    
    // err := http.ListenAndServe(fmt.Sprintf(":%v", *argPort), http.FileServer(http.Dir(assetPath)))
    // if err != nil {
    //   fmt.Println("Error running web server for static assets:", err)
    // }

    mux := routes.New()
    pwd, _ := os.Getwd()
    mux.Static("/asset", pwd)
    mux.Get("/hello/:last/:first", Whoami)

    Reload := func (w http.ResponseWriter, r *http.Request) {
          params := r.URL.Query()
          lastChange := params.Get(":last_change")
          js := `(function () {
          var reloadInterval = 2000
          setTimeout(function () {
            var script = document.createElement('script')
            script.src = 'http://127.0.0.1:%v/reload/%s'
            document.getElementsByTagName('head')[0].appendChild(script)            
          }, reloadInterval)
        })()`
          if lastChange == "" {
            log.Println("First request. Never do reload on this")
          }
          log.Print(fmt.Sprintf("last change on request: %s", lastChange))

          var _changedAt []byte
          changedAt := ""
          _changedAt, err := ioutil.ReadFile("/tmp/" + CHANGE_LOG)
          if err == nil {
            changedAt = string(_changedAt)    
          }
          log.Print(fmt.Sprintf("last change on file: %s", changedAt))

          // log.Print(changedAt)

          if ( string(changedAt) == lastChange || lastChange == "") {
            log.Println("First request or nothing has ever changed. Never do reload on this")
            fmt.Fprintf(w, js, *argPort, changedAt)
          } else {
            //Okay, we got to reload the page
            js = `(function () {window.location.reload(true)})()`
            fmt.Fprintf(w, js)
          }
    }

    mux.Get("/reload", Reload)
    mux.Get("/reload/:last_change", Reload)

    http.Handle("/", mux)
    fmt.Println(fmt.Sprintf("Include this code in your app:\n <script src=\"http://127.0.0.1:%v/reload\"></script>", *argPort))
    http.ListenAndServe(fmt.Sprintf(":%v", *argPort), nil)    
  }

}