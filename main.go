package main

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"sync"
)

// Mailbox is simple queue implementation
type Mailbox struct {
	buf [][]byte
}

// Send appends message to mailbox
func (m *Mailbox) Send(msg []byte) {
	m.buf = append(m.buf, msg)
}

// Pop returns first message from mailbox, returns nil if empty
func (m *Mailbox) Pop() []byte {
	if len(m.buf) == 0 {
		return nil
	}

	var msg []byte
	msg, m.buf = m.buf[0], m.buf[1:]
	return msg
}

// User tracks mailbox for each subscription
type User struct {
	Name      string
	Mailboxes map[*Topic]*Mailbox
}

// NextMessage returns next message if subscribed, nil on empty
func (user *User) NextMessage(topic *Topic) (msg []byte, subscribed bool) {
	mailbox, ok := user.Mailboxes[topic]
	if !ok {
		return
	}

	subscribed = true
	msg = mailbox.Pop()
	return
}

// Topic tracks all users subscribed to it
type Topic struct {
	Name        string
	Subscribers map[*User]struct{}
}

// Subscribe adds user to topic subscriptions
func (topic *Topic) Subscribe(user *User) {
	_, exists := topic.Subscribers[user]
	if exists {
		return
	}

	topic.Subscribers[user] = struct{}{}
	user.Mailboxes[topic] = &Mailbox{}
}

// Unsubscribe removes user from subscriptions, returns false if user
// is not subscribed
func (topic *Topic) Unsubscribe(user *User) bool {
	_, exists := topic.Subscribers[user]
	if !exists {
		return false
	}

	delete(topic.Subscribers, user)
	delete(user.Mailboxes, topic)

	return true
}

// Publish distributes message to all subscribed mailboxes
func (topic *Topic) Publish(message []byte) {
	for user := range topic.Subscribers {
		user.Mailboxes[topic].Send(message)
	}
}

// SubscriptionService tracks all created topics/users
//
// It provides locking, so that all subscriptions operations
// are serialized. Fine-grained locking is not required, as
// all operations are fast (only memory access).
type SubscriptionService struct {
	sync.Mutex
	topics map[string]*Topic
	users  map[string]*User
}

func (svc *SubscriptionService) getTopic(name string) (topic *Topic) {
	var ok bool

	topic, ok = svc.topics[name]
	if !ok {
		topic = &Topic{Name: name, Subscribers: make(map[*User]struct{})}
		svc.topics[name] = topic
	}

	return
}

func (svc *SubscriptionService) getUser(name string) (user *User) {
	var ok bool

	user, ok = svc.users[name]
	if !ok {
		user = &User{Name: name, Mailboxes: make(map[*Topic]*Mailbox)}
		svc.users[name] = user
	}

	return
}

// POST /:user/:topic
func (svc *SubscriptionService) SubscribeUserHandler(rw http.ResponseWriter, req *http.Request) {
	svc.Lock()
	defer svc.Unlock()

	topicname, username := mux.Vars(req)["topic"], mux.Vars(req)["username"]

	topic := svc.getTopic(topicname)
	user := svc.getUser(username)

	topic.Subscribe(user)
}

// DELETE /:user/:topic
func (svc *SubscriptionService) UnsubscribeUserHandler(rw http.ResponseWriter, req *http.Request) {
	svc.Lock()
	defer svc.Unlock()

	topicname, username := mux.Vars(req)["topic"], mux.Vars(req)["username"]

	topic := svc.getTopic(topicname)
	user := svc.getUser(username)

	if !topic.Unsubscribe(user) {
		rw.WriteHeader(404)
	}
}

// POST /:topic
func (svc *SubscriptionService) PublishHandler(rw http.ResponseWriter, req *http.Request) {
	svc.Lock()
	defer svc.Unlock()

	topicname := mux.Vars(req)["topic"]

	topic := svc.getTopic(topicname)

	message, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	topic.Publish(message)
}

// GET /:topic/:user
func (svc *SubscriptionService) NextMessageHandler(rw http.ResponseWriter, req *http.Request) {
	svc.Lock()
	defer svc.Unlock()

	topicname, username := mux.Vars(req)["topic"], mux.Vars(req)["username"]

	topic := svc.getTopic(topicname)
	user := svc.getUser(username)

	message, ok := user.NextMessage(topic)
	if !ok {
		rw.WriteHeader(404)
		return
	}

	if message == nil {
		rw.WriteHeader(204)
		return
	}

	_, err := rw.Write(message)
	if err != nil {
		panic(err)
	}
}

func main() {
	svc := &SubscriptionService{
		topics: make(map[string]*Topic),
		users:  make(map[string]*User)}

	r := mux.NewRouter()
	r.HandleFunc("/{topic}/{username}", svc.SubscribeUserHandler).Methods("POST")
	r.HandleFunc("/{topic}/{username}", svc.UnsubscribeUserHandler).Methods("DELETE")
	r.HandleFunc("/{topic}", svc.PublishHandler).Methods("POST")
	r.HandleFunc("/{topic}/{username}", svc.NextMessageHandler).Methods("GET")

	http.Handle("/", r)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
