package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
)

var (
	mtx        sync.RWMutex
	members    = flag.String("members", "", "comma seperated list of members")
	checkurl   = flag.String("url", "http://localhost:3000", "")
	adv_addr   = flag.String("adv", "", "")
	items      = map[string]string{}
	broadcasts *memberlist.TransmitLimitedQueue
	currNode = ""
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct{}

type update struct {
	Action string
	Data   map[string]string
}

func init() {
	flag.Parse()
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	var updates []*update
	if err := json.Unmarshal(b[1:], &updates); err != nil {
		return
	}
	mtx.Lock()
	for _, u := range updates {
	    for k, v := range u.Data {
		switch u.Action {
		    case "add":
		 	items[k] = v
		    case "del":
		        delete(items,k)
		}
	    }
	}
	mtx.Unlock()
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	mtx.RLock()
	m := items
	mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	mtx.Lock()
	for k, v := range m {
		items[k] = v
	}
	mtx.Unlock()
}

type eventDelegate struct{}

func (ed *eventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Println("A node has joined: " + node.String())
}

func (ed *eventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Println("A node has left: " + node.String())
}

func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("A node was updated: " + node.String())
}

func addNode(key string, val string) {
	mtx.Lock()
	items[key] = val
	mtx.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: "add",
			Data: map[string]string{
				key: val,
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})
}

func removeNode(key string) {
	mtx.Lock()
	delete(items, key)
	mtx.Unlock()

	b, err := json.Marshal([]*update{{
		Action: "del",
		Data: map[string]string{
			key: "",
		},
	}})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})
}

func start() error {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLocalConfig()
	c.Events = &eventDelegate{}
	c.Delegate = &delegate{}
	c.AdvertiseAddr = *adv_addr
	c.BindPort = 0
	c.Name = hostname + "-" + uuid.NewUUID().String()
	m, err := memberlist.Create(c)
	if err != nil {
		return err
	}
	if len(*members) > 0 {
		parts := strings.Split(*members, ",")
		_, err := m.Join(parts)
		if err != nil {
			return err
		}
	}
	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	node := m.LocalNode()

	currNode = node.Addr.String()

	fmt.Printf("Local member %s:%d\n", node.Addr, node.Port)
	return nil
}

func displayNodeStatus() {
    fmt.Println("total number of active services is ", len(items))
    for k, v := range items {
	fmt.Println(k,v)
    }
}

func checkState(url string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
	    select {
	        case <-ticker.C:
		    resp, err := http.Get(url)
		    if err == nil {
			addNode(currNode, "active")
		    } else {
			removeNode(currNode)
		    }
		    displayNodeStatus()
                    if resp != nil {
		        defer resp.Body.Close()
		    }
	    }
	}
}

func main() {
	if err := start(); err != nil {
		fmt.Println(err)
	}

	checkState(*checkurl)
}
