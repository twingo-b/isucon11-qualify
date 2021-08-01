package models

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type JsonArray []*Json

func (j *JsonArray) Commit() error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("initialize.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}

type Json struct {
	JiaUserId   string      `json:"jia_user_id"`
	IsuListById IsuListById `json:"isu_list_by_id"`
}
type IsuListById map[string]JsonIsuInfo

type JsonIsuInfo struct {
	Id            int            `json:"id"`
	Name          string         `json:"name"`
	ImageFileHash [md5.Size]byte `json:"image_file_hash"`
	Character     string         `json:"character"`
	Conditions    JsonConditions `json:"conditions"`
	CreatedAt     time.Time      `json:"created_at"`
}

func ToJsonIsuInfo(id int, isu Isu, conditions JsonConditions) JsonIsuInfo {
	return JsonIsuInfo{
		id,
		isu.Name,
		md5.Sum(isu.Image),
		isu.Character,
		conditions,
		isu.CreatedAt,
	}
}

type JsonConditions struct {
	Info     []JsonCondition `json:"info"`
	Warning  []JsonCondition `json:"warning"`
	Critical []JsonCondition `json:"critical"`
}

type JsonCondition struct {
	Timestamp      int64          `json:"timestamp"`
	IsSitting      bool           `json:"is_sitting"`
	IsDirty        bool           `json:"is_dirty"`
	IsOverweight   bool           `json:"is_overweight"`
	IsBroken       bool           `json:"is_broken"`
	Message        string         `json:"message"`
	CreatedAt      time.Time      `json:"created_at"`
	OwnerIsuUUID   string         `json:"owner_isu_uuid"`
	OwnerIsuID     int            `json:"owner_isu_id"`
	ConditionLevel ConditionLevel `json:"condition_level"`
}

func (j *JsonConditions) AddCondition(condition Condition, isuId int) error {
	jsonCondition := JsonCondition{
		// JST分マイナスすると何故かちょうどよい
		// DB上はJSTなのでUnixtimeはJST時間に変換されている（UTC時間+9表記）
		// JST時間なのにどこかでTimezone情報なくなって時刻表記だけになる→UTCとして解釈される→JST環境では更に+9時間されて本来より9時間多い値になるとか？
		condition.Timestamp.Add(-9 * time.Hour).Unix(),
		condition.IsSitting,
		condition.IsDirty,
		condition.IsOverweight,
		condition.IsBroken,
		condition.Message,
		condition.CreatedAt,
		condition.Isu.JIAIsuUUID,
		isuId,
		condition.ConditionLevel(),
	}
	switch condition.ConditionLevel() {
	case ConditionLevelInfo:
		j.Info = append(j.Info, jsonCondition)
	case ConditionLevelWarning:
		j.Warning = append(j.Warning, jsonCondition)
	case ConditionLevelCritical:
		j.Critical = append(j.Critical, jsonCondition)
	default:
		return fmt.Errorf("想定外のConditionLevelです。")
	}
	return nil
}
