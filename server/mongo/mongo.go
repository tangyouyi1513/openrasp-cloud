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

}
