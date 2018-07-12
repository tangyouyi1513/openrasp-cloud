package mongo

import (
	"gopkg.in/mgo.v2"
	. "openrasp-cloud/config"
	log "github.com/sirupsen/logrus"
)
var session *mgo.Session

func init() {
	session, err := mgo.Dial(Conf.MongoAddress)
	if err != nil {
		log.WithError(err).Panic("failed to connect mongodb: %s", Conf.MongoAddress)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	//c := session.DB("test").C("people")
	//err = c.Insert(&Person{"Ale", "111111"}, &Person{"Cla", "222222222"})
	//if err != nil {
	//	panic(err)
	//}
	//result := Person{}
	//err = c.Find(bson.M{"name": "Ale"}).One(&result)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Phone:", result.Phone)
}
