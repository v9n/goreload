// @TODO
// This is quite dirty. we have so many way to improve it:
//  * Use websocket for live reload instead of JSONP as currently
//  * Generate temp dir, multiple instance
//  * Pure-Go solution to watch file system
//  
package main

import (
  "log"
  "github.com/howeyc/fsnotify"
  "flag"
  "fmt"
  "net/http"
  "os"
  "os/signal"
  "github.com/drone/routes"
  "path/filepath"
  "io"
  "io/ioutil"
  // "os/exec"
  "math/rand"
  "time"
  "strings"
  "crypto/md5"
)

const (
  CHANGE_LOG = "goreload.log.v01"
  DEFAULT_PORT = 51203
)

func Whoami(w http.ResponseWriter, r *http.Request) {
  params := r.URL.Query()
  lastName := params.Get(":last")
  firstName := params.Get(":first")
  fmt.Fprintf(w, "Hey, %s %s. Let include <script> tag to do live reload :-)", firstName, lastName)
}

func BroadcastChange(ev *fsnotify.FileEvent) {
  log.Println("event:", ev)
  contents,_ := ioutil.ReadFile(ev.Name)
  
  h := md5.New()
  io.WriteString(h, string(contents))
  // io.Writ
  hash := h.Sum(nil)
  fmt.Printf("%x", h.Sum(nil))
  log.Print(hash)
  ioutil.WriteFile("/tmp/" + CHANGE_LOG, []byte(fmt.Sprintf("%x", h.Sum(nil))), 0777)      
  //rand.Seed(time.Now().Unix())
  // content := fmt.Sprintf("%v", rand.Int())
}

func main() {
  os.Mkdir("tmp", 0777)
  // Open a channel for signal processing
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt, os.Kill)
  go func() {
  for sig := range c {
    fmt.Println("Signal received:", sig)
    //Clean up 
    fmt.Println("Cleaning up...")
    os.Remove("/tmp/" + CHANGE_LOG)
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
    os.Exit(0)  
  }
    
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal(err)
  }

  done := make(chan bool)
  // Process events
  go func() {
    for {
      select {
        case ev := <-watcher.Event:
          BroadcastChange(ev)
        case err := <-watcher.Error:
          log.Println("error:", err)
        }
    }
  }()

  f := func(d string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if !info.IsDir() || strings.Contains(d, ".git") || strings.Contains(d, ".svn") {
      return nil
    }

    fmt.Println(fmt.Sprintf("Watch: %s", d))
    err = watcher.Watch(d)
    if err != nil {
      log.Println(err)
    }  
    return nil
  }

  filepath.Walk(*argPath, f)

  // <-done

    //Ok, so 
    //fswatch ~/Sites/goreload "goreload -n $RANDOM"
    // Watch the change
    // go func() {
    //   c1 := make(chan bool)
    //   path, _ := os.Getwd()
    //   watchCmd := exec.Command(path + "/fswatch", "~/Sites/goreload ", "\"" + path + "/goreload -n changed\"")
    //   //watchCmd := exec.Command("ls", "~/Sites/goreload ", "\"" + path + "/goreload -n changed\"")
    //   err := watchCmd.Run()
    //   //out, err := watchCmd.Output()
    //   // log.Println(out)
    //   if err != nil {
    //       log.Fatal(err)
    //       return
    //   }  
    //   <- c1
    // }()

  // Give the user some kind of feedback
  fmt.Println(fmt.Sprintf("Starting static file server at %s on port %v", *argPath, *argPort))

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

  mux.Get("/reload" , Reload)
  mux.Get("/reload/" , Reload)
  mux.Get("/reload/:last_change", Reload)

  http.Handle("/", mux)
  fmt.Println(fmt.Sprintf("Include this code in your app:\n <script src=\"http://127.0.0.1:%v/reload\"></script>", *argPort))
  http.ListenAndServe(fmt.Sprintf(":%v", *argPort), nil)    
  <- done
  watcher.Close()
}