package mongo

import (
	"gopkg.in/mgo.v2"
	. "openrasp-cloud/config"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	MgoDbName = "openrasp"
)

var masterSession *mgo.Session

type SessionItem struct {
	session  *mgo.Session
	database *mgo.Database
}

func init() {
	var err error
	masterSession, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{Conf.MongoAddress},
		Timeout: time.Second * 20,
		//Source:    Conf.MongoAuthDB,
		//Username:  Conf.MongoUser,
		//Password:  Conf.MongoPassword,
		PoolLimit: 4096, // 连接池容量
	})
	if err != nil {
		log.WithError(err).Panic("failed to connect mongodb: %s", Conf.MongoAddress)
	}

	masterSession.SetMode(mgo.Monotonic, true)
}

func Clone() *SessionItem {
	newSession := masterSession.Clone()
	return &SessionItem{session: newSession, database: newSession.DB(MgoDbName)}
}

func (session *SessionItem) Close() {
	session.session.Close()
}

func (session *SessionItem) Insert(collectionName string, docs ...interface{}) error {
	return session.database.C(collectionName).Insert(docs)
}

func (session *SessionItem) Find(collectionName string, docs interface{}) *mgo.Query {
	return session.database.C(collectionName).Find(docs)
}

func (session *SessionItem) Upsert(collectionName string, selector interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	return session.database.C(collectionName).Upsert(selector, update)
}

//func (session *SessionItem) Upsert(collectionName string, selector interface{}, update interface{}) error {
//	return session.database.C(collectionName).EnsureIndex(mgo.Index{
//		Key:        []string{"Id"},
//		Unique:     true,
//		DropDups:   true,
//		Background: true, // See notes.
//		Sparse:     true,})
//}
