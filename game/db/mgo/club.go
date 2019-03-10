package mgo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type ClubDb struct {
	ID           int64    `json:"ID"`
	Name         string   `json:"Name"`
	CreateUserID uint64   `json:"CreateUserID"`
	Notice       string   `json:"Notice"`
	Arg          string   `json:"Arg"`
	Members      []uint64 `json:"Members"`
}

func incClubID() (int64, error) {
	result := bson.M{}
	_, err := mgoSess.DB("").C("clubid").Find(nil).Apply(mgo.Change{
		Upsert:    true,
		ReturnNew: true,
		Update:    bson.M{"$inc": bson.M{"max": int64(1)}},
	}, result)
	if err != nil {
		return 0, err
	}

	r, _ := result["max"].(int64)
	return r, nil
}

func CreateClub(name string, createUserID uint64, notice, arg string) (club *ClubDb, err error) {
	club = &ClubDb{}
	club.ID, err = incClubID()
	if err != nil {
		return
	}

	club.Name = name
	club.CreateUserID = createUserID
	club.Notice = notice
	club.Arg = arg

	err = mgoSess.DB("").C("clubinfo").Insert(club)
	return
}

func UpdateClub(id int64, notice string, arg string) (club *ClubDb, err error) {
	club = &ClubDb{}
	_, err = mgoSess.DB("").C("clubinfo").Find(bson.M{"id": id}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$set": bson.M{"notice": notice, "arg": arg}},
	}, club)
	return
}

func RemoveClub(id int64) error {
	return mgoSess.DB("").C("clubinfo").Remove(bson.M{"id": id})
}

func JoinClub(clubID int64, userID uint64) (club *ClubDb, err error) {
	club = &ClubDb{}
	_, err = mgoSess.DB("").C("clubinfo").Find(bson.M{"id": clubID}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$addToSet": bson.M{"members": userID}},
	}, club)
	return
}

func ExitClub(clubID int64, userID uint64) (club *ClubDb, err error) {
	club = &ClubDb{}
	_, err = mgoSess.DB("").C("clubinfo").Find(bson.M{"id": clubID}).Apply(mgo.Change{
		Upsert:    false,
		ReturnNew: true,
		Update:    bson.M{"$pull": bson.M{"members": userID}},
	}, club)
	return
}

func QueryClubByID(id int64) (club *ClubDb, err error) {
	club = &ClubDb{}
	err = mgoSess.DB("").C("clubinfo").Find(bson.M{"id": id}).One(club)
	return
}

func QueryClubByMember(userID uint64) (list []*ClubDb, err error) {
	list = make([]*ClubDb, 0)
	err = mgoSess.DB("").C("clubinfo").Find(bson.M{"members": bson.M{"$all": []uint64{userID}}}).All(&list)
	return
}
