package redis

import (
	uuid "github.com/satori/go.uuid"
)

const (
	redisMemberKey     = "member:"
	chatroomExpireTime = 43200
)

func (conn *RedisConnection) SetChatroom(chatroomID string, token ...string) (string, error) {
	if chatroomID == "" {
		chatroomID = uuid.NewV4().String()
	}

	p := conn.client.Pipeline()
	p.SAdd(redisMemberKey+chatroomID, token)
	p.Expire(redisMemberKey+chatroomID, chatroomExpireTime)
	_, err := p.Exec()

	if err != nil {
		return "", err
	}

	return chatroomID, nil
}

func (conn *RedisConnection) GetMember(chatroomID string) ([]string, error) {
	return conn.client.SMembers(redisMemberKey + chatroomID).Result()
}

func (conn *RedisConnection) RemoveMember(chatroomID string, tokens ...string) error {
	_, err := conn.client.SRem(redisMemberKey+chatroomID, tokens).Result()
	if err != nil {
		return err
	}

	return nil
}

func (conn *RedisConnection) RenewExpireTime(chatroomID string) error {
	_, err := conn.client.Expire(redisMemberKey+chatroomID, chatroomExpireTime).Result()
	if err != nil {
		return err
	}

	return nil
}
