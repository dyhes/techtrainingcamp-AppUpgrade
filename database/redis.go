package database

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var rdb *redis.Client
var cur_id int

const EPTIME = 1

func RedisInitClient() {
	//初始化客户端
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

}

func RedisUpdateDownloadStatus(ruleid string, status bool) error {
	RedisInitClient()
	err := rdb.HIncrBy(ctx, ruleid, "hit_count", 1).Err()
	checkErr(err)
	rdb.Expire(ctx, ruleid, EPTIME*time.Minute)
	if status {
		err = rdb.HIncrBy(ctx, ruleid, "download_count", 1).Err()
		checkErr(err)
	}
	return err
}

func RedisQueryRuleByID(ruleid string) (*[]map[string]string, *[]string, error) {
	RedisInitClient()
	val, err := rdb.HGetAll(ctx, ruleid).Result()
	checkErr(err)
	rdb.Expire(ctx, ruleid, EPTIME*time.Minute)
	s := strings.Split(val["device_list"], ",")
	devices := make([]map[string]string, 0)
	devices = append(devices, val)
	return &devices, &s, err

}

func RedisDeleteRule(ruleid string) error {
	RedisInitClient()
	err := rdb.SRem(ctx, "IDList", ruleid).Err()
	checkErr(err)
	err = rdb.Del(ctx, ruleid).Err()
	checkErr(err)
	err = rdb.Del(ctx, ruleid+"s").Err()
	checkErr(err)
	return err
}

//Redis 更新规则，如果没有则创建，有则覆盖
func RedisUpdateRule(ruleid string, r map[string]string, devices []string) error {
	RedisInitClient()
	err := rdb.SAdd(ctx, "IDList", ruleid).Err()
	checkErr(err)
	err = rdb.HMSet(ctx, ruleid, r).Err()
	checkErr(err)
	rdb.Expire(ctx, ruleid, EPTIME*time.Minute)
	//s := strings.Split(r["device_list"], ",")
	rdb.Del(ctx, ruleid+"s")
	err = rdb.SAdd(ctx, ruleid+"s", devices).Err()
	checkErr(err)
	rdb.Expire(ctx, ruleid+"s", EPTIME*time.Minute)
	return err
}

func RedisUpdateRuleWithList(ruleid string, r map[string]string) error {
	RedisInitClient()
	err := rdb.HMSet(ctx, ruleid, r).Err()
	checkErr(err)
	rdb.Expire(ctx, ruleid, EPTIME*time.Minute)
	s := strings.Split(r["device_list"], ",")
	err = rdb.SAdd(ctx, ruleid+"s", s).Err()
	checkErr(err)
	rdb.Expire(ctx, ruleid+"s", EPTIME*time.Minute)
	return err
}

func RedisGetRuleAttr(ruleid string, attrcode string) (string, error) {
	RedisInitClient()
	val, err := rdb.HGet(ctx, ruleid, attrcode).Result()
	rdb.Expire(ctx, ruleid, EPTIME*time.Minute)
	return val, err

}

func RedisCheckWhiteList(ruleid string, userid string) (bool, error) {
	RedisInitClient()
	val, err := rdb.SIsMember(ctx, ruleid+"s", userid).Result()
	rdb.Expire(ctx, ruleid+"s", EPTIME*time.Minute)
	return val, err
}

func GetIDList() (*[]string, error) {
	RedisInitClient()
	val, err := rdb.SMembers(ctx, "IDList").Result()
	checkErr(err)
	return &val, err
}

// func RedisAddRule(r map[string]string, white_list []string) error {
// 	err := rdb.HMSet(ctx, strconv.Itoa(cur_id), r).Err()
// 	if err != nil {
// 		return err
// 	}
// 	err = rdb.SAdd(ctx, strconv.Itoa(cur_id)+"s", white_list).Err()
// 	if err != nil {
// 		return err
// 	}
// 	cur_id++
// 	rdb.Incr(ctx, "cur_id")
// 	return err
// }