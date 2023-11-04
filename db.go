package main

import "database/sql"

type LemmyDatabase struct {
	db *sql.DB
}

func (l *LemmyDatabase) GetCaptchaAnswer(uuid string) (string, error) {
	answer := ""
	res, err := l.db.Query("select answer from captcha_answer where uuid = $1", uuid)
	if err != nil {
		return answer, err
	}
	defer res.Close()
	if !res.Next() {
		return answer, nil
	}
	err = res.Scan(&answer)
	return answer, err
}
