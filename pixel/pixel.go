package pixel

import (
    "sort"
    "bytes"
    "fmt"
    "net/http"
    "io/ioutil"

    "appengine"
    "appengine/mail"
    "appengine/datastore"
)

type Environment struct {
    SendFrom string
    SendTo string
}

func init() {
    http.HandleFunc("/", handler)
    //http.HandleFunc("/updateEnv", updateEnvHandler)
}

/*
func updateEnvHandler(w http.ResponseWriter, r *http.Request) {
    context := appengine.NewContext(r)

    key := datastore.NewKey(context, "Environment", "env", 0, nil)
    env := Environment {
        SendFrom: "",
        SendTo: "",
    }
    if _, err := datastore.Put(context, key, &env); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Fprint(w, "env updated")
}
*/

func handler(w http.ResponseWriter, r *http.Request) {
    context := appengine.NewContext(r)
    sendEmail := true

    // Load config data from datastore
    key := datastore.NewKey(context, "Environment", "env", 0, nil)
    var env Environment
    if err := datastore.Get(context, key, &env); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Construct the email message to send.
    sortedKeys := make([]string, len(r.Header))
    i := 0
    for k, _ := range r.Header {
        sortedKeys[i] = k
        i++
    }
    sort.Strings(sortedKeys)

    var buffer bytes.Buffer
    for _, k := range sortedKeys {
        v := r.Header[k]
        buffer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
    }
    buffer.WriteString(fmt.Sprintf("Remote-Addr: %s\n", r.RemoteAddr))

    path := r.URL.Path[1:]
    buffer.WriteString(path)

    if sendEmail {
        msg := &mail.Message{
            Sender: env.SendFrom,
            To: []string{env.SendTo},
            Subject: "See-through Signal",
            Body: buffer.String(),
        }

        if err := mail.Send(context, msg); err != nil {
            context.Errorf("Couldn't send email: %v", err)
        }
    }

    w.Header().Set("Content-Type", "image/gif")
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    w.Header().Set("Pragma", "no-cache")
    w.Header().Set("Expires", "0")

    gif, err := ioutil.ReadFile("pixel.gif")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprint(w, string(gif))
}
